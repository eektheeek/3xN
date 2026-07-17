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
	ID          string                   `json:"id"`
	PerformedAt string                   `json:"performedAt"`
	StartedAt   string                   `json:"startedAt,omitempty"`
	DurationSec int                      `json:"durationSec"`
	IsDeload    bool                     `json:"isDeload"`
	CreatedAt   string                   `json:"createdAt"`
	Exercises   []WorkoutSessionExercise `json:"exercises,omitempty"`
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
