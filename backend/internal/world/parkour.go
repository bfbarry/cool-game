package world

// Checkpoint is one node in a parkour course. Reaching it in order awards
// economy.PointsPerParkourCheckpoint; completing the whole course awards
// economy.PointsParkourCourseBonus. Point-awarding itself happens in game/
// (which owns per-player progress) — this package only defines geometry.
type Checkpoint struct {
	Index  int
	X, Y   float64
	Radius float64
}

// DefaultParkourCourse is a placeholder chain of checkpoints.
//
// TODO(owner): design real parkour geometry (platforms to jump between)
// alongside these checkpoints in Default()'s Platforms list.
func DefaultParkourCourse() []Checkpoint {
	return []Checkpoint{
		{Index: 0, X: 500, Y: 2200, Radius: 60},
		{Index: 1, X: 900, Y: 2000, Radius: 60},
		{Index: 2, X: 1300, Y: 1800, Radius: 60},
		{Index: 3, X: 1700, Y: 1600, Radius: 60},
	}
}

// Reached reports whether (x, y) is within range of the given checkpoint.
func (c Checkpoint) Reached(x, y float64) bool {
	return Distance(x, y, c.X, c.Y) <= c.Radius
}

// ProgressTracker records, per player, the next checkpoint index expected.
// Lives in game/ normally (it's per-session state, not map geometry) — this
// type is provided here as a convenience starting point.
//
// TODO(owner): move/adapt this into game/world.go's per-player bookkeeping,
// wire it to economy.AwardParkourCheckpoint / AwardParkourCourseComplete.
type ProgressTracker struct {
	NextIndex map[string]int // playerID -> next expected checkpoint index
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{NextIndex: make(map[string]int)}
}

// TryAdvance checks if the player has reached their next expected
// checkpoint; if so, advances their progress and returns true.
func (t *ProgressTracker) TryAdvance(playerID string, course []Checkpoint, x, y float64) (checkpoint Checkpoint, advanced bool, courseComplete bool) {
	next := t.NextIndex[playerID]
	if next >= len(course) {
		return Checkpoint{}, false, false
	}
	cp := course[next]
	if !cp.Reached(x, y) {
		return Checkpoint{}, false, false
	}
	t.NextIndex[playerID] = next + 1
	complete := t.NextIndex[playerID] >= len(course)
	return cp, true, complete
}
