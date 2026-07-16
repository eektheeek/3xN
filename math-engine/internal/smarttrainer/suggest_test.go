package smarttrainer

import (
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

func linearPolicy() entities.ProgressionPolicy {
	return entities.ProgressionPolicy{
		Strategy:       "linear_step",
		IncreaseStepKg: 2.5,
		DecreaseStepKg: 2.5,
	}
}

func defaultMetrics() entities.UserMetrics {
	return entities.UserMetrics{
		FatigueThreshold:   0.96,
		ReadinessThreshold: 0.75,
	}
}

func TestSuggestNext_increase_after_100x5(t *testing.T) {
	res := SuggestNext(Input{
		LastWorkingWeightKg: 100,
		Outcome: entities.ExerciseOutcome{
			PlanCompleted: true, ReadyToProgress: true,
		},
		ProgressionPolicy: linearPolicy(),
		FatigueModifier:   0.97,
		ReadinessModifier: 0.99,
		UserMetrics:       defaultMetrics(),
	})
	if res.SuggestedNextWeightKg != 102.5 {
		t.Fatalf("target = %.1f, want 102.5", res.SuggestedNextWeightKg)
	}
	if res.Action != ActionIncrease {
		t.Fatalf("action = %q, want increase", res.Action)
	}
}

func TestSuggestNext_hold_plan_not_completed(t *testing.T) {
	res := SuggestNext(Input{
		LastWorkingWeightKg: 100,
		Outcome:             entities.ExerciseOutcome{PlanCompleted: false},
		ProgressionPolicy:   linearPolicy(),
		FatigueModifier:     1,
		ReadinessModifier:   1,
		UserMetrics:         defaultMetrics(),
	})
	if res.SuggestedNextWeightKg != 100 || res.Action != ActionHold {
		t.Fatalf("target=%.1f action=%q, want 100 hold", res.SuggestedNextWeightKg, res.Action)
	}
}

func TestSuggestNext_hold_not_ready(t *testing.T) {
	res := SuggestNext(Input{
		LastWorkingWeightKg: 100,
		Outcome: entities.ExerciseOutcome{
			PlanCompleted: true, ReadyToProgress: false,
		},
		ProgressionPolicy: linearPolicy(),
		FatigueModifier:   1,
		ReadinessModifier: 1,
		UserMetrics:       defaultMetrics(),
	})
	if res.Action != ActionHold || res.SuggestedNextWeightKg != 100 {
		t.Fatalf("target=%.1f action=%q, want 100 hold", res.SuggestedNextWeightKg, res.Action)
	}
}

func TestSuggestNext_decrease_recovery(t *testing.T) {
	res := SuggestNext(Input{
		LastWorkingWeightKg: 100,
		Outcome: entities.ExerciseOutcome{
			PlanCompleted: true, ReadyToProgress: true,
		},
		ProgressionPolicy: linearPolicy(),
		FatigueModifier:   0.94,
		ReadinessModifier: 1,
		UserMetrics:       defaultMetrics(),
	})
	if res.Action != ActionDecrease || res.SuggestedNextWeightKg != 97.5 {
		t.Fatalf("target=%.1f action=%q, want 97.5 decrease", res.SuggestedNextWeightKg, res.Action)
	}
}

func TestSuggestNext_deload_week(t *testing.T) {
	deload := &entities.DeloadPolicy{
		Strategy: "fixed_3_plus_1", IntensityDropPct: 0.10,
	}
	res := SuggestNext(Input{
		LastWorkingWeightKg: 100,
		Outcome: entities.ExerciseOutcome{
			PlanCompleted: true, ReadyToProgress: true,
		},
		ProgressionPolicy: linearPolicy(),
		FatigueModifier:   1,
		ReadinessModifier: 1,
		DeloadActive:      true,
		DeloadPolicy:      deload,
		UserMetrics:       defaultMetrics(),
	})
	if res.SuggestedNextWeightKg != 90 {
		t.Fatalf("target = %.1f, want 90 (100 - 10%%)", res.SuggestedNextWeightKg)
	}
	if res.Action != ActionDecrease {
		t.Fatalf("action = %q, want decrease", res.Action)
	}
}
