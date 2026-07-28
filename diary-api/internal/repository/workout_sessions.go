package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
	"github.com/google/uuid"
)

// CreateWorkoutSessionInput is the payload to persist one gym visit.
type CreateWorkoutSessionInput struct {
	PerformedAt string
	IsDeload    bool
	Exercises   []CreateWorkoutSessionExerciseInput
}

// CreateWorkoutSessionExerciseInput is one exercise block in the session.
type CreateWorkoutSessionExerciseInput struct {
	ExerciseID string
	Sets       []CreateSetInput
}

// CreateSetInput is one performed set.
type CreateSetInput struct {
	SetNumber   int
	Reps        int
	DurationSec int
	WeightKg    float64
	AssistKg    float64
}

// CreateWorkoutSession stores a workout session with exercises and sets in one transaction.
func (r *Repository) CreateWorkoutSession(in CreateWorkoutSessionInput) (models.WorkoutSession, error) {
	if len(in.Exercises) == 0 {
		return models.WorkoutSession{}, fmt.Errorf("workout session requires at least one exercise")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	session := models.WorkoutSession{
		ID:          uuid.NewString(),
		PerformedAt: in.PerformedAt,
		StartedAt:   in.PerformedAt,
		DurationSec: 0,
		IsDeload:    in.IsDeload,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Exercises:   make([]models.WorkoutSessionExercise, 0, len(in.Exercises)),
	}
	if session.PerformedAt == "" {
		session.PerformedAt = session.CreatedAt
		session.StartedAt = session.CreatedAt
	}

	_, err = tx.Exec(
		`INSERT INTO workout_sessions (id, performed_at, is_deload, created_at, started_at, duration_sec)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID, session.PerformedAt, boolToInt(session.IsDeload), session.CreatedAt,
		session.StartedAt, session.DurationSec,
	)
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("insert workout session: %w", err)
	}

	for i, exIn := range in.Exercises {
		if exIn.ExerciseID == "" {
			return models.WorkoutSession{}, fmt.Errorf("exercise at position %d: empty exerciseId", i+1)
		}
		if len(exIn.Sets) == 0 {
			return models.WorkoutSession{}, fmt.Errorf("exercise %s: at least one set required", exIn.ExerciseID)
		}

		var exists int
		err := tx.QueryRow(`SELECT 1 FROM exercises WHERE id = ?`, exIn.ExerciseID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return models.WorkoutSession{}, fmt.Errorf("exercise %s: %w", exIn.ExerciseID, ErrNotFound)
		}
		if err != nil {
			return models.WorkoutSession{}, fmt.Errorf("check exercise %s: %w", exIn.ExerciseID, err)
		}

		target, err := loadExerciseTargetTx(tx, exIn.ExerciseID)
		if err != nil {
			return models.WorkoutSession{}, fmt.Errorf("load target for %s: %w", exIn.ExerciseID, err)
		}

		wse := models.WorkoutSessionExercise{
			ID:               uuid.NewString(),
			WorkoutSessionID: session.ID,
			ExerciseID:       exIn.ExerciseID,
			Position:         i + 1,
			Target:           target,
			Sets:             make([]models.Set, 0, len(exIn.Sets)),
		}

		snapSets, snapReps, snapHold, snapWeight, snapAssist := targetSnapshotArgs(target)
		_, err = tx.Exec(
			`INSERT INTO workout_session_exercises (
			   id, workout_session_id, exercise_id, position,
			   target_sets, target_reps, target_hold_sec, target_weight_kg, target_assist_kg
			 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			wse.ID, wse.WorkoutSessionID, wse.ExerciseID, wse.Position,
			snapSets, snapReps, snapHold, snapWeight, snapAssist,
		)
		if err != nil {
			return models.WorkoutSession{}, fmt.Errorf("insert workout session exercise: %w", err)
		}

		for _, setIn := range exIn.Sets {
			set := models.Set{
				ID:                       uuid.NewString(),
				WorkoutSessionExerciseID: wse.ID,
				SetNumber:                setIn.SetNumber,
				Reps:                     setIn.Reps,
				DurationSec:              setIn.DurationSec,
				WeightKg:                 setIn.WeightKg,
				AssistKg:                 setIn.AssistKg,
			}
			_, err = tx.Exec(
				`INSERT INTO sets (id, workout_session_exercise_id, set_number, reps, weight_kg, assist_kg, duration_sec)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				set.ID, set.WorkoutSessionExerciseID, set.SetNumber, set.Reps, set.WeightKg, set.AssistKg, set.DurationSec,
			)
			if err != nil {
				return models.WorkoutSession{}, fmt.Errorf("insert set: %w", err)
			}
			wse.Sets = append(wse.Sets, set)
		}

		session.Exercises = append(session.Exercises, wse)
	}

	if err := tx.Commit(); err != nil {
		return models.WorkoutSession{}, fmt.Errorf("commit tx: %w", err)
	}
	return session, nil
}

// StartWorkoutSession creates an empty session for incremental logging.
// id is required (client-generated UUID). Replaying the same id returns the existing row.
// created is false when the session already existed (idempotent replay).
func (r *Repository) StartWorkoutSession(id, performedAt string, isDeload bool, workoutPlanID string) (models.WorkoutSession, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return models.WorkoutSession{}, false, fmt.Errorf("id is required")
	}
	if _, err := uuid.Parse(id); err != nil {
		return models.WorkoutSession{}, false, fmt.Errorf("id must be a valid UUID")
	}

	if existing, err := r.GetWorkoutSession(id); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, ErrNotFound) {
		return models.WorkoutSession{}, false, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	session := models.WorkoutSession{
		ID:            id,
		PerformedAt:   performedAt,
		StartedAt:     now,
		DurationSec:   0,
		IsDeload:      isDeload,
		WorkoutPlanID: workoutPlanID,
		CreatedAt:     now,
		Exercises:     []models.WorkoutSessionExercise{},
	}
	if session.PerformedAt == "" {
		session.PerformedAt = now
	}

	if workoutPlanID != "" {
		var exists int
		err := r.db.QueryRow(`SELECT 1 FROM workout_plans WHERE id = ?`, workoutPlanID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return models.WorkoutSession{}, false, fmt.Errorf("workout plan %s: %w", workoutPlanID, ErrNotFound)
		}
		if err != nil {
			return models.WorkoutSession{}, false, fmt.Errorf("check workout plan: %w", err)
		}
	}

	var planArg any
	if workoutPlanID != "" {
		planArg = workoutPlanID
	}

	_, err := r.db.Exec(
		`INSERT INTO workout_sessions (id, performed_at, is_deload, created_at, started_at, duration_sec, workout_plan_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.PerformedAt, boolToInt(session.IsDeload), session.CreatedAt,
		session.StartedAt, session.DurationSec, planArg,
	)
	if err != nil {
		return models.WorkoutSession{}, false, fmt.Errorf("insert workout session: %w", err)
	}
	return session, true, nil
}

// FinishWorkoutSession stores total duration for a session.
// If cycleID is set and the session matches that cycle's current step, advances the cycle
// in the same request (links the session to the completed step).
func (r *Repository) FinishWorkoutSession(sessionID string, durationSec int, cycleID string) (models.WorkoutSession, error) {
	if durationSec < 0 {
		return models.WorkoutSession{}, fmt.Errorf("durationSec must be >= 0")
	}

	res, err := r.db.Exec(
		`UPDATE workout_sessions SET duration_sec = ? WHERE id = ?`,
		durationSec, sessionID,
	)
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("finish workout session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("finish workout session rows: %w", err)
	}
	if n == 0 {
		return models.WorkoutSession{}, ErrNotFound
	}

	cycleID = strings.TrimSpace(cycleID)
	if cycleID != "" {
		session, err := r.GetWorkoutSession(sessionID)
		if err != nil {
			return models.WorkoutSession{}, err
		}
		cycle, err := r.GetCycle(cycleID)
		if err != nil {
			return models.WorkoutSession{}, fmt.Errorf("cycle: %w", err)
		}
		if !cycle.Completed && len(cycle.Steps) > 0 {
			step, ok := cycleStepAt(cycle, cycle.CurrentStep)
			if ok && session.WorkoutPlanID != "" && step.WorkoutPlanID == session.WorkoutPlanID {
				if _, err := r.AdvanceCycle(cycleID, sessionID); err != nil {
					return models.WorkoutSession{}, fmt.Errorf("advance cycle after finish: %w", err)
				}
			}
		}
	}

	return r.GetWorkoutSession(sessionID)
}

func cycleStepAt(c models.Cycle, position int) (models.CycleStep, bool) {
	for _, s := range c.Steps {
		if s.Position == position {
			return s, true
		}
	}
	return models.CycleStep{}, false
}

// SaveSessionExercise upserts one exercise block inside an existing session.
func (r *Repository) SaveSessionExercise(
	sessionID string,
	position int,
	exerciseID string,
	sets []CreateSetInput,
) (models.WorkoutSessionExercise, error) {
	if position < 1 {
		return models.WorkoutSessionExercise{}, fmt.Errorf("position must be >= 1")
	}
	if exerciseID == "" {
		return models.WorkoutSessionExercise{}, fmt.Errorf("exerciseId is required")
	}
	if len(sets) == 0 {
		return models.WorkoutSessionExercise{}, fmt.Errorf("at least one set required")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var sessionExists int
	err = tx.QueryRow(`SELECT 1 FROM workout_sessions WHERE id = ?`, sessionID).Scan(&sessionExists)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WorkoutSessionExercise{}, ErrNotFound
	}
	if err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("check session %s: %w", sessionID, err)
	}

	var exerciseExists int
	err = tx.QueryRow(`SELECT 1 FROM exercises WHERE id = ?`, exerciseID).Scan(&exerciseExists)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WorkoutSessionExercise{}, fmt.Errorf("exercise %s: %w", exerciseID, ErrNotFound)
	}
	if err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("check exercise %s: %w", exerciseID, err)
	}

	target, err := loadExerciseTargetTx(tx, exerciseID)
	if err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("load target for %s: %w", exerciseID, err)
	}
	snapSets, snapReps, snapHold, snapWeight, snapAssist := targetSnapshotArgs(target)

	var wse models.WorkoutSessionExercise
	err = tx.QueryRow(
		`SELECT id, workout_session_id, exercise_id, position
		 FROM workout_session_exercises
		 WHERE workout_session_id = ? AND exercise_id = ?`,
		sessionID, exerciseID,
	).Scan(&wse.ID, &wse.WorkoutSessionID, &wse.ExerciseID, &wse.Position)
	if errors.Is(err, sql.ErrNoRows) {
		wse = models.WorkoutSessionExercise{
			ID:               uuid.NewString(),
			WorkoutSessionID: sessionID,
			ExerciseID:       exerciseID,
			Position:         position,
			Target:           target,
		}
		_, err = tx.Exec(
			`INSERT INTO workout_session_exercises (
			   id, workout_session_id, exercise_id, position,
			   target_sets, target_reps, target_hold_sec, target_weight_kg, target_assist_kg
			 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			wse.ID, wse.WorkoutSessionID, wse.ExerciseID, wse.Position,
			snapSets, snapReps, snapHold, snapWeight, snapAssist,
		)
		if err != nil {
			return models.WorkoutSessionExercise{}, fmt.Errorf("insert workout session exercise: %w", err)
		}
	} else if err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("find workout session exercise: %w", err)
	} else {
		_, err = tx.Exec(
			`UPDATE workout_session_exercises
			 SET position = ?,
			     target_sets = ?, target_reps = ?, target_hold_sec = ?,
			     target_weight_kg = ?, target_assist_kg = ?
			 WHERE id = ?`,
			position, snapSets, snapReps, snapHold, snapWeight, snapAssist, wse.ID,
		)
		if err != nil {
			return models.WorkoutSessionExercise{}, fmt.Errorf("update session exercise snapshot: %w", err)
		}
		wse.Position = position
		wse.Target = target
	}

	if _, err := tx.Exec(`DELETE FROM sets WHERE workout_session_exercise_id = ?`, wse.ID); err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("delete old sets: %w", err)
	}

	wse.Sets = make([]models.Set, 0, len(sets))
	for _, setIn := range sets {
		set := models.Set{
			ID:                       uuid.NewString(),
			WorkoutSessionExerciseID: wse.ID,
			SetNumber:                setIn.SetNumber,
			Reps:                     setIn.Reps,
			DurationSec:              setIn.DurationSec,
			WeightKg:                 setIn.WeightKg,
			AssistKg:                 setIn.AssistKg,
		}
		_, err = tx.Exec(
			`INSERT INTO sets (id, workout_session_exercise_id, set_number, reps, weight_kg, assist_kg, duration_sec)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			set.ID, set.WorkoutSessionExerciseID, set.SetNumber, set.Reps, set.WeightKg, set.AssistKg, set.DurationSec,
		)
		if err != nil {
			return models.WorkoutSessionExercise{}, fmt.Errorf("insert set: %w", err)
		}
		wse.Sets = append(wse.Sets, set)
	}

	if err := tx.Commit(); err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("commit tx: %w", err)
	}
	return wse, nil
}

// ListWorkoutSessions returns completed sessions (with at least one exercise), newest first.
func (r *Repository) ListWorkoutSessions() ([]models.WorkoutSessionSummary, error) {
	rows, err := r.db.Query(
		`SELECT ws.id, ws.performed_at, ws.is_deload, ws.created_at, ws.duration_sec,
		        COUNT(wse.id) AS exercise_count,
		        GROUP_CONCAT(e.name, char(31) ORDER BY wse.position) AS exercise_names,
		        MAX(c.id) AS cycle_id,
		        MAX(c.name) AS cycle_name,
		        MAX(cs.position) AS cycle_step
		 FROM workout_sessions ws
		 INNER JOIN workout_session_exercises wse ON wse.workout_session_id = ws.id
		 INNER JOIN exercises e ON e.id = wse.exercise_id
		 LEFT JOIN cycle_steps cs ON cs.completed_session_id = ws.id
		 LEFT JOIN cycles c ON c.id = cs.cycle_id
		 GROUP BY ws.id
		 ORDER BY ws.performed_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list workout sessions: %w", err)
	}
	defer rows.Close()

	var out []models.WorkoutSessionSummary
	for rows.Next() {
		var s models.WorkoutSessionSummary
		var deload int
		var namesRaw sql.NullString
		var cycleID, cycleName sql.NullString
		var cycleStep sql.NullInt64
		if err := rows.Scan(
			&s.ID, &s.PerformedAt, &deload, &s.CreatedAt, &s.DurationSec,
			&s.ExerciseCount, &namesRaw, &cycleID, &cycleName, &cycleStep,
		); err != nil {
			return nil, fmt.Errorf("scan workout session summary: %w", err)
		}
		s.IsDeload = deload == 1
		if namesRaw.Valid && namesRaw.String != "" {
			s.ExerciseNames = strings.Split(namesRaw.String, "\x1f")
		} else {
			s.ExerciseNames = []string{}
		}
		if cycleID.Valid {
			s.CycleID = cycleID.String
		}
		if cycleName.Valid {
			s.CycleName = cycleName.String
		}
		if cycleStep.Valid {
			s.CycleStep = int(cycleStep.Int64)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list workout sessions rows: %w", err)
	}
	if out == nil {
		out = []models.WorkoutSessionSummary{}
	}
	return out, nil
}

// GetWorkoutSession returns a workout session with exercises and sets.
func (r *Repository) GetWorkoutSession(id string) (models.WorkoutSession, error) {
	var session models.WorkoutSession
	var deload int
	var startedAt sql.NullString
	var planID sql.NullString
	var cycleID, cycleName sql.NullString
	var cycleStep sql.NullInt64
	err := r.db.QueryRow(
		`SELECT ws.id, ws.performed_at, ws.is_deload, ws.created_at, ws.started_at, ws.duration_sec, ws.workout_plan_id,
		        c.id, c.name, cs.position
		 FROM workout_sessions ws
		 LEFT JOIN cycle_steps cs ON cs.completed_session_id = ws.id
		 LEFT JOIN cycles c ON c.id = cs.cycle_id
		 WHERE ws.id = ?`,
		id,
	).Scan(
		&session.ID, &session.PerformedAt, &deload, &session.CreatedAt,
		&startedAt, &session.DurationSec, &planID,
		&cycleID, &cycleName, &cycleStep,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WorkoutSession{}, ErrNotFound
	}
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("get workout session: %w", err)
	}
	session.IsDeload = deload == 1
	if startedAt.Valid {
		session.StartedAt = startedAt.String
	}
	if planID.Valid {
		session.WorkoutPlanID = planID.String
	}
	if cycleID.Valid {
		session.CycleID = cycleID.String
	}
	if cycleName.Valid {
		session.CycleName = cycleName.String
	}
	if cycleStep.Valid {
		session.CycleStep = int(cycleStep.Int64)
	}

	exRows, err := r.db.Query(
		`SELECT wse.id, wse.workout_session_id, wse.exercise_id, wse.position,
		        e.name, e.kind, e.supports_assist,
		        wse.target_sets, wse.target_reps, wse.target_hold_sec,
		        wse.target_weight_kg, wse.target_assist_kg
		 FROM workout_session_exercises wse
		 INNER JOIN exercises e ON e.id = wse.exercise_id
		 WHERE wse.workout_session_id = ?
		 ORDER BY wse.position ASC`,
		id,
	)
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("list workout session exercises: %w", err)
	}
	defer exRows.Close()

	for exRows.Next() {
		var wse models.WorkoutSessionExercise
		var assist int
		var targetSets, targetReps, targetHoldSec sql.NullInt64
		var targetWeightKg, targetAssistKg sql.NullFloat64
		if err := exRows.Scan(
			&wse.ID, &wse.WorkoutSessionID, &wse.ExerciseID, &wse.Position,
			&wse.ExerciseName, &wse.Kind, &assist,
			&targetSets, &targetReps, &targetHoldSec, &targetWeightKg, &targetAssistKg,
		); err != nil {
			return models.WorkoutSession{}, fmt.Errorf("scan workout session exercise: %w", err)
		}
		wse.Kind = normalizeExerciseKind(wse.Kind)
		wse.SupportsAssist = assist == 1
		if targetSets.Valid {
			wse.Target = &models.ExerciseTarget{
				ExerciseID: wse.ExerciseID,
				TargetSets: int(targetSets.Int64),
				TargetReps: int(targetReps.Int64),
				HoldSec:    int(targetHoldSec.Int64),
				WeightKg:   targetWeightKg.Float64,
				AssistKg:   targetAssistKg.Float64,
			}
		}

		sets, err := r.listSets(wse.ID)
		if err != nil {
			return models.WorkoutSession{}, err
		}
		wse.Sets = sets
		session.Exercises = append(session.Exercises, wse)
	}
	if err := exRows.Err(); err != nil {
		return models.WorkoutSession{}, fmt.Errorf("workout session exercises rows: %w", err)
	}
	if session.Exercises == nil {
		session.Exercises = []models.WorkoutSessionExercise{}
	}
	return session, nil
}

func (r *Repository) listSets(workoutSessionExerciseID string) ([]models.Set, error) {
	rows, err := r.db.Query(
		`SELECT id, workout_session_exercise_id, set_number, reps, weight_kg, assist_kg, duration_sec
		 FROM sets
		 WHERE workout_session_exercise_id = ?
		 ORDER BY set_number ASC`,
		workoutSessionExerciseID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sets: %w", err)
	}
	defer rows.Close()

	var out []models.Set
	for rows.Next() {
		var s models.Set
		if err := rows.Scan(
			&s.ID, &s.WorkoutSessionExerciseID, &s.SetNumber, &s.Reps, &s.WeightKg, &s.AssistKg, &s.DurationSec,
		); err != nil {
			return nil, fmt.Errorf("scan set: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sets rows: %w", err)
	}
	if out == nil {
		out = []models.Set{}
	}
	return out, nil
}

// loadExerciseTargetTx returns the current catalog target for an exercise, or nil if unset.
func loadExerciseTargetTx(tx *sql.Tx, exerciseID string) (*models.ExerciseTarget, error) {
	var t models.ExerciseTarget
	err := tx.QueryRow(
		`SELECT exercise_id, target_sets, target_reps, hold_sec, weight_kg, assist_kg
		 FROM exercise_targets WHERE exercise_id = ?`,
		exerciseID,
	).Scan(&t.ExerciseID, &t.TargetSets, &t.TargetReps, &t.HoldSec, &t.WeightKg, &t.AssistKg)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// targetSnapshotArgs maps a target pointer to nullable INSERT/UPDATE args.
func targetSnapshotArgs(t *models.ExerciseTarget) (sets, reps, holdSec any, weightKg, assistKg any) {
	if t == nil {
		return nil, nil, nil, nil, nil
	}
	return t.TargetSets, t.TargetReps, t.HoldSec, t.WeightKg, t.AssistKg
}
