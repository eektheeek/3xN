package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
	"github.com/google/uuid"
)

// CreateWorkoutPlanInput defines a new workout template.
type CreateWorkoutPlanInput struct {
	Name        string
	ExerciseIDs []string
}

// CreateWorkoutPlan stores a named plan with ordered exercises.
func (r *Repository) CreateWorkoutPlan(in CreateWorkoutPlanInput) (models.WorkoutPlan, error) {
	if in.Name == "" {
		return models.WorkoutPlan{}, fmt.Errorf("name is required")
	}
	if len(in.ExerciseIDs) == 0 {
		return models.WorkoutPlan{}, fmt.Errorf("at least one exercise is required")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return models.WorkoutPlan{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	plan := models.WorkoutPlan{
		ID:        uuid.NewString(),
		Name:      in.Name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	_, err = tx.Exec(
		`INSERT INTO workout_plans (id, name, created_at) VALUES (?, ?, ?)`,
		plan.ID, plan.Name, plan.CreatedAt,
	)
	if err != nil {
		return models.WorkoutPlan{}, fmt.Errorf("insert plan: %w", err)
	}

	for i, exID := range in.ExerciseIDs {
		var exists int
		err := tx.QueryRow(`SELECT 1 FROM exercises WHERE id = ?`, exID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return models.WorkoutPlan{}, fmt.Errorf("exercise %s: %w", exID, ErrNotFound)
		}
		if err != nil {
			return models.WorkoutPlan{}, fmt.Errorf("check exercise: %w", err)
		}

		_, err = tx.Exec(
			`INSERT INTO workout_plan_exercises (id, workout_plan_id, exercise_id, position)
			 VALUES (?, ?, ?, ?)`,
			uuid.NewString(), plan.ID, exID, i+1,
		)
		if err != nil {
			return models.WorkoutPlan{}, fmt.Errorf("insert plan exercise: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return models.WorkoutPlan{}, fmt.Errorf("commit: %w", err)
	}
	return r.GetWorkoutPlan(plan.ID)
}

// ListWorkoutPlans returns all saved templates with exercise preview.
func (r *Repository) ListWorkoutPlans() ([]models.WorkoutPlan, error) {
	rows, err := r.db.Query(
		`SELECT id, name, created_at FROM workout_plans ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	var out []models.WorkoutPlan
	for rows.Next() {
		var p models.WorkoutPlan
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		exercises, err := r.listPlanExercises(p.ID)
		if err != nil {
			return nil, err
		}
		p.Exercises = exercises
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plan rows: %w", err)
	}
	if out == nil {
		out = []models.WorkoutPlan{}
	}
	return out, nil
}

// GetWorkoutPlan returns one plan with exercises.
func (r *Repository) GetWorkoutPlan(id string) (models.WorkoutPlan, error) {
	var p models.WorkoutPlan
	err := r.db.QueryRow(
		`SELECT id, name, created_at FROM workout_plans WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.WorkoutPlan{}, ErrNotFound
	}
	if err != nil {
		return models.WorkoutPlan{}, fmt.Errorf("get plan: %w", err)
	}
	exercises, err := r.listPlanExercises(id)
	if err != nil {
		return models.WorkoutPlan{}, err
	}
	p.Exercises = exercises
	return p, nil
}

func (r *Repository) listPlanExercises(planID string) ([]models.WorkoutPlanExercise, error) {
	rows, err := r.db.Query(
		`SELECT wpe.id, wpe.workout_plan_id, wpe.exercise_id, wpe.position, e.name
		 FROM workout_plan_exercises wpe
		 JOIN exercises e ON e.id = wpe.exercise_id
		 WHERE wpe.workout_plan_id = ?
		 ORDER BY wpe.position ASC`,
		planID,
	)
	if err != nil {
		return nil, fmt.Errorf("list plan exercises: %w", err)
	}
	defer rows.Close()

	var out []models.WorkoutPlanExercise
	for rows.Next() {
		var pe models.WorkoutPlanExercise
		if err := rows.Scan(&pe.ID, &pe.WorkoutPlanID, &pe.ExerciseID, &pe.Position, &pe.ExerciseName); err != nil {
			return nil, fmt.Errorf("scan plan exercise: %w", err)
		}
		out = append(out, pe)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plan exercise rows: %w", err)
	}
	if out == nil {
		out = []models.WorkoutPlanExercise{}
	}
	return out, nil
}
