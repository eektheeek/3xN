package entities

// SessionModifiers holds per-session load adjustments (fatigue, readiness).
type SessionModifiers struct {
	FatigueModifier   float64
	ReadinessModifier float64
}
