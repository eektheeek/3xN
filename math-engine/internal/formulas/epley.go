package formulas

// Estimate1RMEpley returns estimated one-rep max using Boyd Epley's formula:
// e1RM = W * (1 + R/30). e1RM is the estimated one-rep max in kg. Where W is the weight lifted in kg and R is the number of reps.
// Best used for roughly 1–12 reps; returns 0 if weight <= 0 or reps < 1.
func Estimate1RMEpley(weightKg float64, reps int) float64 {
	if weightKg <= 0 || reps < 1 {
		return 0
	}
	return weightKg * (1 + float64(reps)/30)
}
