package smarttrainer

import (
	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

// SessionInput is normalized domain data for a full-session load forecast.
type SessionInput struct {
	Session           entities.WorkoutSession
	UserMetrics       entities.UserMetrics
	ProgressionPolicy entities.ProgressionPolicy
	DeloadPolicy      entities.DeloadPolicy
	TrainingWeekIndex int
	SessionModifiers  entities.SessionModifiers
}

// ExerciseResult is one stack entry: exercise snapshot plus optional next-weight forecast.
type ExerciseResult struct {
	Exercise   entities.Exercise
	Suggestion *Suggestion // working blocks only
}

// SessionResult is the smart-trainer verdict for the whole session.
type SessionResult struct {
	Exercises    []ExerciseResult
	DeloadActive bool
}

// ForecastSession computes next working weights for all exercises in stack order.
func ForecastSession(in SessionInput) SessionResult {
	deloadActive := IsDeloadWeek(in.DeloadPolicy, in.TrainingWeekIndex)
	mods := in.SessionModifiers

	results := make([]ExerciseResult, 0, len(in.Session.Exercises))
	for _, ex := range in.Session.Exercises {
		if ex.ExerciseType != entities.ExerciseTypeWorking {
			results = append(results, ExerciseResult{Exercise: ex})
			continue
		}

		outcome := entities.ExerciseOutcome{}
		if ex.ExerciseOutcome != nil {
			outcome = *ex.ExerciseOutcome
		}

		sug := SuggestNext(Input{
			LastWorkingWeightKg: lastWorkingWeight(ex),
			Outcome:             outcome,
			ProgressionPolicy:   in.ProgressionPolicy,
			FatigueModifier:     mods.FatigueModifier,
			ReadinessModifier:   mods.ReadinessModifier,
			DeloadActive:        deloadActive,
			DeloadPolicy:        &in.DeloadPolicy,
			UserMetrics:         in.UserMetrics,
		})

		results = append(results, ExerciseResult{Exercise: ex, Suggestion: &sug})
	}

	return SessionResult{Exercises: results, DeloadActive: deloadActive}
}

func lastWorkingWeight(ex entities.Exercise) float64 {
	if len(ex.Sets) == 0 {
		return 0
	}
	return ex.Sets[len(ex.Sets)-1].WeightKg
}
