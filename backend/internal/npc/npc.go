// Package npc implements hostile, non-player enemies that wander the map
// and award Points when killed (see economy.AwardNPCKill).
//
// TODO(owner): real AI — patrol/aggro/attack state machine, pathing around
// world geometry, actually damaging nearby players.
package npc

import "github.com/nir-game/server/internal/protocol"

const (
	DefaultMaxHealth = 40
	DefaultDamage    = 5
	HitRadius        = 20.0
)

type NPC struct {
	ID      string
	Type    string
	X, Y    float64
	Health  int
	Damage  int
	AIState string // "idle" | "patrol" | "aggro" | "attack"
}

func New(id, npcType string, x, y float64) *NPC {
	return &NPC{
		ID:      id,
		Type:    npcType,
		X:       x,
		Y:       y,
		Health:  DefaultMaxHealth,
		Damage:  DefaultDamage,
		AIState: "idle",
	}
}

// Update runs one tick of NPC AI.
//
// TODO(owner): implement. Stub currently does nothing (NPC stands still).
func (n *NPC) Update(dt float64, nearestPlayerX, nearestPlayerY float64, nearestPlayerDist float64) {
	_ = dt
	_ = nearestPlayerX
	_ = nearestPlayerY
	_ = nearestPlayerDist
}

// --- combat.Hittable implementation ---

func (n *NPC) Position() (x, y float64) { return n.X, n.Y }

func (n *NPC) Radius() float64 { return HitRadius }

func (n *NPC) TakeDamage(amount int) (newHealth int, died bool) {
	n.Health -= amount
	if n.Health < 0 {
		n.Health = 0
	}
	return n.Health, n.Health == 0
}

func (n *NPC) ToWire() protocol.NPCState {
	return protocol.NPCState{
		ID:      n.ID,
		Type:    n.Type,
		X:       n.X,
		Y:       n.Y,
		Health:  n.Health,
		AIState: n.AIState,
	}
}
