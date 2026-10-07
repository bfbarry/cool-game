// Package protocol defines the wire contract between the Go server and the
// TypeScript client. This file is canonical — see CONTEXT.md §3 for the
// human-readable summary. If you change a shape here, mirror the change in
// frontend/src/net/protocol.ts in the same change, and update CONTEXT.md.
package protocol

import "encoding/json"

// Envelope wraps every message sent over the websocket, outbound direction:
// Data is any protocol *Data struct above, marshaled as-is.
type Envelope struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// RawEnvelope is the inbound shape: Data is left as raw JSON until the
// handler knows Type and can unmarshal into the right *Data struct. See
// server/client.go for where this is decoded.
type RawEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ---- Client -> Server payloads ----

type JoinData struct {
	Name           string `json:"name"`
	ColorPrimary   string `json:"colorPrimary"`
	ColorSecondary string `json:"colorSecondary"`
	HatID          string `json:"hatId"`
}

type InputData struct {
	Seq      uint32  `json:"seq"`
	DX       float64 `json:"dx"` // movement intent, -1..1
	DY       float64 `json:"dy"`
	AimAngle float64 `json:"aimAngle"` // radians
	Firing   bool    `json:"firing"`
	Tick     uint64  `json:"tick"`
}

type EnterStoreData struct{}
type ExitStoreData struct{}

type PurchaseData struct {
	ItemID string `json:"itemId"`
}

type EnterVehicleData struct {
	VehicleID string `json:"vehicleId"`
}

type ExitVehicleData struct {
	VehicleID string `json:"vehicleId"`
}

// ---- Server -> Client payloads ----

type Cosmetics struct {
	ColorPrimary   string `json:"colorPrimary"`
	ColorSecondary string `json:"colorSecondary"`
	HatID          string `json:"hatId"`
}

type PlayerState struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	VX        float64   `json:"vx"`
	VY        float64   `json:"vy"`
	AimAngle  float64   `json:"aimAngle"`
	Health    int       `json:"health"`
	MaxHealth int       `json:"maxHealth"`
	Points    int       `json:"points"`
	WeaponID  string    `json:"weaponId"`
	Cosmetics Cosmetics `json:"cosmetics"`
	State     string    `json:"state"`     // "on_foot" | "in_vehicle"
	VehicleID string    `json:"vehicleId"` // set when State == in_vehicle
}

type NPCState struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Health  int     `json:"health"`
	AIState string  `json:"aiState"`
}

type VehicleState struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Angle   float64 `json:"angle"`
	Health  int     `json:"health"`
	PilotID string  `json:"pilotId"`
}

type ProjectileState struct {
	ID    string  `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
	Speed float64 `json:"speed"`
	Owner string  `json:"owner"`
}

type MapDef struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	// TODO(world owner): platforms, parkour checkpoints, store zones, spawn points.
}

type StoreCatalog struct {
	Items []StoreItem `json:"items"`
}

type StoreItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"` // "weapon" | "vehicle" | "cosmetic"
	Cost int    `json:"cost"`
}

type WelcomeData struct {
	PlayerID     string       `json:"playerId"`
	Map          MapDef       `json:"map"`
	StoreCatalog StoreCatalog `json:"storeCatalog"`
}

type StateData struct {
	Tick        uint64            `json:"tick"`
	Players     []PlayerState     `json:"players"`
	NPCs        []NPCState        `json:"npcs"`
	Vehicles    []VehicleState    `json:"vehicles"`
	Projectiles []ProjectileState `json:"projectiles"`
}

type PlayerJoinedData struct {
	PlayerID string `json:"playerId"`
}

type PlayerLeftData struct {
	PlayerID string `json:"playerId"`
}

type DamageData struct {
	TargetID  string `json:"targetId"`
	SourceID  string `json:"sourceId"`
	Amount    int    `json:"amount"`
	NewHealth int    `json:"newHealth"`
}

type DeathData struct {
	VictimID      string `json:"victimId"`
	KillerID      string `json:"killerId"`
	PointsAwarded int    `json:"pointsAwarded"`
	Cause         string `json:"cause"` // "player" | "npc_kill" | "fall" | ...
}

type PointsUpdateData struct {
	PlayerID string `json:"playerId"`
	Points   int    `json:"points"`
}

type PurchaseResultData struct {
	Success bool   `json:"success"`
	ItemID  string `json:"itemId"`
	Points  int    `json:"points"`
	Reason  string `json:"reason,omitempty"`
}

// Message type string constants, shared by server/ and game/ when building Envelopes.
const (
	TypeJoin         = "join"
	TypeInput        = "input"
	TypeEnterStore   = "enter_store"
	TypeExitStore    = "exit_store"
	TypePurchase     = "purchase"
	TypeEnterVehicle = "enter_vehicle"
	TypeExitVehicle  = "exit_vehicle"

	TypeWelcome        = "welcome"
	TypeState          = "state"
	TypePlayerJoined   = "player_joined"
	TypePlayerLeft     = "player_left"
	TypeDamage         = "damage"
	TypeDeath          = "death"
	TypePointsUpdate   = "points_update"
	TypePurchaseResult = "purchase_result"
)
