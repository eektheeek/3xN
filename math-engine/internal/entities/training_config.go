package entities

// TrainingConstraints limits how often the user trains per week.
type TrainingConstraints struct {
	SessionsPerWeek int
}

// ProgressionPolicy is a user-configurable preset for load progression.
type ProgressionPolicy struct {
	Strategy       string // preset id, e.g. "percent_e1rm", "rir_target"
	TargetPercent  float64
	IncreaseStepKg float64
	DecreaseStepKg float64
}

// DeloadPolicy is a user-configurable preset for recovery microcycles.
type DeloadPolicy struct {
	Strategy         string  // preset id, e.g. "fixed_3_plus_1", "fixed_2_plus_1", "custom"
	LoadWeeks        int     // hard weeks before deload
	DeloadWeeks      int     // deload weeks in cycle
	IntensityDropPct float64 // multiply working weight, e.g. 0.10 = -10%
	VolumeDropPct    float64 // reduce sets/volume, e.g. 0.40 = -40%
	MinRIR           int     // minimum reps in reserve on deload week
}
