// Package vehicles implements rideable vehicles (helicopters, etc.) — the
// big-ticket items in the Points economy. A vehicle is a separate entity
// from any player; a player "mounts" one to control it.
//
// TODO(owner): flesh out flight physics for the helicopter, mount/dismount
// rules (proximity check, one pilot at a time), and vehicle combat (does it
// have its own health/weapon? can it be shot down?).
package vehicles

import "github.com/nir-game/server/internal/protocol"

type Type struct {
	ID        string
	Name      string
	Cost      int
	MaxHealth int
	MaxSpeed  float64
}

// Catalog mirrors the vehicle entries in economy.Catalog — keep costs in
// sync (see economy/store.go TODO about generating one from the other).
var Catalog = map[string]Type{
	"helicopter": {ID: "helicopter", Name: "Helicopter", Cost: 5000, MaxHealth: 300, MaxSpeed: 400},
}

// Vehicle is a server-side vehicle instance in the world.
type Vehicle struct {
	ID      string
	Type    string
	X, Y    float64
	Angle   float64
	Health  int
	PilotID string // empty if unmanned
}

func Spawn(id, typeID string, x, y float64) (*Vehicle, bool) {
	t, ok := Catalog[typeID]
	if !ok {
		return nil, false
	}
	return &Vehicle{ID: id, Type: typeID, X: x, Y: y, Health: t.MaxHealth}, true
}

// Mount assigns a pilot if the vehicle is currently unmanned.
func (v *Vehicle) Mount(playerID string) bool {
	if v.PilotID != "" {
		return false
	}
	v.PilotID = playerID
	return true
}

func (v *Vehicle) Dismount() {
	v.PilotID = ""
}

// Update advances vehicle physics by dt seconds based on pilot input.
//
// TODO(owner): implement actual helicopter flight model (lift, drag,
// rotation) instead of this placeholder.
func (v *Vehicle) Update(dt float64, dx, dy, angle float64) {
	t := Catalog[v.Type]
	v.Angle = angle
	v.X += dx * t.MaxSpeed * dt
	v.Y += dy * t.MaxSpeed * dt
}

func (v *Vehicle) ToWire() protocol.VehicleState {
	return protocol.VehicleState{
		ID:      v.ID,
		Type:    v.Type,
		X:       v.X,
		Y:       v.Y,
		Angle:   v.Angle,
		Health:  v.Health,
		PilotID: v.PilotID,
	}
}
