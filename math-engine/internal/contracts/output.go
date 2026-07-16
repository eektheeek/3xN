package contracts

// CoreOutput is the wire format for math-engine analysis (see schema/core_output_v1.json).
type CoreOutput struct {
	Version         string            `json:"version"`
	SessionID       string            `json:"sessionId"`
	SessionDate     string            `json:"sessionDate"`
	UserID          string            `json:"userId"`
	ComputedMetrics ComputedMetrics   `json:"computedMetrics"`
	ExerciseMetrics []ExerciseMetric  `json:"exerciseMetrics"`
	SessionContext  SessionContext    `json:"sessionContext"`
	Warnings        []string          `json:"warnings,omitempty"`
	CreatedAt       string            `json:"createdAt"`
}

// SessionContext holds session-wide modifiers for smart trainer.
type SessionContext struct {
	FatigueModifier   float64 `json:"fatigueModifier"`
	ReadinessModifier float64 `json:"readinessModifier"`
	DeloadActive      bool    `json:"deloadActive"`
	TrainingWeekIndex int     `json:"trainingWeekIndex"`
}

type ComputedMetrics struct {
	VolumeLoad *MetricValue `json:"volumeLoad,omitempty"`
}

type MetricValue struct {
	Value         float64 `json:"value"`
	Unit          string  `json:"unit"`
	Confidence    float64 `json:"confidence"`
	Applicability string  `json:"applicability"`
}

type ExerciseMetric struct {
	ExerciseID    string          `json:"exerciseId"`
	ExerciseName  string          `json:"exerciseName"`
	ExerciseType  string          `json:"exerciseType"`
	VolumeLoad    *MetricValue    `json:"volumeLoad,omitempty"`
	SmartTrainer  *SmartTrainerOut `json:"smartTrainer,omitempty"`
}

// SmartTrainerOut is the next-session load suggestion for one working exercise.
type SmartTrainerOut struct {
	LastWorkingWeightKg   float64  `json:"lastWorkingWeightKg"`
	SuggestedNextWeightKg float64  `json:"suggestedNextWeightKg"`
	MinWeightKg           float64  `json:"minWeightKg"`
	MaxWeightKg           float64  `json:"maxWeightKg"`
	Action                string   `json:"action"`
	ReasonCodes           []string `json:"reasonCodes"`
	Message               string   `json:"message"`
}
