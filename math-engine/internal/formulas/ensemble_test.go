package formulas

import "testing"

func TestEstimate1RMEnsemble_golden_50x8(t *testing.T) {
	const (
		weightKg = 50.0
		reps     = 8
		want     = 62.1 // median(Epley, Brzycki, Lombardi)
	)

	got := Estimate1RMEnsemble(weightKg, reps)

	const epsilon = 0.05
	if diff := got - want; diff < -epsilon || diff > epsilon {
		t.Fatalf("Estimate1RMEnsemble(%v, %d) = %.2f, want %.1f (±%.2f)", weightKg, reps, got, want, epsilon)
	}
}
