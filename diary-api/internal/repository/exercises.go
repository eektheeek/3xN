package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
	"github.com/google/uuid"
)

// CreateExercise inserts a new exercise into the catalog.
func (r *Repository) CreateExercise(name, muscleGroup string, supportsAssist bool) (models.Exercise, error) {
	ex := models.Exercise{
		ID:             uuid.NewString(),
		Name:           name,
		MuscleGroup:    muscleGroup,
		SupportsAssist: supportsAssist,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	_, err := r.db.Exec(
		`INSERT INTO exercises (id, name, muscle_group, supports_assist, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		ex.ID, ex.Name, ex.MuscleGroup, boolToInt(ex.SupportsAssist), ex.CreatedAt,
	)
	if err != nil {
		return models.Exercise{}, fmt.Errorf("insert exercise: %w", err)
	}
	return ex, nil
}

// ListExercises returns all exercises with targets when present.
func (r *Repository) ListExercises() ([]models.Exercise, error) {
	rows, err := r.db.Query(
		`SELECT e.id, e.name, e.muscle_group, e.supports_assist, e.created_at,
		        t.target_sets, t.target_reps, t.weight_kg, t.assist_kg
		 FROM exercises e
		 LEFT JOIN exercise_targets t ON t.exercise_id = e.id
		 ORDER BY e.created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	var out []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var assist int
		var targetSets, targetReps sql.NullInt64
		var weightKg, assistKg sql.NullFloat64
		if err := rows.Scan(
			&ex.ID, &ex.Name, &ex.MuscleGroup, &assist, &ex.CreatedAt,
			&targetSets, &targetReps, &weightKg, &assistKg,
		); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		ex.SupportsAssist = assist == 1
		if targetSets.Valid {
			ex.Target = &models.ExerciseTarget{
				ExerciseID: ex.ID,
				TargetSets: int(targetSets.Int64),
				TargetReps: int(targetReps.Int64),
				WeightKg:   weightKg.Float64,
				AssistKg:   assistKg.Float64,
			}
		}
		out = append(out, ex)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list exercises rows: %w", err)
	}
	if out == nil {
		out = []models.Exercise{}
	}
	return out, nil
}

// GetExercise returns one exercise and its target when present.
func (r *Repository) GetExercise(id string) (models.Exercise, error) {
	var ex models.Exercise
	var assist int
	err := r.db.QueryRow(
		`SELECT id, name, muscle_group, supports_assist, created_at
		 FROM exercises WHERE id = ?`,
		id,
	).Scan(&ex.ID, &ex.Name, &ex.MuscleGroup, &assist, &ex.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Exercise{}, ErrNotFound
	}
	if err != nil {
		return models.Exercise{}, fmt.Errorf("get exercise: %w", err)
	}
	ex.SupportsAssist = assist == 1

	target, err := r.getTarget(id)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return models.Exercise{}, err
	}
	if err == nil {
		ex.Target = &target
	}
	return ex, nil
}

// SetTarget upserts the current goal for an exercise.
func (r *Repository) SetTarget(exerciseID string, sets, reps int, weightKg, assistKg float64) (models.ExerciseTarget, error) {
	if _, err := r.GetExercise(exerciseID); err != nil {
		return models.ExerciseTarget{}, err
	}

	target := models.ExerciseTarget{
		ExerciseID: exerciseID,
		TargetSets: sets,
		TargetReps: reps,
		WeightKg:   weightKg,
		AssistKg:   assistKg,
	}

	_, err := r.db.Exec(
		`INSERT INTO exercise_targets (exercise_id, target_sets, target_reps, weight_kg, assist_kg)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(exercise_id) DO UPDATE SET
		   target_sets = excluded.target_sets,
		   target_reps = excluded.target_reps,
		   weight_kg = excluded.weight_kg,
		   assist_kg = excluded.assist_kg`,
		target.ExerciseID, target.TargetSets, target.TargetReps, target.WeightKg, target.AssistKg,
	)
	if err != nil {
		return models.ExerciseTarget{}, fmt.Errorf("upsert target: %w", err)
	}
	return target, nil
}

func (r *Repository) getTarget(exerciseID string) (models.ExerciseTarget, error) {
	var t models.ExerciseTarget
	err := r.db.QueryRow(
		`SELECT exercise_id, target_sets, target_reps, weight_kg, assist_kg
		 FROM exercise_targets WHERE exercise_id = ?`,
		exerciseID,
	).Scan(&t.ExerciseID, &t.TargetSets, &t.TargetReps, &t.WeightKg, &t.AssistKg)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ExerciseTarget{}, ErrNotFound
	}
	if err != nil {
		return models.ExerciseTarget{}, fmt.Errorf("get target: %w", err)
	}
	return t, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
