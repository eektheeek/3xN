package smarttrainer

import (
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

func TestForecastSession_increaseOnWorking(t *testing.T) {
	outcome := &entities.ExerciseOutcome{PlanCompleted: true, ReadyToProgress: true}
	session := entities.WorkoutSession{
		Exercises: []entities.Exercise{
			{ExerciseID: "w", ExerciseType: entities.ExerciseTypeWarmup, Name: "Warm"},
			{
				ExerciseID:      "sq",
				ExerciseType:    entities.ExerciseTypeWorking,
				Name:            "Squat",
				ExerciseOutcome: outcome,
				Sets:            []entities.SetEntry{{WeightKg: 100, Reps: 5}},
			},
		},
	}

	res := ForecastSession(SessionInput{
		Session: session,
		ProgressionPolicy: entities.ProgressionPolicy{
			Strategy:        "linear_step",
			IncreaseStepKg:  2.5,
			DecreaseStepKg:  2.5,
		},
		DeloadPolicy: entities.DeloadPolicy{Strategy: "fixed_3_plus_1", LoadWeeks: 3, DeloadWeeks: 1},
		TrainingWeekIndex: 1,
		SessionModifiers:  entities.SessionModifiers{FatigueModifier: 1, ReadinessModifier: 1},
	})

	if len(res.Exercises) != 2 {
		t.Fatalf("exercises = %d", len(res.Exercises))
	}
	if res.Exercises[0].Suggestion != nil {
		t.Fatal("warmup must not have suggestion")
	}
	sug := res.Exercises[1].Suggestion
	if sug == nil || sug.SuggestedNextWeightKg != 102.5 || sug.Action != ActionIncrease {
		t.Fatalf("suggestion = %+v", sug)
	}
}

func TestForecastSession_deloadWeek(t *testing.T) {
	outcome := &entities.ExerciseOutcome{PlanCompleted: true, ReadyToProgress: true}
	session := entities.WorkoutSession{
		Exercises: []entities.Exercise{{
			ExerciseType:    entities.ExerciseTypeWorking,
			ExerciseOutcome: outcome,
			Sets:            []entities.SetEntry{{WeightKg: 100, Reps: 5}},
		}},
	}

	res := ForecastSession(SessionInput{
		Session: session,
		ProgressionPolicy: entities.ProgressionPolicy{IncreaseStepKg: 2.5, DecreaseStepKg: 2.5},
		DeloadPolicy: entities.DeloadPolicy{
			Strategy:         "fixed_3_plus_1",
			LoadWeeks:        3,
			DeloadWeeks:      1,
			IntensityDropPct: 0.1,
		},
		TrainingWeekIndex: 4,
		SessionModifiers:  entities.SessionModifiers{FatigueModifier: 1, ReadinessModifier: 1},
	})

	if !res.DeloadActive {
		t.Fatal("expected deload week")
	}
	if res.Exercises[0].Suggestion.Action != ActionDecrease {
		t.Fatalf("action = %s", res.Exercises[0].Suggestion.Action)
	}
}
