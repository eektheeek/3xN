package smarttrainer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/normalize"
)

const outputVersion = "analysis_result_v1"

// Run normalizes wire input, forecasts next loads, attaches volume metrics, returns CoreOutput.
func Run(in contracts.CoreInput) (contracts.CoreOutput, error) {
	session, err := normalize.WorkoutSession(in.RawTrainingLog)
	if err != nil {
		return contracts.CoreOutput{}, err
	}

	forecast := ForecastSession(SessionInput{
		Session:           session,
		UserMetrics:       normalize.UserMetrics(in.UserMetrics),
		ProgressionPolicy: normalize.ProgressionPolicy(in.ProgressionPolicy),
		DeloadPolicy:      normalize.DeloadPolicy(in.DeloadPolicy),
		TrainingWeekIndex: in.TrainingWeekIndex,
		SessionModifiers:  normalize.SessionModifiers(in.SessionModifiers),
	})

	sessionVol := SessionWorkingVolumeLoad(session)
	exerciseMetrics := make([]contracts.ExerciseMetric, 0, len(forecast.Exercises))

	for _, er := range forecast.Exercises {
		ex := er.Exercise
		em := contracts.ExerciseMetric{
			ExerciseID:   ex.ExerciseID,
			ExerciseName: ex.Name,
			ExerciseType: ex.ExerciseType,
		}
		if er.Suggestion != nil {
			sug := er.Suggestion
			em.VolumeLoad = &contracts.MetricValue{
				Value:         VolumeLoad(ex),
				Unit:          "kg_reps",
				Confidence:    0.97,
				Applicability: "ready",
			}
			em.SmartTrainer = toSmartTrainerOut(sug)
		}
		exerciseMetrics = append(exerciseMetrics, em)
	}

	mods := normalize.SessionModifiers(in.SessionModifiers)
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
			FatigueModifier:   mods.FatigueModifier,
			ReadinessModifier: mods.ReadinessModifier,
			DeloadActive:      forecast.DeloadActive,
			TrainingWeekIndex: in.TrainingWeekIndex,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// toSmartTrainerOut maps internal progression result to the API response field.
func toSmartTrainerOut(sug *Suggestion) *contracts.SmartTrainerOut {
	return &contracts.SmartTrainerOut{
		LastWorkingWeightKg:   sug.LastWorkingWeightKg,
		SuggestedNextWeightKg: sug.SuggestedNextWeightKg,
		MinWeightKg:           sug.MinWeightKg,
		MaxWeightKg:           sug.MaxWeightKg,
		Action:                string(sug.Action),
		ReasonCodes:           sug.ReasonCodes,
		Message:               sug.Message,
	}
}

// ParseAndRun decodes JSON bytes into CoreInput and runs the smart-trainer pipeline.
func ParseAndRun(data []byte) (contracts.CoreOutput, error) {
	var in contracts.CoreInput
	if err := json.Unmarshal(data, &in); err != nil {
		return contracts.CoreOutput{}, fmt.Errorf("decode input: %w", err)
	}
	return Run(in)
}
