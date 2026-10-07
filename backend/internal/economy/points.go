// Package economy owns the Points ledger and the in-world Store: how
// points are earned (kills, NPC kills, parkour) and spent (weapons,
// vehicles, cosmetics). Keep this the single source of truth for prices —
// see CONTEXT.md §5.
package economy

// Point award constants. Tune freely, but keep CONTEXT.md §5 in sync.
const (
	PointsPerPlayerKill        = 100
	PointsPerNPCKill           = 25
	PointsPerParkourCheckpoint = 15
	PointsParkourCourseBonus   = 75
)

// Earner is anything that accumulates points — satisfied by *player.Player
// without economy/ importing player/ (mirrors the Hittable pattern in
// combat/), keeping module dependencies one-directional.
type Earner interface {
	AddPoints(amount int)
}

func AwardPlayerKill(killer Earner) {
	if killer == nil {
		return
	}
	killer.AddPoints(PointsPerPlayerKill)
}

func AwardNPCKill(killer Earner) {
	if killer == nil {
		return
	}
	killer.AddPoints(PointsPerNPCKill)
}

func AwardParkourCheckpoint(p Earner) {
	p.AddPoints(PointsPerParkourCheckpoint)
}

func AwardParkourCourseComplete(p Earner) {
	p.AddPoints(PointsParkourCourseBonus)
}
