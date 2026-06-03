package normalize

import (
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
)

func TestWorkoutSession_mapsExercises(t *testing.T) {
	log := contracts.RawTrainingLog{
		SessionID:          "s1",
		UserID:             "u1",
		StartedAt:          "2026-05-28T17:00:00Z",
		CompletedAt:        "2026-05-28T18:00:00Z",
		SessionDurationSec: 3600,
		Exercises: []contracts.Exercise{{
			ExerciseID: "ex1", Name: "Squat", MuscleGroup: "legs", BlockType: "working",
			Sets: []contracts.SetEntry{{SetNumber: 1, WeightKg: 100, Reps: 5}},
		}},
	}
	session, err := WorkoutSession(log)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Exercises) != 1 || session.Exercises[0].BlockType != "working" || session.Exercises[0].Sets[0].WeightKg != 100 {
		t.Fatalf("unexpected session: %+v", session)
	}
}
