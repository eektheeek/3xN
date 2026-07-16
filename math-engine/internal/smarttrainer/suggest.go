package smarttrainer

import (
	"fmt"
	"math"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

// Action is the suggested change for the next working weight.
type Action string

const (
	ActionIncrease Action = "increase"
	ActionHold     Action = "hold"
	ActionDecrease Action = "decrease"
)

const (
	defaultFatigueThreshold   = 0.96
	defaultReadinessThreshold = 0.75
)

// Suggestion is the smart-trainer verdict for the next session (linear progression).
type Suggestion struct {
	LastWorkingWeightKg float64
	SuggestedNextWeightKg float64
	MinWeightKg         float64
	MaxWeightKg         float64
	Action              Action
	ReasonCodes         []string
	Message             string
}

// Input holds per-exercise data for linear_step progression.
type Input struct {
	LastWorkingWeightKg float64
	Outcome             entities.ExerciseOutcome
	ProgressionPolicy   entities.ProgressionPolicy
	FatigueModifier     float64
	ReadinessModifier   float64
	DeloadActive        bool
	DeloadPolicy        *entities.DeloadPolicy
	UserMetrics         entities.UserMetrics
}

// SuggestNext computes next working weight from last performance (linear_step only).
func SuggestNext(in Input) Suggestion {
	policy := in.ProgressionPolicy
	last := in.LastWorkingWeightKg
	stepUp := policy.IncreaseStepKg
	stepDown := policy.DecreaseStepKg
	if stepUp <= 0 {
		stepUp = 2.5
	}
	if stepDown <= 0 {
		stepDown = stepUp
	}

	fatigue := in.FatigueModifier
	readiness := in.ReadinessModifier
	if fatigue <= 0 {
		fatigue = 1
	}
	if readiness <= 0 {
		readiness = 1
	}

	reasons := []string{"linear_step_policy"}
	target := last
	action := ActionHold

	outcome := in.Outcome
	fatigueTh, readinessTh := modifierThresholds(in.UserMetrics)

	switch {
	case !outcome.PlanCompleted:
		target = last
		action = ActionHold
		reasons = append(reasons, "plan_not_completed")
	case in.DeloadActive && in.DeloadPolicy != nil:
		target = roundToStep(last*(1-in.DeloadPolicy.IntensityDropPct), stepDown)
		action = ActionDecrease
		if target >= last {
			action = ActionHold
		}
		reasons = append(reasons, "deload_week")
	case fatigue < fatigueTh || readiness < readinessTh:
		target = roundToStep(math.Max(0, last-stepDown), stepDown)
		action = ActionDecrease
		reasons = append(reasons, "recovery_guard")
	case !outcome.ReadyToProgress:
		target = last
		action = ActionHold
		reasons = append(reasons, "not_ready_to_progress")
	default:
		target = roundToStep(last+stepUp, stepUp)
		action = ActionIncrease
		reasons = append(reasons, "plan_completed", "progression_ok")
	}

	minW := math.Max(0, target-stepDown)
	maxW := target + stepUp
	if target < minW {
		minW = target
	}
	if target > maxW {
		maxW = target
	}

	return Suggestion{
		LastWorkingWeightKg: last,
		SuggestedNextWeightKg: target,
		MinWeightKg:         minW,
		MaxWeightKg:         maxW,
		Action:              action,
		ReasonCodes:         reasons,
		Message:             messageFor(action, last, target, stepUp),
	}
}

func messageFor(action Action, last, target, stepUp float64) string {
	switch action {
	case ActionIncrease:
		return fmt.Sprintf("Next time try %.1f kg (+%.1f from %.1f).", target, stepUp, last)
	case ActionDecrease:
		return fmt.Sprintf("Next time try %.1f kg (below %.1f).", target, last)
	default:
		return fmt.Sprintf("Keep %.1f kg for the next session.", last)
	}
}

func modifierThresholds(m entities.UserMetrics) (fatigueTh, readinessTh float64) {
	fatigueTh = m.FatigueThreshold
	if fatigueTh <= 0 {
		fatigueTh = defaultFatigueThreshold
	}
	readinessTh = m.ReadinessThreshold
	if readinessTh <= 0 {
		readinessTh = defaultReadinessThreshold
	}
	return fatigueTh, readinessTh
}

func roundToStep(kg, step float64) float64 {
	if step <= 0 {
		return kg
	}
	return math.Round(kg/step) * step
}
