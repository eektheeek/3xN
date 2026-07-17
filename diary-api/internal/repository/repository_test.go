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

	updated, err := repo.UpdateExercise(ex.ID, "Pull-up", "back", false)
	if err != nil {
		t.Fatalf("update exercise: %v", err)
	}
	if updated.Name != "Pull-up" || updated.SupportsAssist {
		t.Fatalf("unexpected updated exercise: %+v", updated)
	}
	if updated.Target == nil || updated.Target.AssistKg != 0 {
		t.Fatalf("expected assist cleared on target, got %+v", updated.Target)
	}

	list, err := repo.ListExercises()
	if err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	if len(list) != 1 || list[0].Target == nil || list[0].Target.AssistKg != 0 {
		t.Fatalf("expected cleared assist on listed exercise, got %+v", list)
	}

	session, err := repo.StartWorkoutSession("2026-07-16T18:00:00Z", false)
	if err != nil {
		t.Fatalf("start workout session: %v", err)
	}

	block, err := repo.SaveSessionExercise(session.ID, 1, ex.ID, []repository.CreateSetInput{
		{SetNumber: 1, Reps: 12, WeightKg: 0, AssistKg: 25},
		{SetNumber: 2, Reps: 12, WeightKg: 0, AssistKg: 25},
		{SetNumber: 3, Reps: 12, WeightKg: 0, AssistKg: 25},
	})
	if err != nil {
		t.Fatalf("save session exercise: %v", err)
	}
	if len(block.Sets) != 3 {
		t.Fatalf("unexpected block sets: %+v", block)
	}

	block2, err := repo.SaveSessionExercise(session.ID, 1, ex.ID, []repository.CreateSetInput{
		{SetNumber: 1, Reps: 10, WeightKg: 0, AssistKg: 20},
	})
	if err != nil {
		t.Fatalf("update session exercise: %v", err)
	}
	if len(block2.Sets) != 1 || block2.Sets[0].Reps != 10 {
		t.Fatalf("unexpected updated block: %+v", block2)
	}

	summaries, err := repo.ListWorkoutSessions()
	if err != nil {
		t.Fatalf("list workout sessions: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 session in list, got %+v", summaries)
	}
	if summaries[0].ExerciseCount != 1 || len(summaries[0].ExerciseNames) != 1 {
		t.Fatalf("unexpected summary: %+v", summaries[0])
	}

	finished, err := repo.FinishWorkoutSession(session.ID, 3725)
	if err != nil {
		t.Fatalf("finish workout session: %v", err)
	}
	if finished.DurationSec != 3725 || finished.StartedAt == "" {
		t.Fatalf("unexpected finished session: %+v", finished)
	}

	loaded, err := repo.GetWorkoutSession(session.ID)
	if err != nil {
		t.Fatalf("get workout session: %v", err)
	}
	if loaded.ID != session.ID || len(loaded.Exercises) != 1 || len(loaded.Exercises[0].Sets) != 1 {
		t.Fatalf("unexpected loaded session: %+v", loaded)
	}
	if loaded.Exercises[0].ExerciseName != "Pull-up" {
		t.Fatalf("expected exercise name on session exercise, got %+v", loaded.Exercises[0])
	}

	fullSession, err := repo.CreateWorkoutSession(repository.CreateWorkoutSessionInput{
		PerformedAt: "2026-07-16T19:00:00Z",
		IsDeload:    false,
		Exercises: []repository.CreateWorkoutSessionExerciseInput{
			{
				ExerciseID: ex.ID,
				Sets: []repository.CreateSetInput{
					{SetNumber: 1, Reps: 12, WeightKg: 0, AssistKg: 25},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create workout session: %v", err)
	}
	if len(fullSession.Exercises) != 1 || len(fullSession.Exercises[0].Sets) != 1 {
		t.Fatalf("unexpected session shape: %+v", fullSession)
	}
}
