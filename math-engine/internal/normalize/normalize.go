package normalize

import (
	"fmt"
	"time"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

// WorkoutSession maps wire training log to domain session.
func WorkoutSession(sessionLog contracts.RawTrainingLog) (entities.WorkoutSession, error) {
	started, err := time.Parse(time.RFC3339, sessionLog.StartedAt)
	if err != nil {
		return entities.WorkoutSession{}, fmt.Errorf("startedAt: %w", err)
	}
	completed, err := time.Parse(time.RFC3339, sessionLog.CompletedAt)
	if err != nil {
		return entities.WorkoutSession{}, fmt.Errorf("completedAt: %w", err)
	}

	exercises := make([]entities.Exercise, len(sessionLog.Exercises))
	for i, exLog := range sessionLog.Exercises {
		sets := make([]entities.SetEntry, len(exLog.Sets))
		for j, setLog := range exLog.Sets {
			sets[j] = entities.SetEntry{
				SetNumber: setLog.SetNumber,
				WeightKg:  setLog.WeightKg,
				Reps:      setLog.Reps,
			}
		}
		var intervalProto *entities.IntervalProtocol
		if exLog.IntervalProtocol != nil {
			protoLog := exLog.IntervalProtocol
			intervalProto = &entities.IntervalProtocol{
				ProtocolID:  protoLog.ProtocolID,
				WorkSec:     protoLog.WorkSec,
				RestSec:     protoLog.RestSec,
				DisplayName: protoLog.DisplayName,
			}
		}
		var outcome *entities.ExerciseOutcome
		if exLog.ExerciseOutcome != nil {
			o := exLog.ExerciseOutcome
			outcome = &entities.ExerciseOutcome{
				PlanCompleted:   o.PlanCompleted,
				ReadyToProgress: o.ReadyToProgress,
			}
		}
		exercises[i] = entities.Exercise{
			ExerciseID:       exLog.ExerciseID,
			Name:             exLog.Name,
			MuscleGroup:      exLog.MuscleGroup,
			ExerciseType:     exLog.ExerciseType,
			ExerciseOutcome:  outcome,
			Sets:             sets,
			IntervalProtocol: intervalProto,
		}
	}

	return entities.WorkoutSession{
		SessionID:          sessionLog.SessionID,
		UserID:             sessionLog.UserID,
		StartedAt:          started,
		CompletedAt:        completed,
		SessionDurationSec: sessionLog.SessionDurationSec,
		Exercises:          exercises,
	}, nil
}

// SessionModifiers maps wire session modifiers to domain.
func SessionModifiers(m contracts.SessionModifiers) entities.SessionModifiers {
	return entities.SessionModifiers{
		FatigueModifier:   m.FatigueModifier,
		ReadinessModifier: m.ReadinessModifier,
	}
}

// UserMetrics maps wire profile to domain metrics.
func UserMetrics(m contracts.UserMetrics) entities.UserMetrics {
	return entities.UserMetrics{
		Age:                 m.Age,
		BodyWeightKg:        m.BodyWeightKg,
		RecoverySensitivity: m.RecoverySensitivity,
		FatigueThreshold:    m.FatigueThreshold,
		ReadinessThreshold:  m.ReadinessThreshold,
	}
}

// ProgressionPolicy maps wire policy to domain policy.
func ProgressionPolicy(p contracts.ProgressionPolicy) entities.ProgressionPolicy {
	return entities.ProgressionPolicy{
		Strategy:       p.Strategy,
		TargetPercent:  p.TargetPercent,
		IncreaseStepKg: p.IncreaseStepKg,
		DecreaseStepKg: p.DecreaseStepKg,
	}
}

// DeloadPolicy maps wire deload policy to domain policy.
func DeloadPolicy(p contracts.DeloadPolicy) entities.DeloadPolicy {
	return entities.DeloadPolicy{
		Strategy:         p.Strategy,
		LoadWeeks:        p.LoadWeeks,
		DeloadWeeks:      p.DeloadWeeks,
		IntensityDropPct: p.IntensityDropPct,
		VolumeDropPct:    p.VolumeDropPct,
		MinRIR:           p.MinRIR,
	}
}
