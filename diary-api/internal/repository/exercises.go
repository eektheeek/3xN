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
func (r *Repository) CreateExercise(name, muscleGroup string, supportsAssist bool, protocolID string) (models.Exercise, error) {
	if protocolID != "" {
		if _, err := r.GetIntervalProtocol(protocolID); err != nil {
			return models.Exercise{}, err
		}
	}

	ex := models.Exercise{
		ID:             uuid.NewString(),
		Name:           name,
		MuscleGroup:    muscleGroup,
		SupportsAssist: supportsAssist,
		ProtocolID:     protocolID,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	var protocolArg any
	if protocolID != "" {
		protocolArg = protocolID
	}

	_, err := r.db.Exec(
		`INSERT INTO exercises (id, name, muscle_group, supports_assist, created_at, protocol_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		ex.ID, ex.Name, ex.MuscleGroup, boolToInt(ex.SupportsAssist), ex.CreatedAt, protocolArg,
	)
	if err != nil {
		return models.Exercise{}, fmt.Errorf("insert exercise: %w", err)
	}
	return r.GetExercise(ex.ID)
}

// UpdateExercise updates catalog fields for an existing exercise.
// protocolID empty string clears the attachment. When supportsAssist is turned off, assistKg is cleared.
func (r *Repository) UpdateExercise(id, name, muscleGroup string, supportsAssist bool, protocolID string) (models.Exercise, error) {
	if name == "" {
		return models.Exercise{}, fmt.Errorf("name is required")
	}
	if protocolID != "" {
		if _, err := r.GetIntervalProtocol(protocolID); err != nil {
			return models.Exercise{}, err
		}
	}

	var protocolArg any
	if protocolID != "" {
		protocolArg = protocolID
	}

	res, err := r.db.Exec(
		`UPDATE exercises
		 SET name = ?, muscle_group = ?, supports_assist = ?, protocol_id = ?
		 WHERE id = ?`,
		name, muscleGroup, boolToInt(supportsAssist), protocolArg, id,
	)
	if err != nil {
		return models.Exercise{}, fmt.Errorf("update exercise: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return models.Exercise{}, fmt.Errorf("update exercise rows: %w", err)
	}
	if n == 0 {
		return models.Exercise{}, ErrNotFound
	}

	if !supportsAssist {
		if _, err := r.db.Exec(
			`UPDATE exercise_targets SET assist_kg = 0 WHERE exercise_id = ?`,
			id,
		); err != nil {
			return models.Exercise{}, fmt.Errorf("clear assist on target: %w", err)
		}
	}

	return r.GetExercise(id)
}

const exerciseSelectSQL = `
SELECT e.id, e.name, e.muscle_group, e.supports_assist, e.created_at, e.protocol_id,
       t.target_sets, t.target_reps, t.weight_kg, t.assist_kg,
       p.name, p.work_sec, p.rest_sec, p.warmup_extra, p.created_at, p.prepare_sec
FROM exercises e
LEFT JOIN exercise_targets t ON t.exercise_id = e.id
LEFT JOIN interval_protocols p ON p.id = e.protocol_id`

// ListExercises returns all exercises with targets and protocols when present.
func (r *Repository) ListExercises() ([]models.Exercise, error) {
	rows, err := r.db.Query(exerciseSelectSQL + ` ORDER BY e.created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	var out []models.Exercise
	for rows.Next() {
		ex, err := scanExercise(rows)
		if err != nil {
			return nil, err
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

// GetExercise returns one exercise and its target/protocol when present.
func (r *Repository) GetExercise(id string) (models.Exercise, error) {
	ex, err := scanExercise(r.db.QueryRow(exerciseSelectSQL+` WHERE e.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return models.Exercise{}, ErrNotFound
	}
	if err != nil {
		return models.Exercise{}, err
	}
	return ex, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanExercise(row scannable) (models.Exercise, error) {
	var ex models.Exercise
	var assist int
	var protocolID sql.NullString
	var targetSets, targetReps sql.NullInt64
	var weightKg, assistKg sql.NullFloat64
	var pName sql.NullString
	var pWork, pRest, pWarmup, pPrepare sql.NullInt64
	var pCreated sql.NullString

	err := row.Scan(
		&ex.ID, &ex.Name, &ex.MuscleGroup, &assist, &ex.CreatedAt, &protocolID,
		&targetSets, &targetReps, &weightKg, &assistKg,
		&pName, &pWork, &pRest, &pWarmup, &pCreated, &pPrepare,
	)
	if err != nil {
		return models.Exercise{}, err
	}
	ex.SupportsAssist = assist == 1
	if protocolID.Valid {
		ex.ProtocolID = protocolID.String
	}
	if targetSets.Valid {
		ex.Target = &models.ExerciseTarget{
			ExerciseID: ex.ID,
			TargetSets: int(targetSets.Int64),
			TargetReps: int(targetReps.Int64),
			WeightKg:   weightKg.Float64,
			AssistKg:   assistKg.Float64,
		}
	}
	if pName.Valid && ex.ProtocolID != "" {
		ex.Protocol = &models.IntervalProtocol{
			ID:          ex.ProtocolID,
			Name:        pName.String,
			PrepareSec:  int(pPrepare.Int64),
			WorkSec:     int(pWork.Int64),
			RestSec:     int(pRest.Int64),
			WarmupExtra: pWarmup.Int64 == 1,
			CreatedAt:   pCreated.String,
		}
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

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
