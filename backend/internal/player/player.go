// Package player owns player state: position/movement, health bookkeeping
// hooks (actual damage math lives in combat/), and cosmetic customization.
//
// TODO(owner): flesh out movement physics (acceleration, collision with
// world geometry from world/), and respawn behavior.
package player

import "github.com/nir-game/server/internal/protocol"

const (
	DefaultMaxHealth = 100
	MoveSpeed        = 220.0 // units/sec, tune freely
)

// Player is the authoritative server-side representation of a connected
// player. Field names intentionally mirror protocol.PlayerState so
// converting to a wire snapshot is mechanical (see ToWire).
type Player struct {
	ID        string
	Name      string
	X, Y      float64
	VX, VY    float64
	AimAngle  float64
	Health    int
	MaxHealth int
	Points    int
	WeaponID  string
	Cosmetics protocol.Cosmetics
	State     string // "on_foot" | "in_vehicle"
	VehicleID string
}

// New creates a freshly-spawned player from a join request's cosmetics.
func New(id string, join protocol.JoinData, spawnX, spawnY float64) *Player {
	return &Player{
		ID:        id,
		Name:      join.Name,
		X:         spawnX,
		Y:         spawnY,
		Health:    DefaultMaxHealth,
		MaxHealth: DefaultMaxHealth,
		WeaponID:  "pistol", // TODO(combat owner): pull starter weapon from catalog
		Cosmetics: protocol.Cosmetics{
			ColorPrimary:   join.ColorPrimary,
			ColorSecondary: join.ColorSecondary,
			HatID:          join.HatID,
		},
		State: "on_foot",
	}
}

// ApplyInput advances the player's intended velocity/aim from client input.
// Actual integration (position += velocity * dt) happens in game/loop.go so
// it can be done uniformly with collision against world geometry.
//
// TODO(owner): clamp dx/dy to a unit vector, apply MoveSpeed, handle
// in-vehicle state (movement should be delegated to vehicles/ instead).
func (p *Player) ApplyInput(in protocol.InputData) {
	p.AimAngle = in.AimAngle
	p.VX = in.DX * MoveSpeed
	p.VY = in.DY * MoveSpeed
}

// Integrate moves the player by its current velocity over dt seconds.
//
// TODO(world owner): collide against map geometry / parkour platforms
// instead of moving freely.
func (p *Player) Integrate(dt float64) {
	if p.State == "in_vehicle" {
		return // vehicle module drives position in this state
	}
	p.X += p.VX * dt
	p.Y += p.VY * dt
}

// ToWire converts authoritative state into the snapshot shape sent to clients.
func (p *Player) ToWire() protocol.PlayerState {
	return protocol.PlayerState{
		ID:        p.ID,
		Name:      p.Name,
		X:         p.X,
		Y:         p.Y,
		VX:        p.VX,
		VY:        p.VY,
		AimAngle:  p.AimAngle,
		Health:    p.Health,
		MaxHealth: p.MaxHealth,
		Points:    p.Points,
		WeaponID:  p.WeaponID,
		Cosmetics: p.Cosmetics,
		State:     p.State,
		VehicleID: p.VehicleID,
	}
}

// IsAlive reports whether the player can currently take actions / be hit.
func (p *Player) IsAlive() bool {
	return p.Health > 0
}

// --- combat.Hittable implementation ---
// Lets combat.ResolveHit operate on a *Player without combat/ importing
// player/ (avoids an import cycle, since player/ stays dependency-free).

const HitRadius = 24.0 // units; tune against sprite size

func (p *Player) Position() (x, y float64) { return p.X, p.Y }

func (p *Player) Radius() float64 { return HitRadius }

// TakeDamage applies damage and reports the resulting health + whether this
// hit killed the player. Does not award points or reset state on death —
// that's game/'s job, coordinating with economy/.
func (p *Player) TakeDamage(amount int) (newHealth int, died bool) {
	p.Health -= amount
	if p.Health < 0 {
		p.Health = 0
	}
	return p.Health, p.Health == 0
}

// --- economy.Earner implementation ---

// AddPoints credits the player. Satisfies economy.Earner without player/
// importing economy/ (same one-directional-dependency pattern as Hittable).
func (p *Player) AddPoints(amount int) {
	p.Points += amount
}

// --- economy.Spender implementation ---

func (p *Player) PointsBalance() int { return p.Points }

func (p *Player) DeductPoints(amount int) bool {
	if p.Points < amount {
		return false
	}
	p.Points -= amount
	return true
}

func (p *Player) GrantWeapon(itemID string) {
	p.WeaponID = itemID
}

// GrantVehicleAccess marks that the player has purchased a vehicle.
//
// TODO(vehicles owner): decide how ownership works — does buying a
// helicopter spawn one immediately, or unlock it for later spawn/mount at a
// vehicle pad? Wire that decision in here and in vehicles/.
func (p *Player) GrantVehicleAccess(itemID string) {
	// Intentionally left as a TODO hook; see comment above.
}

// Respawn resets health/position after a death.
//
// TODO(world owner): pick a real spawn point instead of the origin.
func (p *Player) Respawn(spawnX, spawnY float64) {
	p.Health = p.MaxHealth
	p.X, p.Y = spawnX, spawnY
	p.VX, p.VY = 0, 0
	p.State = "on_foot"
	p.VehicleID = ""
}
