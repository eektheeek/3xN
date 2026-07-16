package repository

import (
	"database/sql"
	"errors"
	"fmt"
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
		IsDeload:    in.IsDeload,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Exercises:   make([]models.WorkoutSessionExercise, 0, len(in.Exercises)),
	}
	if session.PerformedAt == "" {
		session.PerformedAt = session.CreatedAt
	}

	_, err = tx.Exec(
		`INSERT INTO workout_sessions (id, performed_at, is_deload, created_at)
		 VALUES (?, ?, ?, ?)`,
		session.ID, session.PerformedAt, boolToInt(session.IsDeload), session.CreatedAt,
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

// GetWorkoutSession returns a workout session with exercises and sets.
func (r *Repository) GetWorkoutSession(id string) (models.WorkoutSession, error) {
	var session models.WorkoutSession
	var deload int
	err := r.db.QueryRow(
		`SELECT id, performed_at, is_deload, created_at
		 FROM workout_sessions WHERE id = ?`,
		id,
	).Scan(&session.ID, &session.PerformedAt, &deload, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WorkoutSession{}, ErrNotFound
	}
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("get workout session: %w", err)
	}
	session.IsDeload = deload == 1

	exRows, err := r.db.Query(
		`SELECT id, workout_session_id, exercise_id, position
		 FROM workout_session_exercises
		 WHERE workout_session_id = ?
		 ORDER BY position ASC`,
		id,
	)
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("list workout session exercises: %w", err)
	}
	defer exRows.Close()

	for exRows.Next() {
		var wse models.WorkoutSessionExercise
		if err := exRows.Scan(&wse.ID, &wse.WorkoutSessionID, &wse.ExerciseID, &wse.Position); err != nil {
			return models.WorkoutSession{}, fmt.Errorf("scan workout session exercise: %w", err)
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
