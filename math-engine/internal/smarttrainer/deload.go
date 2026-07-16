package smarttrainer

import "github.com/eektheeek/dead-lift-project/math-engine/internal/entities"

// IsDeloadWeek reports whether weekIndex (1-based position in the training cycle) is a deload week.
func IsDeloadWeek(policy entities.DeloadPolicy, weekIndex int) bool {
	if weekIndex < 1 {
		return false
	}
	load, deload := cycleLengths(policy)
	if load < 1 || deload < 1 {
		return false
	}
	cycleLen := load + deload
	pos := (weekIndex-1)%cycleLen + 1
	return pos > load
}

func cycleLengths(policy entities.DeloadPolicy) (loadWeeks, deloadWeeks int) {
	switch policy.Strategy {
	case "fixed_3_plus_1":
		return 3, 1
	case "fixed_2_plus_1":
		return 2, 1
	default:
		return policy.LoadWeeks, policy.DeloadWeeks
	}
}
