package game

import (
	"time"

	"github.com/nir-game/server/internal/combat"
	"github.com/nir-game/server/internal/economy"
	"github.com/nir-game/server/internal/npc"
	"github.com/nir-game/server/internal/player"
	"github.com/nir-game/server/internal/protocol"
)

const TickRate = 20 // Hz, see CONTEXT.md §7

// Run starts the fixed-tick authoritative loop. Blocks — call in its own
// goroutine from cmd/server/main.go.
func (w *World) Run() {
	interval := time.Second / TickRate
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	last := time.Now()
	for range ticker.C {
		now := time.Now()
		dt := now.Sub(last).Seconds()
		last = now

		w.Tick(dt)
		w.sender.Broadcast(protocol.Envelope{Type: protocol.TypeState, Data: w.Snapshot()})
	}
}

// Tick advances all simulation by dt seconds: movement, projectiles/combat
// resolution, NPC AI, vehicles, and parkour checkpoint checks. This is the
// single place everything gets stitched together each tick.
func (w *World) Tick(dt float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.tick++

	for _, p := range w.players {
		if !p.IsAlive() {
			continue
		}
		p.Integrate(dt)
		w.checkParkour(p)
	}

	// TODO(npc owner): spawn NPCs from w.mapDef.NPCSpawns up to some cap,
	// and give them real behavior in npc.Update (currently a no-op stub).
	for _, n := range w.npcs {
		n.Update(dt, 0, 0, 0)
	}

	w.advanceProjectiles()

	// TODO(vehicles owner): advance vehicle physics from pilot input, which
	// currently has no transport (vehicles don't yet receive `input`
	// messages while mounted — see handleInput in world.go).
}

// advanceProjectiles moves projectiles, resolves hits against players and
// NPCs, awards points via economy/, and removes expired/consumed shots.
func (w *World) advanceProjectiles() {
	for id, proj := range w.projectiles {
		alive := proj.Advance(1.0 / TickRate)

		hit := w.resolveProjectileHit(proj)
		if hit || !alive {
			delete(w.projectiles, id)
		}
	}
}

func (w *World) resolveProjectileHit(proj *combat.Projectile) bool {
	for targetID, p := range w.players {
		if targetID == proj.OwnerID || !p.IsAlive() {
			continue
		}
		evt, ok := combat.ResolveHit(proj, targetID, p)
		if !ok {
			continue
		}
		w.sender.Broadcast(protocol.Envelope{Type: protocol.TypeDamage, Data: protocol.DamageData{
			TargetID: evt.TargetID, SourceID: evt.SourceID, Amount: evt.Amount, NewHealth: evt.NewHealth,
		}})
		if evt.Died {
			w.onPlayerDeath(p, proj.OwnerID, "player")
		}
		return true
	}
	for targetID, n := range w.npcs {
		evt, ok := combat.ResolveHit(proj, targetID, n)
		if !ok {
			continue
		}
		w.sender.Broadcast(protocol.Envelope{Type: protocol.TypeDamage, Data: protocol.DamageData{
			TargetID: evt.TargetID, SourceID: evt.SourceID, Amount: evt.Amount, NewHealth: evt.NewHealth,
		}})
		if evt.Died {
			w.onNPCDeath(n, proj.OwnerID)
		}
		return true
	}
	return false
}

func (w *World) onPlayerDeath(victim *player.Player, killerID string, cause string) {
	awarded := 0
	if killer, ok := w.players[killerID]; ok && killer != victim {
		economy.AwardPlayerKill(killer)
		awarded = economy.PointsPerPlayerKill
		w.sender.Broadcast(protocol.Envelope{Type: protocol.TypePointsUpdate, Data: protocol.PointsUpdateData{PlayerID: killerID, Points: killer.Points}})
	}
	w.sender.Broadcast(protocol.Envelope{Type: protocol.TypeDeath, Data: protocol.DeathData{
		VictimID: victim.ID, KillerID: killerID, PointsAwarded: awarded, Cause: cause,
	}})
	spawn := w.mapDef.PlayerSpawn(w.nextSpawn)
	w.nextSpawn++
	victim.Respawn(spawn.X, spawn.Y)
}

func (w *World) onNPCDeath(n *npc.NPC, killerID string) {
	if killer, ok := w.players[killerID]; ok {
		economy.AwardNPCKill(killer)
		w.sender.Broadcast(protocol.Envelope{Type: protocol.TypePointsUpdate, Data: protocol.PointsUpdateData{PlayerID: killerID, Points: killer.Points}})
	}
	w.sender.Broadcast(protocol.Envelope{Type: protocol.TypeDeath, Data: protocol.DeathData{
		VictimID: n.ID, KillerID: killerID, PointsAwarded: economy.PointsPerNPCKill, Cause: "npc_kill",
	}})
	delete(w.npcs, n.ID)
	// TODO(npc owner): respawn NPCs after a delay instead of just removing.
}

func (w *World) checkParkour(p *player.Player) {
	cp, advanced, complete := w.parkour.TryAdvance(p.ID, w.mapDef.Checkpoints, p.X, p.Y)
	if !advanced {
		return
	}
	_ = cp
	economy.AwardParkourCheckpoint(p)
	if complete {
		economy.AwardParkourCourseComplete(p)
	}
	w.sender.SendTo(p.ID, protocol.Envelope{Type: protocol.TypePointsUpdate, Data: protocol.PointsUpdateData{PlayerID: p.ID, Points: p.Points}})
}

// Snapshot builds the full authoritative state sent to clients each tick.
//
// TODO(server/game owner): this is a full snapshot every tick, fine for a
// small player count; consider delta compression if bandwidth becomes an
// issue (see CONTEXT.md networking notes).
func (w *World) Snapshot() protocol.StateData {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Initialized non-nil (rather than left as zero-value slices) so
	// json.Marshal emits "[]" instead of "null" when empty — clients
	// should be able to range over these unconditionally.
	data := protocol.StateData{
		Tick:        w.tick,
		Players:     []protocol.PlayerState{},
		NPCs:        []protocol.NPCState{},
		Vehicles:    []protocol.VehicleState{},
		Projectiles: []protocol.ProjectileState{},
	}
	for _, p := range w.players {
		data.Players = append(data.Players, p.ToWire())
	}
	for _, n := range w.npcs {
		data.NPCs = append(data.NPCs, n.ToWire())
	}
	for _, v := range w.vehicles {
		data.Vehicles = append(data.Vehicles, v.ToWire())
	}
	for _, proj := range w.projectiles {
		data.Projectiles = append(data.Projectiles, proj.ToWire())
	}
	return data
}
