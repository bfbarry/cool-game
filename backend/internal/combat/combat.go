// Package combat handles weapons, projectiles, hit detection, and damage
// application. Points for kills are awarded by economy/ — combat should
// only report damage/death events and let the caller (game/) hook in
// economy.AwardKillPoints etc., keeping modules decoupled.
//
// TODO(owner): replace the straight-line/no-collision projectile stub with
// real hit detection against players, NPCs, and world geometry.
package combat

import (
	"math"

	"github.com/nir-game/server/internal/protocol"
)

// Projectile is a server-simulated bullet in flight.
type Projectile struct {
	ID       string
	OwnerID  string
	X, Y     float64
	Angle    float64
	Speed    float64
	Damage   float64
	traveled float64
	maxRange float64
}

// Fire spawns a projectile for the given shooter using their equipped
// weapon's stats. Call sites should rate-limit via Weapon.FireRateMs.
func Fire(id, ownerID string, x, y, angle float64, w Weapon) *Projectile {
	return &Projectile{
		ID:       id,
		OwnerID:  ownerID,
		X:        x,
		Y:        y,
		Angle:    angle,
		Speed:    w.BulletSpeed,
		Damage:   float64(w.Damage),
		maxRange: w.Range,
	}
}

// Advance moves the projectile by dt seconds. Returns false once it has
// traveled past its weapon's range and should be removed.
func (p *Projectile) Advance(dt float64) bool {
	dist := p.Speed * dt
	p.X += math.Cos(p.Angle) * dist
	p.Y += math.Sin(p.Angle) * dist
	p.traveled += dist
	return p.traveled <= p.maxRange
}

func (p *Projectile) ToWire() protocol.ProjectileState {
	return protocol.ProjectileState{
		ID:    p.ID,
		X:     p.X,
		Y:     p.Y,
		Angle: p.Angle,
		Speed: p.Speed,
		Owner: p.OwnerID,
	}
}

// Hittable is anything combat can damage: players and NPCs both satisfy
// this via small adapters in their own packages, avoiding an import of
// player/npc here (prevents an import cycle with game/).
type Hittable interface {
	Position() (x, y float64)
	Radius() float64
	TakeDamage(amount int) (newHealth int, died bool)
}

// DamageEvent describes one resolved hit, for the caller to turn into
// protocol.DamageData / protocol.DeathData and award economy points.
type DamageEvent struct {
	TargetID  string
	SourceID  string
	Amount    int
	NewHealth int
	Died      bool
}

// ResolveHit checks whether a projectile currently overlaps a hittable
// target and, if so, applies damage. Returns ok=false if no hit.
func ResolveHit(p *Projectile, targetID string, target Hittable) (DamageEvent, bool) {
	tx, ty := target.Position()
	dx, dy := p.X-tx, p.Y-ty
	distSq := dx*dx + dy*dy
	r := target.Radius()
	if distSq > r*r {
		return DamageEvent{}, false
	}
	newHealth, died := target.TakeDamage(int(p.Damage))
	return DamageEvent{
		TargetID:  targetID,
		SourceID:  p.OwnerID,
		Amount:    int(p.Damage),
		NewHealth: newHealth,
		Died:      died,
	}, true
}
