package formulas

import "math"

// Estimate1RMLombardi returns estimated one-rep max using Fred Lombardi's formula:
// e1RM = W * R^0.10.
// Useful across a wider rep range; returns 0 if weight <= 0 or reps < 1.
func Estimate1RMLombardi(weightKg float64, reps int) float64 {
	if weightKg <= 0 || reps < 1 {
		return 0
	}
	return weightKg * math.Pow(float64(reps), 0.10)
}
