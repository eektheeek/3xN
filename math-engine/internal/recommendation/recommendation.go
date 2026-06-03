package recommendation

import (
	"math"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/formulas"
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

// Result is a next-session weight recommendation for one exercise.
type Result struct {
	Action         Action
	TargetWeightKg float64
	MinWeightKg    float64
	MaxWeightKg    float64
	ReasonCodes    []string
}

// Input holds data needed to recommend the next working weight.
type Input struct {
	E1RM                float64
	LastWorkingWeightKg float64
	ProgressionPolicy   entities.ProgressionPolicy
	FatigueModifier     float64
	ReadinessModifier   float64
	DeloadActive        bool
	DeloadPolicy        *entities.DeloadPolicy
	UserMetrics         entities.UserMetrics
}

// RecommendNext computes target weight and action from e1RM and user policy.
func RecommendNext(in Input) Result {
	policy := in.ProgressionPolicy
	fatigue := in.FatigueModifier
	readiness := in.ReadinessModifier
	if fatigue <= 0 {
		fatigue = 1
	}
	if readiness <= 0 {
		readiness = 1
	}

	target := in.E1RM * policy.TargetPercent * fatigue * readiness
	reasons := []string{"percent_e1rm_policy"}

	if in.DeloadActive && in.DeloadPolicy != nil {
		target *= 1 - in.DeloadPolicy.IntensityDropPct
		reasons = append(reasons, "deload_week")
	}

	target = roundToStep(target, policy.IncreaseStepKg)
	minW := math.Max(0, target-policy.DecreaseStepKg)
	maxW := target + policy.IncreaseStepKg

	fatigueTh, readinessTh := modifierThresholds(in.UserMetrics)

	action := ActionHold
	last := in.LastWorkingWeightKg

	switch {
	case fatigue < fatigueTh || readiness < readinessTh:
		action = ActionDecrease
		target = roundToStep(target-policy.DecreaseStepKg, policy.DecreaseStepKg)
		reasons = append(reasons, "recovery_guard")
	case last > 0 && target >= last+policy.IncreaseStepKg:
		action = ActionIncrease
		reasons = append(reasons, "progression_ok")
	case last > 0 && target <= last-policy.DecreaseStepKg:
		action = ActionDecrease
		reasons = append(reasons, "below_last_performance")
	default:
		reasons = append(reasons, "baseline_ready")
	}

	if target < minW {
		minW = target
	}
	if target > maxW {
		maxW = target
	}

	return Result{
		Action:         action,
		TargetWeightKg: target,
		MinWeightKg:    minW,
		MaxWeightKg:    maxW,
		ReasonCodes:    reasons,
	}
}

// ExerciseE1RM returns the highest ensemble e1RM across all sets in an exercise.
func ExerciseE1RM(ex entities.Exercise) float64 {
	var best float64
	for _, set := range ex.Sets {
		e1rm := formulas.Estimate1RMEnsemble(set.WeightKg, set.Reps)
		if e1rm > best {
			best = e1rm
		}
	}
	return best
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
