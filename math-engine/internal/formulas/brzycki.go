package formulas

// Estimate1RMBrzycki returns estimated one-rep max using Matt Brzycki's formula:
// e1RM = W * 36 / (37 - R).
// Best used for roughly 1–10 reps; returns 0 if weight <= 0, reps < 1, or reps >= 37.
func Estimate1RMBrzycki(weightKg float64, reps int) float64 {
	if weightKg <= 0 || reps < 1 || reps >= 37 {
		return 0
	}
	return weightKg * 36 / float64(37-reps)
}
