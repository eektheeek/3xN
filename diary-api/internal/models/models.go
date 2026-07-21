package models

// Exercise is a catalog movement the user can log.
type Exercise struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	MuscleGroup    string            `json:"muscleGroup"`
	SupportsAssist bool              `json:"supportsAssist"`
	ProtocolID     string            `json:"protocolId,omitempty"`
	CreatedAt      string            `json:"createdAt"`
	Target         *ExerciseTarget   `json:"target,omitempty"`
	Protocol       *IntervalProtocol `json:"protocol,omitempty"`
}

// IntervalProtocol is a saved Tabata / interval template.
type IntervalProtocol struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PrepareSec  int    `json:"prepareSec"` // countdown before first work
	WorkSec     int    `json:"workSec"`
	RestSec     int    `json:"restSec"`
	WarmupExtra bool   `json:"warmupExtra"` // rounds = sets + 1 when true
	CreatedAt   string `json:"createdAt"`
}

// ExerciseTarget is the current goal for an exercise.
type ExerciseTarget struct {
	ExerciseID string  `json:"exerciseId"`
	TargetSets int     `json:"sets"`
	TargetReps int     `json:"reps"`
	WeightKg   float64 `json:"weightKg"`
	AssistKg   float64 `json:"assistKg"` // 0 when unused; only meaningful if exercise.supportsAssist
}

// WorkoutSession is one gym visit / training day.
type WorkoutSession struct {
	ID            string                   `json:"id"`
	PerformedAt   string                   `json:"performedAt"`
	StartedAt     string                   `json:"startedAt,omitempty"`
	DurationSec   int                      `json:"durationSec"`
	IsDeload      bool                     `json:"isDeload"`
	WorkoutPlanID string                   `json:"workoutPlanId,omitempty"`
	CycleID       string                   `json:"cycleId,omitempty"`
	CycleName     string                   `json:"cycleName,omitempty"`
	CycleStep     int                      `json:"cycleStep,omitempty"`
	CreatedAt     string                   `json:"createdAt"`
	Exercises     []WorkoutSessionExercise `json:"exercises,omitempty"`
}

// Cycle is a named sequential track of session templates (e.g. "Дом", "Улица").
type Cycle struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	CurrentStep int         `json:"currentStep"` // 1-based; currentStep > len(steps) means completed
	Completed   bool        `json:"completed"`   // true when all steps finished
	OnHome      bool        `json:"onHome"`      // shown on the home screen
	CreatedAt   string      `json:"createdAt"`
	Steps       []CycleStep `json:"steps"`
}

// CycleStep is one slot in a training cycle.
type CycleStep struct {
	ID                  string `json:"id"`
	CycleID             string `json:"cycleId"`
	Position            int    `json:"position"`
	WorkoutPlanID       string `json:"workoutPlanId"`
	WorkoutPlanName     string `json:"workoutPlanName,omitempty"`
	CompletedSessionID  string `json:"completedSessionId,omitempty"`
	CompletedPerformedAt string `json:"completedPerformedAt,omitempty"`
}

// WorkoutSessionExercise is one exercise block inside a workout session.
type WorkoutSessionExercise struct {
	ID               string `json:"id"`
	WorkoutSessionID string `json:"workoutSessionId"`
	ExerciseID       string `json:"exerciseId"`
	Position         int    `json:"position"`
	ExerciseName     string `json:"exerciseName,omitempty"`
	SupportsAssist   bool   `json:"supportsAssist,omitempty"`
	Sets             []Set  `json:"sets,omitempty"`
}

// WorkoutSessionSummary is a lightweight row for diary list/calendar views.
type WorkoutSessionSummary struct {
	ID            string   `json:"id"`
	PerformedAt   string   `json:"performedAt"`
	DurationSec   int      `json:"durationSec"`
	IsDeload      bool     `json:"isDeload"`
	CreatedAt     string   `json:"createdAt"`
	ExerciseCount int      `json:"exerciseCount"`
	ExerciseNames []string `json:"exerciseNames"`
	CycleID       string   `json:"cycleId,omitempty"`
	CycleName     string   `json:"cycleName,omitempty"`
	CycleStep     int      `json:"cycleStep,omitempty"`
}

// Set is one performed set.
type Set struct {
	ID                       string  `json:"id"`
	WorkoutSessionExerciseID string  `json:"workoutSessionExerciseId"`
	SetNumber                int     `json:"setNumber"`
	Reps                     int     `json:"reps"`
	WeightKg                 float64 `json:"weightKg"`
	AssistKg                 float64 `json:"assistKg"` // 0 when unused
}

// WorkoutPlan is a saved workout template (stack of exercises).
type WorkoutPlan struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	CreatedAt  string                `json:"createdAt"`
	Exercises  []WorkoutPlanExercise `json:"exercises,omitempty"`
}

// WorkoutPlanExercise is one exercise slot in a plan.
type WorkoutPlanExercise struct {
	ID            string `json:"id"`
	WorkoutPlanID string `json:"workoutPlanId"`
	ExerciseID    string `json:"exerciseId"`
	Position      int    `json:"position"`
	ExerciseName  string `json:"exerciseName,omitempty"`
}
