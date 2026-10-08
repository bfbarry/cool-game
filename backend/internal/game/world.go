// Package game orchestrates everything else: it holds the authoritative
// World (all players/npcs/vehicles/projectiles), implements
// server.GameHandler to receive client messages, and runs the fixed-tick
// loop (loop.go) that advances state and broadcasts snapshots.
//
// This file wires player/, combat/, economy/, vehicles/, npc/, and world/
// together. If you're adding a new gameplay system, this is usually where
// it gets plugged in alongside the existing ones.
package game

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nir-game/server/internal/combat"
	"github.com/nir-game/server/internal/economy"
	"github.com/nir-game/server/internal/npc"
	"github.com/nir-game/server/internal/player"
	"github.com/nir-game/server/internal/protocol"
	"github.com/nir-game/server/internal/server"
	"github.com/nir-game/server/internal/vehicles"
	gworld "github.com/nir-game/server/internal/world"
)

type World struct {
	mu sync.Mutex

	sender server.Sender
	mapDef *gworld.Map

	players     map[string]*player.Player
	npcs        map[string]*npc.NPC
	vehicles    map[string]*vehicles.Vehicle
	projectiles map[string]*combat.Projectile

	inStore  map[string]bool // playerID -> currently inside a store zone (client-declared, server-verified on purchase)
	lastFire map[string]time.Time
	parkour  *gworld.ProgressTracker

	tick      uint64
	nextSpawn int
	idCounter uint64
}

func NewWorld(sender server.Sender) *World {
	return &World{
		sender:      sender,
		mapDef:      gworld.Default(),
		players:     make(map[string]*player.Player),
		npcs:        make(map[string]*npc.NPC),
		vehicles:    make(map[string]*vehicles.Vehicle),
		projectiles: make(map[string]*combat.Projectile),
		inStore:     make(map[string]bool),
		lastFire:    make(map[string]time.Time),
		parkour:     gworld.NewProgressTracker(),
	}
}

func (w *World) nextID(prefix string) string {
	w.idCounter++
	return fmt.Sprintf("%s-%d", prefix, w.idCounter)
}

// --- server.GameHandler implementation ---

func (w *World) HandleConnect(playerID string) {
	log.Printf("game: %s connected (awaiting join)", playerID)
}

func (w *World) HandleMessage(playerID string, raw protocol.RawEnvelope) {
	switch raw.Type {
	case protocol.TypeJoin:
		var data protocol.JoinData
		if decode(raw.Data, &data) {
			w.handleJoin(playerID, data)
		}
	case protocol.TypeInput:
		var data protocol.InputData
		if decode(raw.Data, &data) {
			w.handleInput(playerID, data)
		}
	case protocol.TypeEnterStore:
		w.mu.Lock()
		w.inStore[playerID] = true
		w.mu.Unlock()
	case protocol.TypeExitStore:
		w.mu.Lock()
		w.inStore[playerID] = false
		w.mu.Unlock()
	case protocol.TypePurchase:
		var data protocol.PurchaseData
		if decode(raw.Data, &data) {
			w.handlePurchase(playerID, data)
		}
	case protocol.TypeEnterVehicle:
		var data protocol.EnterVehicleData
		if decode(raw.Data, &data) {
			w.handleEnterVehicle(playerID, data)
		}
	case protocol.TypeExitVehicle:
		var data protocol.ExitVehicleData
		if decode(raw.Data, &data) {
			w.handleExitVehicle(playerID, data)
		}
	default:
		log.Printf("game: unknown message type %q from %s", raw.Type, playerID)
	}
}

func (w *World) HandleDisconnect(playerID string) {
	w.mu.Lock()
	delete(w.players, playerID)
	delete(w.inStore, playerID)
	delete(w.lastFire, playerID)
	w.mu.Unlock()
	w.sender.Broadcast(protocol.Envelope{Type: protocol.TypePlayerLeft, Data: protocol.PlayerLeftData{PlayerID: playerID}})
}

// --- message handlers ---

func (w *World) handleJoin(playerID string, data protocol.JoinData) {
	player.ValidateJoin(&data)

	w.mu.Lock()
	spawn := w.mapDef.PlayerSpawn(w.nextSpawn)
	w.nextSpawn++
	p := player.New(playerID, data, spawn.X, spawn.Y)
	w.players[playerID] = p
	w.mu.Unlock()

	w.sender.SendTo(playerID, protocol.Envelope{
		Type: protocol.TypeWelcome,
		Data: protocol.WelcomeData{
			PlayerID: playerID,
			Map: protocol.MapDef{
				Width:  w.mapDef.Width,
				Height: w.mapDef.Height,
			},
			StoreCatalog: protocol.StoreCatalog{Items: economy.Catalog},
		},
	})
	w.sender.Broadcast(protocol.Envelope{Type: protocol.TypePlayerJoined, Data: protocol.PlayerJoinedData{PlayerID: playerID}})
}

func (w *World) handleInput(playerID string, data protocol.InputData) {
	w.mu.Lock()
	p, ok := w.players[playerID]
	if !ok || !p.IsAlive() {
		w.mu.Unlock()
		return
	}
	p.ApplyInput(data)
	shouldFire := data.Firing && w.canFireLocked(playerID, p)
	w.mu.Unlock()

	if shouldFire {
		w.fireWeapon(p)
	}
}

// canFireLocked checks weapon fire-rate cooldown. Caller must already hold
// w.mu — this does not lock it itself (previously it did, which deadlocked
// against the caller's own lock since sync.Mutex isn't reentrant; every
// left-click would freeze the whole server).
func (w *World) canFireLocked(playerID string, p *player.Player) bool {
	if p == nil {
		return false
	}
	weapon, ok := combat.WeaponByID(p.WeaponID)
	if !ok {
		return false
	}
	last, seen := w.lastFire[playerID]
	if seen && time.Since(last) < time.Duration(weapon.FireRateMs)*time.Millisecond {
		return false
	}
	w.lastFire[playerID] = time.Now()
	return true
}

func (w *World) fireWeapon(p *player.Player) {
	weapon, ok := combat.WeaponByID(p.WeaponID)
	if !ok {
		return
	}
	w.mu.Lock()
	proj := combat.Fire(w.nextID("proj"), p.ID, p.X, p.Y, p.AimAngle, weapon)
	w.projectiles[proj.ID] = proj
	w.mu.Unlock()
}

func (w *World) handlePurchase(playerID string, data protocol.PurchaseData) {
	w.mu.Lock()
	p, ok := w.players[playerID]
	inStore := w.inStore[playerID]
	w.mu.Unlock()
	if !ok {
		return
	}
	if !inStore || !w.mapDef.InStoreZone(p.X, p.Y) {
		w.sender.SendTo(playerID, protocol.Envelope{
			Type: protocol.TypePurchaseResult,
			Data: protocol.PurchaseResultData{Success: false, ItemID: data.ItemID, Points: p.Points, Reason: "not in store"},
		})
		return
	}
	result := economy.Purchase(p, data.ItemID)
	w.sender.SendTo(playerID, protocol.Envelope{Type: protocol.TypePurchaseResult, Data: result})
	if result.Success {
		w.sender.Broadcast(protocol.Envelope{Type: protocol.TypePointsUpdate, Data: protocol.PointsUpdateData{PlayerID: playerID, Points: result.Points}})
	}
}

func (w *World) handleEnterVehicle(playerID string, data protocol.EnterVehicleData) {
	w.mu.Lock()
	defer w.mu.Unlock()
	v, ok := w.vehicles[data.VehicleID]
	p, pok := w.players[playerID]
	if !ok || !pok {
		return
	}
	if v.Mount(playerID) {
		p.State = "in_vehicle"
		p.VehicleID = v.ID
	}
}

func (w *World) handleExitVehicle(playerID string, data protocol.ExitVehicleData) {
	w.mu.Lock()
	defer w.mu.Unlock()
	v, ok := w.vehicles[data.VehicleID]
	p, pok := w.players[playerID]
	if !ok || !pok {
		return
	}
	if v.PilotID == playerID {
		v.Dismount()
		p.State = "on_foot"
		p.VehicleID = ""
	}
}

// decode is a tiny helper around json.Unmarshal with logging, used when
// dispatching RawEnvelope payloads above.
func decode(raw []byte, out any) bool {
	if err := json.Unmarshal(raw, out); err != nil {
		log.Printf("game: decode error: %v", err)
		return false
	}
	return true
}
