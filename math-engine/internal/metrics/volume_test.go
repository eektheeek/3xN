package metrics

import (
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

func TestVolumeLoad_exercise(t *testing.T) {
	ex := entities.Exercise{
		Sets: []entities.SetEntry{
			{WeightKg: 100, Reps: 5},
			{WeightKg: 100, Reps: 5},
			{WeightKg: 50, Reps: 10},
		},
	}
	got := VolumeLoad(ex)
	want := 100*5 + 100*5 + 50*10 // 1500
	if got != float64(want) {
		t.Fatalf("VolumeLoad() = %v, want %v", got, want)
	}
}

func TestSessionVolumeLoad(t *testing.T) {
	session := entities.WorkoutSession{
		Exercises: []entities.Exercise{
			{Sets: []entities.SetEntry{{WeightKg: 100, Reps: 5}}},
			{Sets: []entities.SetEntry{{WeightKg: 50, Reps: 8}}},
		},
	}
	got := SessionVolumeLoad(session)
	want := 100*5 + 50*8 // 900
	if got != float64(want) {
		t.Fatalf("SessionVolumeLoad() = %v, want %v", got, want)
	}
}

func TestSessionWorkingVolumeLoad_skipsWarmup(t *testing.T) {
	session := entities.WorkoutSession{
		Exercises: []entities.Exercise{
			{BlockType: entities.BlockTypeWarmup, Sets: []entities.SetEntry{{WeightKg: 16, Reps: 12}}},
			{BlockType: entities.BlockTypeWorking, Sets: []entities.SetEntry{{WeightKg: 100, Reps: 5}}},
		},
	}
	got := SessionWorkingVolumeLoad(session)
	if got != 500 {
		t.Fatalf("SessionWorkingVolumeLoad() = %v, want 500", got)
	}
}
