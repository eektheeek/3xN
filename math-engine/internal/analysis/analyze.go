package analysis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/formulas"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/metrics"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/normalize"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/recommendation"
)

const outputVersion = "analysis_result_v1"

// Run performs a minimal v1 analysis from validated core input.
func Run(in contracts.CoreInput) (contracts.CoreOutput, error) {
	session, err := normalize.WorkoutSession(in.RawTrainingLog)
	if err != nil {
		return contracts.CoreOutput{}, err
	}

	userMetrics := normalize.UserMetrics(in.UserMetrics)
	progression := normalize.ProgressionPolicy(in.ProgressionPolicy)
	deload := normalize.DeloadPolicy(in.DeloadPolicy)

	deloadActive := recommendation.IsDeloadWeek(deload, in.TrainingWeekIndex)
	sessionMods := normalize.SessionModifiers(in.SessionModifiers)

	sessionVol := metrics.SessionWorkingVolumeLoad(session)
	exerciseMetrics := make([]contracts.ExerciseMetric, 0, len(session.Exercises))

	for _, ex := range session.Exercises {
		if ex.BlockType != entities.BlockTypeWorking {
			exerciseMetrics = append(exerciseMetrics, warmupExerciseMetric(ex))
			continue
		}

		em, e1rm := buildWorkingExerciseMetric(ex)
		rec := recommendation.RecommendNext(recommendation.Input{
			E1RM:                e1rm,
			LastWorkingWeightKg: lastWorkingWeight(ex),
			ProgressionPolicy:   progression,
			FatigueModifier:     sessionMods.FatigueModifier,
			ReadinessModifier:   sessionMods.ReadinessModifier,
			DeloadActive:        deloadActive,
			DeloadPolicy:        &deload,
			UserMetrics:         userMetrics,
		})
		em.Recommendation = &contracts.ExerciseRecommendationOut{
			Action:         string(rec.Action),
			TargetWeightKg: rec.TargetWeightKg,
			MinWeightKg:    rec.MinWeightKg,
			MaxWeightKg:    rec.MaxWeightKg,
			ReasonCodes:    rec.ReasonCodes,
		}
		exerciseMetrics = append(exerciseMetrics, em)
	}

	return contracts.CoreOutput{
		Version:     outputVersion,
		SessionID:   session.SessionID,
		SessionDate: session.StartedAt.Format("2006-01-02"),
		UserID:      session.UserID,
		ComputedMetrics: contracts.ComputedMetrics{
			VolumeLoad: &contracts.MetricValue{
				Value:         sessionVol,
				Unit:          "kg_reps",
				Confidence:    0.98,
				Applicability: "ready",
			},
		},
		ExerciseMetrics: exerciseMetrics,
		SessionContext: contracts.SessionContext{
			TargetPercent:     progression.TargetPercent,
			FatigueModifier:   sessionMods.FatigueModifier,
			ReadinessModifier: sessionMods.ReadinessModifier,
			DeloadActive:      deloadActive,
			TrainingWeekIndex: in.TrainingWeekIndex,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func warmupExerciseMetric(ex entities.Exercise) contracts.ExerciseMetric {
	return contracts.ExerciseMetric{
		ExerciseID:   ex.ExerciseID,
		ExerciseName: ex.Name,
		BlockType:    ex.BlockType,
	}
}

func buildWorkingExerciseMetric(ex entities.Exercise) (contracts.ExerciseMetric, float64) {
	bestWeight, bestReps := bestSet(ex)
	epley := formulas.Estimate1RMEpley(bestWeight, bestReps)
	brzycki := formulas.Estimate1RMBrzycki(bestWeight, bestReps)
	lombardi := formulas.Estimate1RMLombardi(bestWeight, bestReps)
	e1rm := formulas.Estimate1RMEnsemble(bestWeight, bestReps)

	return contracts.ExerciseMetric{
		ExerciseID:   ex.ExerciseID,
		ExerciseName: ex.Name,
		BlockType:    ex.BlockType,
		E1RM: &contracts.E1RMMetric{
			MetricValue: contracts.MetricValue{
				Value:         e1rm,
				Unit:          "kg",
				Confidence:    0.89,
				Applicability: applicabilityForE1RM(bestWeight, bestReps),
			},
			FormulaFamily: "ensemble",
			FormulaBreakdown: map[string]float64{
				"epley":    epley,
				"brzycki":  brzycki,
				"lombardi": lombardi,
			},
		},
		VolumeLoad: &contracts.MetricValue{
			Value:         metrics.VolumeLoad(ex),
			Unit:          "kg_reps",
			Confidence:    0.97,
			Applicability: "ready",
		},
	}, e1rm
}

func bestSet(ex entities.Exercise) (weightKg float64, reps int) {
	var bestE1RM float64
	for _, s := range ex.Sets {
		e := formulas.Estimate1RMEnsemble(s.WeightKg, s.Reps)
		if e > bestE1RM {
			bestE1RM = e
			weightKg = s.WeightKg
			reps = s.Reps
		}
	}
	return weightKg, reps
}

func lastWorkingWeight(ex entities.Exercise) float64 {
	if len(ex.Sets) == 0 {
		return 0
	}
	last := ex.Sets[len(ex.Sets)-1]
	return last.WeightKg
}

func applicabilityForE1RM(weightKg float64, reps int) string {
	if weightKg <= 0 || reps < 1 {
		return "not_applicable"
	}
	if reps > 12 {
		return "partial"
	}
	return "ready"
}

// ParseAndRun decodes JSON bytes into CoreInput and runs analysis.
func ParseAndRun(data []byte) (contracts.CoreOutput, error) {
	var in contracts.CoreInput
	if err := json.Unmarshal(data, &in); err != nil {
		return contracts.CoreOutput{}, fmt.Errorf("decode input: %w", err)
	}
	return Run(in)
}
