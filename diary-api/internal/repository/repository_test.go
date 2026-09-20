package repository_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/auth"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/db"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
	"github.com/google/uuid"
)

func openRepo(t *testing.T) (*repository.Repository, string) {
	t.Helper()
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repo := repository.New(sqlDB)
	hash, err := auth.HashPassword("testpass12")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	user, err := repo.CreateUser(uuid.NewString()+"@test.local", hash)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return repo, user.ID
}

func TestExerciseTargetAndWorkoutSession(t *testing.T) {
	repo, uid := openRepo(t)

	ex, err := repo.CreateExercise(uid, "Wide grip pull-up", "back", "reps", true, "")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	proto, err := repo.CreateIntervalProtocol(uid, repository.CreateIntervalProtocolInput{
		Name:        "40/20 + warmup",
		PrepareSec:  5,
		WorkSec:     40,
		RestSec:     20,
		WarmupExtra: true,
	})
	if err != nil {
		t.Fatalf("create protocol: %v", err)
	}

	target, err := repo.SetTarget(uid, ex.ID, 3, 12, 0, 0, 25)
	if err != nil {
		t.Fatalf("set target: %v", err)
	}
	if target.AssistKg != 25 || target.TargetSets != 3 {
		t.Fatalf("unexpected target: %+v", target)
	}

	got, err := repo.GetExercise(uid, ex.ID)
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if got.Target == nil || got.Target.AssistKg != 25 {
		t.Fatalf("expected target on exercise, got %+v", got)
	}
	if got.Kind != "reps" {
		t.Fatalf("expected kind reps, got %+v", got)
	}

	updated, err := repo.UpdateExercise(uid, ex.ID, "Pull-up", "back", "reps", false, proto.ID)
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

	list, err := repo.ListExercises(uid)
	if err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	if len(list) != 1 || list[0].Target == nil || list[0].Target.AssistKg != 0 {
		t.Fatalf("expected cleared assist on listed exercise, got %+v", list)
	}

	session, _, err := repo.StartWorkoutSession(uid, uuid.NewString(), "2026-07-16T18:00:00Z", false, "")
	if err != nil {
		t.Fatalf("start workout session: %v", err)
	}

	block, err := repo.SaveSessionExercise(uid, session.ID, 1, ex.ID, []repository.CreateSetInput{
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

	block2, err := repo.SaveSessionExercise(uid, session.ID, 1, ex.ID, []repository.CreateSetInput{
		{SetNumber: 1, Reps: 10, WeightKg: 0, AssistKg: 20},
	})
	if err != nil {
		t.Fatalf("update session exercise: %v", err)
	}
	if len(block2.Sets) != 1 || block2.Sets[0].Reps != 10 {
		t.Fatalf("unexpected updated block: %+v", block2)
	}

	summaries, err := repo.ListWorkoutSessions(uid)
	if err != nil {
		t.Fatalf("list workout sessions: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 session in list, got %+v", summaries)
	}
	if summaries[0].ExerciseCount != 1 || len(summaries[0].ExerciseNames) != 1 {
		t.Fatalf("unexpected summary: %+v", summaries[0])
	}

	finished, err := repo.FinishWorkoutSession(uid, session.ID, 3725, "")
	if err != nil {
		t.Fatalf("finish workout session: %v", err)
	}
	if finished.DurationSec != 3725 || finished.StartedAt == "" {
		t.Fatalf("unexpected finished session: %+v", finished)
	}

	loaded, err := repo.GetWorkoutSession(uid, session.ID)
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
	if loaded.Exercises[0].Target == nil || loaded.Exercises[0].Target.TargetSets != 3 || loaded.Exercises[0].Target.TargetReps != 12 {
		t.Fatalf("expected target snapshot on session exercise, got %+v", loaded.Exercises[0].Target)
	}

	fullSession, err := repo.CreateWorkoutSession(uid, repository.CreateWorkoutSessionInput{
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
	repo, uid := openRepo(t)

	ex, err := repo.CreateExercise(uid, "Squat", "legs", "reps", false, "")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	planA, err := repo.CreateWorkoutPlan(uid, repository.CreateWorkoutPlanInput{
		Name: "A", ExerciseIDs: []string{ex.ID},
	})
	if err != nil {
		t.Fatalf("create plan A: %v", err)
	}
	planB, err := repo.CreateWorkoutPlan(uid, repository.CreateWorkoutPlanInput{
		Name: "B", ExerciseIDs: []string{ex.ID},
	})
	if err != nil {
		t.Fatalf("create plan B: %v", err)
	}
	planC, err := repo.CreateWorkoutPlan(uid, repository.CreateWorkoutPlanInput{
		Name: "C", ExerciseIDs: []string{ex.ID},
	})
	if err != nil {
		t.Fatalf("create plan C: %v", err)
	}

	home, err := repo.CreateCycle(uid, "Дом")
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	outdoor, err := repo.CreateCycle(uid, "Улица")
	if err != nil {
		t.Fatalf("create outdoor cycle: %v", err)
	}
	if home.ID == outdoor.ID {
		t.Fatalf("expected distinct cycle ids")
	}

	list, err := repo.ListCycles(uid)
	if err != nil || len(list) != 2 {
		t.Fatalf("list cycles: %v %+v", err, list)
	}

	empty, err := repo.GetCycle(uid, home.ID)
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	if len(empty.Steps) != 0 || empty.CurrentStep != 1 || empty.Name != "Дом" {
		t.Fatalf("unexpected empty cycle: %+v", empty)
	}

	if empty.OnHome {
		t.Fatalf("new cycle should not be on home: %+v", empty)
	}

	cycle, err := repo.ReplaceCycleSteps(uid, home.ID, []string{planA.ID, planB.ID, planC.ID})
	if err != nil {
		t.Fatalf("replace steps: %v", err)
	}
	if len(cycle.Steps) != 3 || cycle.CurrentStep != 1 {
		t.Fatalf("unexpected cycle after replace: %+v", cycle)
	}
	if cycle.Steps[0].WorkoutPlanName != "A" || cycle.Steps[2].Position != 3 {
		t.Fatalf("unexpected steps: %+v", cycle.Steps)
	}

	cycle, err = repo.SetCycleOnHome(uid, home.ID, true)
	if err != nil || !cycle.OnHome {
		t.Fatalf("set on home: %+v %v", cycle, err)
	}
	cycle, err = repo.SetCycleOnHome(uid, home.ID, false)
	if err != nil || cycle.OnHome {
		t.Fatalf("unset on home: %+v %v", cycle, err)
	}

	cycle, err = repo.AdvanceCycle(uid, home.ID, "")
	if err != nil {
		t.Fatalf("advance: %v", err)
	}
	if cycle.CurrentStep != 2 {
		t.Fatalf("expected step 2, got %d", cycle.CurrentStep)
	}
	if !cycle.OnHome {
		t.Fatalf("advance should pin cycle to home: %+v", cycle)
	}

	session, _, err := repo.StartWorkoutSession(uid, uuid.NewString(), "2026-07-20T10:00:00Z", false, planB.ID)
	if err != nil {
		t.Fatalf("start session for link: %v", err)
	}
	// Finish current cycle step (plan B is step 2) — advances in the same call
	finished, err := repo.FinishWorkoutSession(uid, session.ID, 600, home.ID)
	if err != nil {
		t.Fatalf("finish+advance: %v", err)
	}
	if finished.CycleID != home.ID || finished.CycleStep != 2 {
		t.Fatalf("expected cycle link on finished session, got %+v", finished)
	}
	cycle, err = repo.GetCycle(uid, home.ID)
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
	cycle, err = repo.AdvanceCycle(uid, home.ID, "")
	if err != nil {
		t.Fatalf("advance at end: %v", err)
	}
	if !cycle.Completed || cycle.CurrentStep != 4 {
		t.Fatalf("expected completed after last step, got %+v", cycle)
	}
	if cycle.OnHome {
		t.Fatalf("completed cycle should leave home: %+v", cycle)
	}

	repeated, err := repo.RepeatCycle(uid, home.ID)
	if err != nil {
		t.Fatalf("repeat: %v", err)
	}
	if repeated.ID == home.ID || repeated.Completed || repeated.CurrentStep != 1 || !repeated.OnHome {
		t.Fatalf("unexpected repeated cycle: %+v", repeated)
	}
	if len(repeated.Steps) != 3 || repeated.Name != "Дом" {
		t.Fatalf("repeated steps/name: %+v", repeated)
	}
	still, err := repo.GetCycle(uid, home.ID)
	if err != nil || !still.Completed {
		t.Fatalf("original should stay completed: %+v %v", still, err)
	}

	// Shrink a completed clone for cursor normalize check
	_, _ = repo.AdvanceCycle(uid, repeated.ID, "")
	_, _ = repo.AdvanceCycle(uid, repeated.ID, "")
	cycle, err = repo.AdvanceCycle(uid, repeated.ID, "")
	if err != nil || !cycle.Completed {
		t.Fatalf("re-complete clone: %+v %v", cycle, err)
	}

	// Outdoor cycle stays independent
	out, err := repo.GetCycle(uid, outdoor.ID)
	if err != nil || out.CurrentStep != 1 || len(out.Steps) != 0 || out.Completed {
		t.Fatalf("outdoor should be untouched: %+v %v", out, err)
	}

	cycle, err = repo.ReplaceCycleSteps(uid, repeated.ID, []string{planB.ID})
	if err != nil {
		t.Fatalf("shrink steps: %v", err)
	}
	// Was completed (step 4); after shrink to 1 step → stay completed at step 2
	if !cycle.Completed || cycle.CurrentStep != 2 || len(cycle.Steps) != 1 {
		t.Fatalf("unexpected after shrink: %+v", cycle)
	}

	// Original completed history untouched
	still, err = repo.GetCycle(uid, home.ID)
	if err != nil || !still.Completed || len(still.Steps) != 3 {
		t.Fatalf("original history changed: %+v %v", still, err)
	}

	session2, _, err := repo.StartWorkoutSession(uid, uuid.NewString(), "2026-07-20T11:00:00Z", false, planB.ID)
	if err != nil {
		t.Fatalf("start with plan: %v", err)
	}
	loaded, err := repo.GetWorkoutSession(uid, session2.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if loaded.WorkoutPlanID != planB.ID {
		t.Fatalf("expected workoutPlanId %s, got %q", planB.ID, loaded.WorkoutPlanID)
	}
}

func TestHoldExerciseTargetAndSets(t *testing.T) {
	repo, uid := openRepo(t)

	ex, err := repo.CreateExercise(uid, "Bridge hold", "core", "hold", true, "")
	if err != nil {
		t.Fatalf("create hold exercise: %v", err)
	}
	if ex.Kind != "hold" || ex.SupportsAssist {
		t.Fatalf("hold should force assist off: %+v", ex)
	}

	target, err := repo.SetTarget(uid, ex.ID, 3, 12, 60, 10, 5)
	if err != nil {
		t.Fatalf("set hold target: %v", err)
	}
	if target.HoldSec != 60 || target.TargetReps != 0 || target.WeightKg != 10 || target.AssistKg != 0 {
		t.Fatalf("unexpected hold target: %+v", target)
	}

	session, _, err := repo.StartWorkoutSession(uid, uuid.NewString(), "2026-07-21T12:00:00Z", false, "")
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	block, err := repo.SaveSessionExercise(uid, session.ID, 1, ex.ID, []repository.CreateSetInput{
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

	loaded, err := repo.GetWorkoutSession(uid, session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if loaded.Exercises[0].Kind != "hold" {
		t.Fatalf("expected hold kind on block: %+v", loaded.Exercises[0])
	}
	if loaded.Exercises[0].Target == nil || loaded.Exercises[0].Target.HoldSec != 60 {
		t.Fatalf("expected hold target snapshot, got %+v", loaded.Exercises[0].Target)
	}
	sum := 0
	for _, s := range loaded.Exercises[0].Sets {
		sum += s.DurationSec
	}
	if sum != 155 {
		t.Fatalf("expected total hold 155s, got %d", sum)
	}
}

func TestSessionExerciseTargetSnapshotStableAfterCatalogChange(t *testing.T) {
	repo, uid := openRepo(t)

	ex, err := repo.CreateExercise(uid, "Squat", "legs", "reps", false, "")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}
	if _, err := repo.SetTarget(uid, ex.ID, 3, 8, 0, 100, 0); err != nil {
		t.Fatalf("set target: %v", err)
	}

	session, _, err := repo.StartWorkoutSession(uid, uuid.NewString(), "2026-07-25T10:00:00Z", false, "")
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := repo.SaveSessionExercise(uid, session.ID, 1, ex.ID, []repository.CreateSetInput{
		{SetNumber: 1, Reps: 8, WeightKg: 100},
	}); err != nil {
		t.Fatalf("save session exercise: %v", err)
	}

	if _, err := repo.SetTarget(uid, ex.ID, 4, 10, 0, 120, 0); err != nil {
		t.Fatalf("change catalog target: %v", err)
	}

	loaded, err := repo.GetWorkoutSession(uid, session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	got := loaded.Exercises[0].Target
	if got == nil {
		t.Fatal("expected snapshotted target on session exercise")
	}
	if got.TargetSets != 3 || got.TargetReps != 8 || got.WeightKg != 100 {
		t.Fatalf("snapshot should keep save-time goal 3×8 @100, got %+v", got)
	}

	catalog, err := repo.GetExercise(uid, ex.ID)
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if catalog.Target == nil || catalog.Target.TargetSets != 4 || catalog.Target.TargetReps != 10 {
		t.Fatalf("catalog target should be updated, got %+v", catalog.Target)
	}
}

func TestStartWorkoutSessionClientIDRequiredAndIdempotent(t *testing.T) {
	repo, uid := openRepo(t)

	if _, _, err := repo.StartWorkoutSession(uid, "", "2026-07-22T12:00:00Z", false, ""); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, _, err := repo.StartWorkoutSession(uid, "not-a-uuid", "2026-07-22T12:00:00Z", false, ""); err == nil {
		t.Fatal("expected error for invalid uuid")
	}

	id := uuid.NewString()
	first, created, err := repo.StartWorkoutSession(uid, id, "2026-07-22T12:00:00Z", false, "")
	if err != nil || !created {
		t.Fatalf("first start: created=%v err=%v", created, err)
	}
	if first.ID != id {
		t.Fatalf("expected client id %s, got %s", id, first.ID)
	}

	second, createdAgain, err := repo.StartWorkoutSession(uid, id, "2026-07-22T13:00:00Z", true, "")
	if err != nil || createdAgain {
		t.Fatalf("replay start: created=%v err=%v", createdAgain, err)
	}
	if second.ID != id {
		t.Fatalf("replay id mismatch: %s", second.ID)
	}
	if second.IsDeload != first.IsDeload {
		t.Fatalf("idempotent replay should return original row, got isDeload=%v", second.IsDeload)
	}

	loaded, err := repo.GetWorkoutSession(uid, id)
	if err != nil {
		t.Fatalf("get after replay: %v", err)
	}
	if loaded.ID != id || loaded.DurationSec != 0 {
		t.Fatalf("unexpected session after replay: %+v", loaded)
	}
}

func TestExerciseStatsVolumeAndTarget(t *testing.T) {
	repo, uid := openRepo(t)

	ex, err := repo.CreateExercise(uid, "Pull-up", "back", "reps", false, "")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}
	if _, err := repo.SetTarget(uid, ex.ID, 3, 10, 0, 0, 0); err != nil {
		t.Fatalf("set target: %v", err)
	}

	s1, err := repo.CreateWorkoutSession(uid, repository.CreateWorkoutSessionInput{
		PerformedAt: "2026-07-20T10:00:00Z",
		Exercises: []repository.CreateWorkoutSessionExerciseInput{
			{
				ExerciseID: ex.ID,
				Sets: []repository.CreateSetInput{
					{SetNumber: 1, Reps: 8},
					{SetNumber: 2, Reps: 8},
					{SetNumber: 3, Reps: 8},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create session 1: %v", err)
	}

	if _, err := repo.SetTarget(uid, ex.ID, 4, 12, 0, 0, 0); err != nil {
		t.Fatalf("raise target: %v", err)
	}

	s2, err := repo.CreateWorkoutSession(uid, repository.CreateWorkoutSessionInput{
		PerformedAt: "2026-07-25T10:00:00Z",
		Exercises: []repository.CreateWorkoutSessionExerciseInput{
			{
				ExerciseID: ex.ID,
				Sets: []repository.CreateSetInput{
					{SetNumber: 1, Reps: 12},
					{SetNumber: 2, Reps: 12},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create session 2: %v", err)
	}

	stats, err := repo.GetExerciseStats(uid, ex.ID, "all")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Kind != "reps" || stats.Period != "all" {
		t.Fatalf("unexpected meta: %+v", stats)
	}
	if stats.CurrentTargetVolume == nil || *stats.CurrentTargetVolume != 48 {
		t.Fatalf("expected current target 4×12=48, got %+v", stats.CurrentTargetVolume)
	}
	if len(stats.Points) != 2 {
		t.Fatalf("expected 2 points, got %+v", stats.Points)
	}
	if stats.Points[0].SessionID != s1.ID || stats.Points[0].Volume != 24 || stats.Points[0].SetsCount != 3 {
		t.Fatalf("unexpected point 0: %+v", stats.Points[0])
	}
	if stats.Points[0].TargetVolume == nil || *stats.Points[0].TargetVolume != 30 {
		t.Fatalf("expected snapshot target 3×10=30, got %+v", stats.Points[0].TargetVolume)
	}
	if stats.Points[1].SessionID != s2.ID || stats.Points[1].Volume != 24 || stats.Points[1].SetsCount != 2 {
		t.Fatalf("unexpected point 1: %+v", stats.Points[1])
	}
	if stats.Points[1].TargetVolume == nil || *stats.Points[1].TargetVolume != 48 {
		t.Fatalf("expected snapshot target 4×12=48, got %+v", stats.Points[1].TargetVolume)
	}
	if stats.Summary.SessionCount != 2 || stats.Summary.TotalVolume != 48 || stats.Summary.AvgSets != 2.5 {
		t.Fatalf("unexpected summary: %+v", stats.Summary)
	}

	if _, err := repo.GetExerciseStats(uid, ex.ID, "week"); !errors.Is(err, repository.ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}

func TestUserIsolation(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "iso.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repo := repository.New(sqlDB)

	hash, err := auth.HashPassword("testpass12")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	a, err := repo.CreateUser("a@test.local", hash)
	if err != nil {
		t.Fatalf("user a: %v", err)
	}
	b, err := repo.CreateUser("b@test.local", hash)
	if err != nil {
		t.Fatalf("user b: %v", err)
	}

	exA, err := repo.CreateExercise(a.ID, "A only", "back", "reps", false, "")
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	exB, err := repo.CreateExercise(b.ID, "B only", "legs", "reps", false, "")
	if err != nil {
		t.Fatalf("create b: %v", err)
	}

	listA, err := repo.ListExercises(a.ID)
	if err != nil || len(listA) != 1 || listA[0].ID != exA.ID {
		t.Fatalf("list A: %+v err=%v", listA, err)
	}
	if _, err := repo.GetExercise(a.ID, exB.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("A should not see B exercise, got %v", err)
	}
	if _, err := repo.GetExercise(b.ID, exA.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("B should not see A exercise, got %v", err)
	}

	if err := repo.UpdateUserPasswordHash(a.ID, hash); err != nil {
		t.Fatalf("update password: %v", err)
	}
}


