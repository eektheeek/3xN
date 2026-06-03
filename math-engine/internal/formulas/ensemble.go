package formulas

import "sort"

// Estimate1RMEnsemble returns the median of Epley, Brzycki, and Lombardi estimates.
// Invalid formula results (0) are excluded. Returns 0 if none are valid.
func Estimate1RMEnsemble(weightKg float64, reps int) float64 {
	candidates := []float64{
		Estimate1RMEpley(weightKg, reps),
		Estimate1RMBrzycki(weightKg, reps),
		Estimate1RMLombardi(weightKg, reps),
	}

	var valid []float64
	for _, v := range candidates {
		if v > 0 {
			valid = append(valid, v)
		}
	}

	return median(valid)
}

func median(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	sort.Float64s(values)
	mid := n / 2
	if n%2 == 1 {
		return values[mid]
	}
	return (values[mid-1] + values[mid]) / 2
}
