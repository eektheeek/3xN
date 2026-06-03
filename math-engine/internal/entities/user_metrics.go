package entities

// UserMetrics holds athlete profile data used for load and recovery adjustments.
type UserMetrics struct {
	Age                 int
	BodyWeightKg        float64
	RecoverySensitivity float64
	// FatigueThreshold: if fatigue modifier is below this, recommend decrease (0 = use product default).
	FatigueThreshold float64
	// ReadinessThreshold: if readiness modifier is below this, recommend decrease (0 = use product default).
	ReadinessThreshold float64
}
