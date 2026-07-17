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
	SetNumber int
	Reps      int
	WeightKg  float64
	AssistKg  float64
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

		wse := models.WorkoutSessionExercise{
			ID:               uuid.NewString(),
			WorkoutSessionID: session.ID,
			ExerciseID:       exIn.ExerciseID,
			Position:         i + 1,
			Sets:             make([]models.Set, 0, len(exIn.Sets)),
		}

		_, err = tx.Exec(
			`INSERT INTO workout_session_exercises (id, workout_session_id, exercise_id, position)
			 VALUES (?, ?, ?, ?)`,
			wse.ID, wse.WorkoutSessionID, wse.ExerciseID, wse.Position,
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
				WeightKg:                 setIn.WeightKg,
				AssistKg:                 setIn.AssistKg,
			}
			_, err = tx.Exec(
				`INSERT INTO sets (id, workout_session_exercise_id, set_number, reps, weight_kg, assist_kg)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				set.ID, set.WorkoutSessionExerciseID, set.SetNumber, set.Reps, set.WeightKg, set.AssistKg,
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
func (r *Repository) StartWorkoutSession(performedAt string, isDeload bool) (models.WorkoutSession, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	session := models.WorkoutSession{
		ID:          uuid.NewString(),
		PerformedAt: performedAt,
		StartedAt:   now,
		DurationSec: 0,
		IsDeload:    isDeload,
		CreatedAt:   now,
		Exercises:   []models.WorkoutSessionExercise{},
	}
	if session.PerformedAt == "" {
		session.PerformedAt = now
	}

	_, err := r.db.Exec(
		`INSERT INTO workout_sessions (id, performed_at, is_deload, created_at, started_at, duration_sec)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID, session.PerformedAt, boolToInt(session.IsDeload), session.CreatedAt,
		session.StartedAt, session.DurationSec,
	)
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("insert workout session: %w", err)
	}
	return session, nil
}

// FinishWorkoutSession stores total duration for a session.
func (r *Repository) FinishWorkoutSession(sessionID string, durationSec int) (models.WorkoutSession, error) {
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
	return r.GetWorkoutSession(sessionID)
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
		}
		_, err = tx.Exec(
			`INSERT INTO workout_session_exercises (id, workout_session_id, exercise_id, position)
			 VALUES (?, ?, ?, ?)`,
			wse.ID, wse.WorkoutSessionID, wse.ExerciseID, wse.Position,
		)
		if err != nil {
			return models.WorkoutSessionExercise{}, fmt.Errorf("insert workout session exercise: %w", err)
		}
	} else if err != nil {
		return models.WorkoutSessionExercise{}, fmt.Errorf("find workout session exercise: %w", err)
	} else if wse.Position != position {
		_, err = tx.Exec(
			`UPDATE workout_session_exercises SET position = ? WHERE id = ?`,
			position, wse.ID,
		)
		if err != nil {
			return models.WorkoutSessionExercise{}, fmt.Errorf("update position: %w", err)
		}
		wse.Position = position
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
			WeightKg:                 setIn.WeightKg,
			AssistKg:                 setIn.AssistKg,
		}
		_, err = tx.Exec(
			`INSERT INTO sets (id, workout_session_exercise_id, set_number, reps, weight_kg, assist_kg)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			set.ID, set.WorkoutSessionExerciseID, set.SetNumber, set.Reps, set.WeightKg, set.AssistKg,
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
		        GROUP_CONCAT(e.name, char(31) ORDER BY wse.position) AS exercise_names
		 FROM workout_sessions ws
		 INNER JOIN workout_session_exercises wse ON wse.workout_session_id = ws.id
		 INNER JOIN exercises e ON e.id = wse.exercise_id
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
		if err := rows.Scan(
			&s.ID, &s.PerformedAt, &deload, &s.CreatedAt, &s.DurationSec,
			&s.ExerciseCount, &namesRaw,
		); err != nil {
			return nil, fmt.Errorf("scan workout session summary: %w", err)
		}
		s.IsDeload = deload == 1
		if namesRaw.Valid && namesRaw.String != "" {
			s.ExerciseNames = strings.Split(namesRaw.String, "\x1f")
		} else {
			s.ExerciseNames = []string{}
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
	err := r.db.QueryRow(
		`SELECT id, performed_at, is_deload, created_at, started_at, duration_sec
		 FROM workout_sessions WHERE id = ?`,
		id,
	).Scan(
		&session.ID, &session.PerformedAt, &deload, &session.CreatedAt,
		&startedAt, &session.DurationSec,
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

	exRows, err := r.db.Query(
		`SELECT wse.id, wse.workout_session_id, wse.exercise_id, wse.position,
		        e.name, e.supports_assist
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
		if err := exRows.Scan(
			&wse.ID, &wse.WorkoutSessionID, &wse.ExerciseID, &wse.Position,
			&wse.ExerciseName, &assist,
		); err != nil {
			return models.WorkoutSession{}, fmt.Errorf("scan workout session exercise: %w", err)
		}
		wse.SupportsAssist = assist == 1

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
		`SELECT id, workout_session_exercise_id, set_number, reps, weight_kg, assist_kg
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
		if err := rows.Scan(&s.ID, &s.WorkoutSessionExerciseID, &s.SetNumber, &s.Reps, &s.WeightKg, &s.AssistKg); err != nil {
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
