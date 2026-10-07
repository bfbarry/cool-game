// Package world owns the map: static geometry (platforms/walls), spawn
// points, store zones, and parkour course definitions. Other modules query
// this package rather than hardcoding positions.
//
// TODO(owner): replace the placeholder single-room map with a real level —
// platforms for parkour, a store building, NPC spawn areas.
package world

import "math"

type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// Platform is solid parkour/collision geometry.
type Platform struct {
	Rect
}

// StoreZone is an area where `purchase` requests are allowed.
type StoreZone struct {
	Rect
}

type SpawnPoint struct {
	X, Y float64
}

type Map struct {
	Width, Height float64
	Platforms     []Platform
	StoreZones    []StoreZone
	PlayerSpawns  []SpawnPoint
	NPCSpawns     []SpawnPoint
	Checkpoints   []Checkpoint
}

// Default returns the placeholder starter map: an open field, one store
// zone, a handful of spawns, and a short parkour checkpoint chain.
//
// TODO(owner): design a real map. Keep this function as the single place
// that assembles it so protocol.MapDef (sent to clients on `welcome`) and
// server-side collision/zone checks never drift apart.
func Default() *Map {
	return &Map{
		Width:  4000,
		Height: 2400,
		Platforms: []Platform{
			{Rect{X: 0, Y: 2350, W: 4000, H: 50}}, // ground
		},
		StoreZones: []StoreZone{
			{Rect{X: 1800, Y: 2100, W: 400, H: 250}},
		},
		PlayerSpawns: []SpawnPoint{
			{X: 200, Y: 2200}, {X: 3800, Y: 2200}, {X: 2000, Y: 200},
		},
		NPCSpawns: []SpawnPoint{
			{X: 1000, Y: 2200}, {X: 3000, Y: 2200},
		},
		Checkpoints: DefaultParkourCourse(),
	}
}

// RandomPlayerSpawn picks a spawn point. Deterministic round-robin for now.
func (m *Map) PlayerSpawn(index int) SpawnPoint {
	if len(m.PlayerSpawns) == 0 {
		return SpawnPoint{}
	}
	return m.PlayerSpawns[index%len(m.PlayerSpawns)]
}

// InStoreZone reports whether (x,y) is inside any store zone — used to gate
// `purchase` messages server-side (never trust the client's own claim).
func (m *Map) InStoreZone(x, y float64) bool {
	for _, z := range m.StoreZones {
		if z.Contains(x, y) {
			return true
		}
	}
	return false
}

// DistanceTo is a small shared helper; combat/npc packages may want their
// own, but it's trivial enough to duplicate rather than force a dependency.
func Distance(x1, y1, x2, y2 float64) float64 {
	dx, dy := x2-x1, y2-y1
	return math.Hypot(dx, dy)
}
