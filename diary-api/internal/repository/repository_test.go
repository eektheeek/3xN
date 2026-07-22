package repository_test

import (
	"path/filepath"
	"testing"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/db"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
	"github.com/google/uuid"
)

func TestExerciseTargetAndWorkoutSession(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	repo := repository.New(sqlDB)

	ex, err := repo.CreateExercise("Wide grip pull-up", "back", "reps", true, "")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	proto, err := repo.CreateIntervalProtocol(repository.CreateIntervalProtocolInput{
		Name:        "40/20 + warmup",
		PrepareSec:  5,
		WorkSec:     40,
		RestSec:     20,
		WarmupExtra: true,
	})
	if err != nil {
		t.Fatalf("create protocol: %v", err)
	}

	target, err := repo.SetTarget(ex.ID, 3, 12, 0, 0, 25)
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
	if got.Kind != "reps" {
		t.Fatalf("expected kind reps, got %+v", got)
	}

	updated, err := repo.UpdateExercise(ex.ID, "Pull-up", "back", "reps", false, proto.ID)
	if err != nil {
		t.Fatalf("update exercise: %v", err)
	}
	if updated.Name != "Pull-up" || updated.SupportsAssist {
		t.Fatalf("unexpected updated exercise: %+v", updated)
	}
	if updated.Target == nil || updated.Target.AssistKg != 0 {
		t.Fatalf("expected assist cleared on target, got %+v", updated.Target)
	}
	if updated.Protocol == nil || updated.Protocol.ID != proto.ID || !updated.Protocol.WarmupExtra {
		t.Fatalf("expected protocol on exercise, got %+v", updated.Protocol)
	}

	list, err := repo.ListExercises()
	if err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	if len(list) != 1 || list[0].Target == nil || list[0].Target.AssistKg != 0 {
		t.Fatalf("expected cleared assist on listed exercise, got %+v", list)
	}

	session, _, err := repo.StartWorkoutSession(uuid.NewString(), "2026-07-16T18:00:00Z", false, "")
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

	finished, err := repo.FinishWorkoutSession(session.ID, 3725, "")
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
	if loaded.Exercises[0].Kind != "reps" {
		t.Fatalf("expected kind reps on session exercise, got %+v", loaded.Exercises[0])
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

func TestCycleStepsAndAdvance(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "cycle.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repo := repository.New(sqlDB)

	ex, err := repo.CreateExercise("Squat", "legs", "reps", false, "")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	planA, err := repo.CreateWorkoutPlan(repository.CreateWorkoutPlanInput{
		Name: "A", ExerciseIDs: []string{ex.ID},
	})
	if err != nil {
		t.Fatalf("create plan A: %v", err)
	}
	planB, err := repo.CreateWorkoutPlan(repository.CreateWorkoutPlanInput{
		Name: "B", ExerciseIDs: []string{ex.ID},
	})
	if err != nil {
		t.Fatalf("create plan B: %v", err)
	}
	planC, err := repo.CreateWorkoutPlan(repository.CreateWorkoutPlanInput{
		Name: "C", ExerciseIDs: []string{ex.ID},
	})
	if err != nil {
		t.Fatalf("create plan C: %v", err)
	}

	home, err := repo.CreateCycle("Дом")
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	outdoor, err := repo.CreateCycle("Улица")
	if err != nil {
		t.Fatalf("create outdoor cycle: %v", err)
	}
	if home.ID == outdoor.ID {
		t.Fatalf("expected distinct cycle ids")
	}

	list, err := repo.ListCycles()
	if err != nil || len(list) != 2 {
		t.Fatalf("list cycles: %v %+v", err, list)
	}

	empty, err := repo.GetCycle(home.ID)
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	if len(empty.Steps) != 0 || empty.CurrentStep != 1 || empty.Name != "Дом" {
		t.Fatalf("unexpected empty cycle: %+v", empty)
	}

	if empty.OnHome {
		t.Fatalf("new cycle should not be on home: %+v", empty)
	}

	cycle, err := repo.ReplaceCycleSteps(home.ID, []string{planA.ID, planB.ID, planC.ID})
	if err != nil {
		t.Fatalf("replace steps: %v", err)
	}
	if len(cycle.Steps) != 3 || cycle.CurrentStep != 1 {
		t.Fatalf("unexpected cycle after replace: %+v", cycle)
	}
	if cycle.Steps[0].WorkoutPlanName != "A" || cycle.Steps[2].Position != 3 {
		t.Fatalf("unexpected steps: %+v", cycle.Steps)
	}

	cycle, err = repo.SetCycleOnHome(home.ID, true)
	if err != nil || !cycle.OnHome {
		t.Fatalf("set on home: %+v %v", cycle, err)
	}
	cycle, err = repo.SetCycleOnHome(home.ID, false)
	if err != nil || cycle.OnHome {
		t.Fatalf("unset on home: %+v %v", cycle, err)
	}

	cycle, err = repo.AdvanceCycle(home.ID, "")
	if err != nil {
		t.Fatalf("advance: %v", err)
	}
	if cycle.CurrentStep != 2 {
		t.Fatalf("expected step 2, got %d", cycle.CurrentStep)
	}
	if !cycle.OnHome {
		t.Fatalf("advance should pin cycle to home: %+v", cycle)
	}

	session, _, err := repo.StartWorkoutSession(uuid.NewString(), "2026-07-20T10:00:00Z", false, planB.ID)
	if err != nil {
		t.Fatalf("start session for link: %v", err)
	}
	// Finish current cycle step (plan B is step 2) — advances in the same call
	finished, err := repo.FinishWorkoutSession(session.ID, 600, home.ID)
	if err != nil {
		t.Fatalf("finish+advance: %v", err)
	}
	if finished.CycleID != home.ID || finished.CycleStep != 2 {
		t.Fatalf("expected cycle link on finished session, got %+v", finished)
	}
	cycle, err = repo.GetCycle(home.ID)
	if err != nil {
		t.Fatalf("get cycle after finish: %v", err)
	}
	if cycle.CurrentStep != 3 {
		t.Fatalf("expected step 3 after finish+advance, got %d", cycle.CurrentStep)
	}
	linked := cycle.Steps[1]
	if linked.CompletedSessionID != session.ID || linked.CompletedPerformedAt == "" {
		t.Fatalf("expected step 2 linked to session, got %+v", linked)
	}
	cycle, err = repo.AdvanceCycle(home.ID, "")
	if err != nil {
		t.Fatalf("advance at end: %v", err)
	}
	if !cycle.Completed || cycle.CurrentStep != 4 {
		t.Fatalf("expected completed after last step, got %+v", cycle)
	}
	if cycle.OnHome {
		t.Fatalf("completed cycle should leave home: %+v", cycle)
	}

	repeated, err := repo.RepeatCycle(home.ID)
	if err != nil {
		t.Fatalf("repeat: %v", err)
	}
	if repeated.ID == home.ID || repeated.Completed || repeated.CurrentStep != 1 || !repeated.OnHome {
		t.Fatalf("unexpected repeated cycle: %+v", repeated)
	}
	if len(repeated.Steps) != 3 || repeated.Name != "Дом" {
		t.Fatalf("repeated steps/name: %+v", repeated)
	}
	still, err := repo.GetCycle(home.ID)
	if err != nil || !still.Completed {
		t.Fatalf("original should stay completed: %+v %v", still, err)
	}

	// Shrink a completed clone for cursor normalize check
	_, _ = repo.AdvanceCycle(repeated.ID, "")
	_, _ = repo.AdvanceCycle(repeated.ID, "")
	cycle, err = repo.AdvanceCycle(repeated.ID, "")
	if err != nil || !cycle.Completed {
		t.Fatalf("re-complete clone: %+v %v", cycle, err)
	}

	// Outdoor cycle stays independent
	out, err := repo.GetCycle(outdoor.ID)
	if err != nil || out.CurrentStep != 1 || len(out.Steps) != 0 || out.Completed {
		t.Fatalf("outdoor should be untouched: %+v %v", out, err)
	}

	cycle, err = repo.ReplaceCycleSteps(repeated.ID, []string{planB.ID})
	if err != nil {
		t.Fatalf("shrink steps: %v", err)
	}
	// Was completed (step 4); after shrink to 1 step → stay completed at step 2
	if !cycle.Completed || cycle.CurrentStep != 2 || len(cycle.Steps) != 1 {
		t.Fatalf("unexpected after shrink: %+v", cycle)
	}

	// Original completed history untouched
	still, err = repo.GetCycle(home.ID)
	if err != nil || !still.Completed || len(still.Steps) != 3 {
		t.Fatalf("original history changed: %+v %v", still, err)
	}

	session2, _, err := repo.StartWorkoutSession(uuid.NewString(), "2026-07-20T11:00:00Z", false, planB.ID)
	if err != nil {
		t.Fatalf("start with plan: %v", err)
	}
	loaded, err := repo.GetWorkoutSession(session2.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if loaded.WorkoutPlanID != planB.ID {
		t.Fatalf("expected workoutPlanId %s, got %q", planB.ID, loaded.WorkoutPlanID)
	}
}

func TestHoldExerciseTargetAndSets(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "hold.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repo := repository.New(sqlDB)

	ex, err := repo.CreateExercise("Bridge hold", "core", "hold", true, "")
	if err != nil {
		t.Fatalf("create hold exercise: %v", err)
	}
	if ex.Kind != "hold" || ex.SupportsAssist {
		t.Fatalf("hold should force assist off: %+v", ex)
	}

	target, err := repo.SetTarget(ex.ID, 3, 12, 60, 10, 5)
	if err != nil {
		t.Fatalf("set hold target: %v", err)
	}
	if target.HoldSec != 60 || target.TargetReps != 0 || target.WeightKg != 10 || target.AssistKg != 0 {
		t.Fatalf("unexpected hold target: %+v", target)
	}

	session, _, err := repo.StartWorkoutSession(uuid.NewString(), "2026-07-21T12:00:00Z", false, "")
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	block, err := repo.SaveSessionExercise(session.ID, 1, ex.ID, []repository.CreateSetInput{
		{SetNumber: 1, DurationSec: 60},
		{SetNumber: 2, DurationSec: 45},
		{SetNumber: 3, DurationSec: 50},
	})
	if err != nil {
		t.Fatalf("save hold sets: %v", err)
	}
	if len(block.Sets) != 3 || block.Sets[1].DurationSec != 45 {
		t.Fatalf("unexpected hold sets: %+v", block.Sets)
	}

	loaded, err := repo.GetWorkoutSession(session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if loaded.Exercises[0].Kind != "hold" {
		t.Fatalf("expected hold kind on block: %+v", loaded.Exercises[0])
	}
	sum := 0
	for _, s := range loaded.Exercises[0].Sets {
		sum += s.DurationSec
	}
	if sum != 155 {
		t.Fatalf("expected total hold 155s, got %d", sum)
	}
}

func TestStartWorkoutSessionClientIDRequiredAndIdempotent(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "start-id.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repo := repository.New(sqlDB)

	if _, _, err := repo.StartWorkoutSession("", "2026-07-22T12:00:00Z", false, ""); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, _, err := repo.StartWorkoutSession("not-a-uuid", "2026-07-22T12:00:00Z", false, ""); err == nil {
		t.Fatal("expected error for invalid uuid")
	}

	id := uuid.NewString()
	first, created, err := repo.StartWorkoutSession(id, "2026-07-22T12:00:00Z", false, "")
	if err != nil || !created {
		t.Fatalf("first start: created=%v err=%v", created, err)
	}
	if first.ID != id {
		t.Fatalf("expected client id %s, got %s", id, first.ID)
	}

	second, createdAgain, err := repo.StartWorkoutSession(id, "2026-07-22T13:00:00Z", true, "")
	if err != nil || createdAgain {
		t.Fatalf("replay start: created=%v err=%v", createdAgain, err)
	}
	if second.ID != id {
		t.Fatalf("replay id mismatch: %s", second.ID)
	}
	if second.IsDeload != first.IsDeload {
		t.Fatalf("idempotent replay should return original row, got isDeload=%v", second.IsDeload)
	}

	loaded, err := repo.GetWorkoutSession(id)
	if err != nil {
		t.Fatalf("get after replay: %v", err)
	}
	if loaded.ID != id || loaded.DurationSec != 0 {
		t.Fatalf("unexpected session after replay: %+v", loaded)
	}
}

