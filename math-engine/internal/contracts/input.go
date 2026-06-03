package contracts

// CoreInput is the wire format for math-engine (see schema/core_input_v1.json).
type CoreInput struct {
	Version             string              `json:"version" validate:"eq=core_input_v1"`
	RawTrainingLog      RawTrainingLog      `json:"rawTrainingLog" validate:"required"`
	UserMetrics         UserMetrics         `json:"userMetrics" validate:"required"`
	TrainingConstraints TrainingConstraints `json:"trainingConstraints" validate:"required"`
	ProgressionPolicy   ProgressionPolicy   `json:"progressionPolicy" validate:"required"`
	DeloadPolicy        DeloadPolicy        `json:"deloadPolicy" validate:"required"`
	TrainingWeekIndex   int                 `json:"trainingWeekIndex" validate:"required,gte=1"`
	SessionModifiers    SessionModifiers      `json:"sessionModifiers" validate:"required"`
}

// SessionModifiers are per-session fatigue/readiness multipliers (caller-supplied).
type SessionModifiers struct {
	FatigueModifier   float64 `json:"fatigueModifier" validate:"required,gt=0,lte=1"`
	ReadinessModifier float64 `json:"readinessModifier" validate:"required,gt=0,lte=1"`
}

type RawTrainingLog struct {
	SessionID          string     `json:"sessionId" validate:"required"`
	UserID             string     `json:"userId" validate:"required"`
	StartedAt          string     `json:"startedAt" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	CompletedAt        string     `json:"completedAt" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	SessionDurationSec int        `json:"sessionDurationSec" validate:"gt=0"`
	Exercises          []Exercise `json:"exercises" validate:"required,min=1,dive"`
}

type Exercise struct {
	ExerciseID       string            `json:"exerciseId" validate:"required"`
	Name             string            `json:"name" validate:"required"`
	MuscleGroup      string            `json:"muscleGroup" validate:"required"`
	BlockType        string            `json:"blockType" validate:"required,oneof=warmup working"`
	Sets             []SetEntry        `json:"sets" validate:"required,min=1,dive"`
	IntervalProtocol *IntervalProtocol `json:"intervalProtocol,omitempty" validate:"omitempty"`
}

type SetEntry struct {
	SetNumber int     `json:"setNumber" validate:"gte=1"`
	WeightKg  float64 `json:"weightKg" validate:"gte=0"`
	Reps      int     `json:"reps" validate:"gte=0"`
}

type IntervalProtocol struct {
	ProtocolID  string `json:"protocolId" validate:"required"`
	WorkSec     int    `json:"workSec" validate:"gt=0"`
	RestSec     int    `json:"restSec" validate:"gt=0"`
	DisplayName string `json:"displayName" validate:"required"`
}

type UserMetrics struct {
	Age                 int     `json:"age" validate:"gte=1,lte=120"`
	BodyWeightKg        float64 `json:"bodyWeightKg" validate:"gt=0"`
	RecoverySensitivity float64 `json:"recoverySensitivity" validate:"gte=0,lte=1"`
	FatigueThreshold   float64 `json:"fatigueThreshold" validate:"gte=0,lte=1"`
	ReadinessThreshold float64 `json:"readinessThreshold" validate:"gte=0,lte=1"`
}

type TrainingConstraints struct {
	SessionsPerWeek int `json:"sessionsPerWeek" validate:"gte=1,lte=7"`
}

type ProgressionPolicy struct {
	// Strategy is a user-selected preset id (stored in profile/settings).
	Strategy       string  `json:"strategy" validate:"required,oneof=percent_e1rm rir_target"`
	TargetPercent  float64 `json:"targetPercent" validate:"gt=0,lte=1.1"`
	IncreaseStepKg float64 `json:"increaseStepKg" validate:"gt=0"`
	DecreaseStepKg float64 `json:"decreaseStepKg" validate:"gt=0"`
}

type DeloadPolicy struct {
	// Strategy is a user-selected deload preset id (stored in profile/settings).
	Strategy         string  `json:"strategy" validate:"required,oneof=fixed_3_plus_1 fixed_2_plus_1 custom"`
	LoadWeeks        int     `json:"loadWeeks" validate:"gte=1,lte=6"`
	DeloadWeeks      int     `json:"deloadWeeks" validate:"gte=1,lte=2"`
	IntensityDropPct float64 `json:"intensityDropPct" validate:"gte=0.05,lte=0.25"`
	VolumeDropPct    float64 `json:"volumeDropPct" validate:"gte=0.2,lte=0.7"`
	MinRIR           int     `json:"minRIR" validate:"gte=0,lte=6"`
}
