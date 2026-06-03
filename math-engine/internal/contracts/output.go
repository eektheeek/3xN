package contracts

// CoreOutput is the wire format for math-engine analysis (see schema/core_output_v1.json).
type CoreOutput struct {
	Version          string            `json:"version"`
	SessionID        string            `json:"sessionId"`
	SessionDate      string            `json:"sessionDate"`
	UserID           string            `json:"userId"`
	ComputedMetrics  ComputedMetrics   `json:"computedMetrics"`
	ExerciseMetrics []ExerciseMetric `json:"exerciseMetrics"`
	SessionContext  SessionContext   `json:"sessionContext"`
	TrendPoints     []TrendSeries    `json:"trendPoints,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
	CreatedAt       string           `json:"createdAt"`
}

// SessionContext holds modifiers shared across per-exercise recommendations.
type SessionContext struct {
	TargetPercent     float64 `json:"targetPercent"`
	FatigueModifier   float64 `json:"fatigueModifier"`
	ReadinessModifier float64 `json:"readinessModifier"`
	DeloadActive      bool    `json:"deloadActive"`
	TrainingWeekIndex int     `json:"trainingWeekIndex"`
}

type ComputedMetrics struct {
	VolumeLoad *MetricValue `json:"volumeLoad,omitempty"`
}

type MetricValue struct {
	Value          float64 `json:"value"`
	Unit           string  `json:"unit"`
	Confidence     float64 `json:"confidence"`
	Applicability  string  `json:"applicability"`
}

type E1RMMetric struct {
	MetricValue
	FormulaFamily    string             `json:"formulaFamily"`
	FormulaBreakdown map[string]float64 `json:"formulaBreakdown"`
}

type ExerciseMetric struct {
	ExerciseID     string                     `json:"exerciseId"`
	ExerciseName   string                     `json:"exerciseName"`
	BlockType      string                     `json:"blockType"`
	E1RM           *E1RMMetric                `json:"e1rm,omitempty"`
	VolumeLoad     *MetricValue               `json:"volumeLoad,omitempty"`
	Recommendation *ExerciseRecommendationOut `json:"recommendation,omitempty"`
}

// ExerciseRecommendationOut is the next-session load suggestion for one working block.
type ExerciseRecommendationOut struct {
	Action         string   `json:"action"`
	TargetWeightKg float64  `json:"targetWeightKg"`
	MinWeightKg    float64  `json:"minWeightKg"`
	MaxWeightKg    float64  `json:"maxWeightKg"`
	ReasonCodes    []string `json:"reasonCodes"`
}

type TrendSeries struct {
	Metric string       `json:"metric"`
	Window string       `json:"window"`
	Points []TrendPoint `json:"points"`
}

type TrendPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}
