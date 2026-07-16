package repository_test

import (
	"path/filepath"
	"testing"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/db"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

func TestExerciseTargetAndWorkoutSession(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	repo := repository.New(sqlDB)

	ex, err := repo.CreateExercise("Wide grip pull-up", "back", true)
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	target, err := repo.SetTarget(ex.ID, 3, 12, 0, 25)
	if err != nil {
		t.Fatalf("set target: %v", err)
	}
	if target.AssistKg != 25 || target.TargetSets != 3 {
		t.Fatalf("unexpected target: %+v", target)
	}

	got, err := repo.GetExercise(ex.ID)
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if got.Target == nil || got.Target.AssistKg != 25 {
		t.Fatalf("expected target on exercise, got %+v", got)
	}

	session, err := repo.CreateWorkoutSession(repository.CreateWorkoutSessionInput{
		PerformedAt: "2026-07-16T18:00:00Z",
		IsDeload:    false,
		Exercises: []repository.CreateWorkoutSessionExerciseInput{
			{
				ExerciseID: ex.ID,
				Sets: []repository.CreateSetInput{
					{SetNumber: 1, Reps: 12, WeightKg: 0, AssistKg: 25},
					{SetNumber: 2, Reps: 12, WeightKg: 0, AssistKg: 25},
					{SetNumber: 3, Reps: 12, WeightKg: 0, AssistKg: 25},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create workout session: %v", err)
	}
	if len(session.Exercises) != 1 || len(session.Exercises[0].Sets) != 3 {
		t.Fatalf("unexpected session shape: %+v", session)
	}

	loaded, err := repo.GetWorkoutSession(session.ID)
	if err != nil {
		t.Fatalf("get workout session: %v", err)
	}
	if loaded.ID != session.ID || len(loaded.Exercises[0].Sets) != 3 {
		t.Fatalf("unexpected loaded session: %+v", loaded)
	}
}
