package smarttrainer

import "github.com/eektheeek/dead-lift-project/math-engine/internal/entities"

// VolumeLoad returns sum(weightKg * reps) for all sets in an exercise.
func VolumeLoad(ex entities.Exercise) float64 {
	var total float64
	for _, set := range ex.Sets {
		total += set.WeightKg * float64(set.Reps)
	}
	return total
}

// SessionVolumeLoad returns total volume across all exercises (includes warmup).
func SessionVolumeLoad(session entities.WorkoutSession) float64 {
	var total float64
	for _, ex := range session.Exercises {
		total += VolumeLoad(ex)
	}
	return total
}

// SessionWorkingVolumeLoad returns volume for working blocks only.
func SessionWorkingVolumeLoad(session entities.WorkoutSession) float64 {
	var total float64
	for _, ex := range session.Exercises {
		if ex.ExerciseType != entities.ExerciseTypeWorking {
			continue
		}
		total += VolumeLoad(ex)
	}
	return total
}
