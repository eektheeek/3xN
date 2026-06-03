package recommendation

import (
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

func defaultPolicy() entities.ProgressionPolicy {
	return entities.ProgressionPolicy{
		Strategy:       "percent_e1rm",
		TargetPercent:  0.78,
		IncreaseStepKg: 2.5,
		DecreaseStepKg: 2.5,
	}
}

func defaultUserMetrics() entities.UserMetrics {
	return entities.UserMetrics{
		Age:                 40,
		BodyWeightKg:        82,
		RecoverySensitivity: 0.7,
		FatigueThreshold:   0.96,
		ReadinessThreshold: 0.75,
	}
}

func TestRecommendNext_increase(t *testing.T) {
	// e1RM ~62.1, 78% ~48.4, last 45 -> room to progress
	res := RecommendNext(Input{
		E1RM:                62.1,
		LastWorkingWeightKg: 45,
		ProgressionPolicy:   defaultPolicy(),
		FatigueModifier:     1,
		ReadinessModifier:   1,
		UserMetrics:         defaultUserMetrics(),
	})
	if res.Action != ActionIncrease {
		t.Fatalf("action = %q, want increase (target=%.1f)", res.Action, res.TargetWeightKg)
	}
}

func TestRecommendNext_decrease_recovery(t *testing.T) {
	res := RecommendNext(Input{
		E1RM:                62.1,
		LastWorkingWeightKg: 50,
		ProgressionPolicy:   defaultPolicy(),
		FatigueModifier:     0.94,
		ReadinessModifier:   1,
		UserMetrics:         defaultUserMetrics(),
	})
	if res.Action != ActionDecrease {
		t.Fatalf("action = %q, want decrease", res.Action)
	}
}

func TestRecommendNext_customThresholds(t *testing.T) {
	metrics := defaultUserMetrics()
	metrics.FatigueThreshold = 0.98 // stricter than default 0.96

	res := RecommendNext(Input{
		E1RM:                62.1,
		LastWorkingWeightKg: 50,
		ProgressionPolicy:   defaultPolicy(),
		FatigueModifier:     0.97,
		ReadinessModifier:   1,
		UserMetrics:         metrics,
	})
	if res.Action != ActionDecrease {
		t.Fatalf("action = %q, want decrease with fatigueThreshold 0.98 and fatigue 0.97", res.Action)
	}
}

func TestModifierThresholds_defaultsWhenZero(t *testing.T) {
	f, r := modifierThresholds(entities.UserMetrics{})
	if f != defaultFatigueThreshold || r != defaultReadinessThreshold {
		t.Fatalf("defaults = %.2f/%.2f, want %.2f/%.2f", f, r, defaultFatigueThreshold, defaultReadinessThreshold)
	}
}

func TestRecommendNext_deload(t *testing.T) {
	p := entities.DeloadPolicy{
		Strategy:         "fixed_3_plus_1",
		IntensityDropPct: 0.10,
	}
	if !IsDeloadWeek(p, 4) {
		t.Fatal("week 4 should be deload for fixed_3_plus_1")
	}
	res := RecommendNext(Input{
		E1RM:                62.1,
		LastWorkingWeightKg: 50,
		ProgressionPolicy:   defaultPolicy(),
		FatigueModifier:     1,
		ReadinessModifier:   1,
		DeloadActive:        IsDeloadWeek(p, 4),
		DeloadPolicy:        &p,
		UserMetrics:         defaultUserMetrics(),
	})
	if res.TargetWeightKg >= 50 {
		t.Fatalf("deload target = %.1f, want below normal ~48.4", res.TargetWeightKg)
	}
}

func TestExerciseE1RM(t *testing.T) {
	ex := entities.Exercise{
		Sets: []entities.SetEntry{
			{WeightKg: 50, Reps: 8},
			{WeightKg: 100, Reps: 5},
		},
	}
	got := ExerciseE1RM(ex)
	// best set 100x5 -> higher e1RM than 50x8
	if got < 110 {
		t.Fatalf("ExerciseE1RM() = %.1f, expected > 110 from 100x5", got)
	}
}
