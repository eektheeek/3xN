package entities

import "time"

// Exercise block types for stack ordering (warmup blocks skip load recommendations).
const (
	BlockTypeWarmup  = "warmup"
	BlockTypeWorking = "working"
)

// WorkoutSession is a normalized training session used by the math engine.
type WorkoutSession struct {
	SessionID           string
	UserID              string
	StartedAt           time.Time
	CompletedAt         time.Time
	SessionDurationSec  int
	Exercises           []Exercise
}

// Exercise describes a single movement inside one session.
type Exercise struct {
	ExerciseID       string
	Name             string
	MuscleGroup      string
	BlockType        string // warmup | working
	Sets             []SetEntry
	IntervalProtocol *IntervalProtocol
}

// SetEntry is one performed set with load and repetitions.
type SetEntry struct {
	SetNumber int
	WeightKg  float64
	Reps      int
}

// IntervalProtocol defines planned work/rest intervals for an exercise block.
// Number of work intervals is derived from len(Sets), not from protocol fields.
type IntervalProtocol struct {
	ProtocolID string
	WorkSec   int
	RestSec   int
	DisplayName string
}
