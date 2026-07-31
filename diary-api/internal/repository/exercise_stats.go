package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
)

// GetExerciseStats returns per-session volume history for one exercise.
func (r *Repository) GetExerciseStats(exerciseID, period string) (models.ExerciseStats, error) {
	if period != models.StatsPeriod30d && period != models.StatsPeriodAll {
		return models.ExerciseStats{}, ErrInvalidPeriod
	}

	ex, err := r.GetExercise(exerciseID)
	if err != nil {
		return models.ExerciseStats{}, err
	}

	query := `
		SELECT ws.id, ws.performed_at, ws.is_deload,
		       COUNT(s.id) AS sets_count,
		       COALESCE(SUM(s.reps), 0) AS total_reps,
		       COALESCE(SUM(s.duration_sec), 0) AS total_hold,
		       wse.target_sets, wse.target_reps, wse.target_hold_sec
		FROM workout_session_exercises wse
		INNER JOIN workout_sessions ws ON ws.id = wse.workout_session_id
		INNER JOIN sets s ON s.workout_session_exercise_id = wse.id
		WHERE wse.exercise_id = ?`
	args := []any{exerciseID}

	if period == models.StatsPeriod30d {
		since := time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
		query += ` AND ws.performed_at >= ?`
		args = append(args, since)
	}

	query += `
		GROUP BY wse.id
		ORDER BY ws.performed_at ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return models.ExerciseStats{}, fmt.Errorf("query exercise stats: %w", err)
	}
	defer rows.Close()

	points := make([]models.ExerciseStatsPoint, 0)
	var totalSets int
	var totalVolume int

	for rows.Next() {
		var (
			sessionID                                      string
			performedAt                                    string
			deload                                         int
			setsCount, totalReps, totalHold                int
			targetSets, targetReps, targetHold             sql.NullInt64
		)
		if err := rows.Scan(
			&sessionID, &performedAt, &deload,
			&setsCount, &totalReps, &totalHold,
			&targetSets, &targetReps, &targetHold,
		); err != nil {
			return models.ExerciseStats{}, fmt.Errorf("scan exercise stats: %w", err)
		}

		volume := totalReps
		if ex.Kind == models.ExerciseKindHold {
			volume = totalHold
		}

		point := models.ExerciseStatsPoint{
			SessionID:   sessionID,
			PerformedAt: performedAt,
			IsDeload:    deload == 1,
			SetsCount:   setsCount,
			Volume:      volume,
		}
		if targetSets.Valid {
			tv := int(targetSets.Int64)
			if ex.Kind == models.ExerciseKindHold {
				if targetHold.Valid {
					tv *= int(targetHold.Int64)
					point.TargetVolume = &tv
				}
			} else if targetReps.Valid {
				tv *= int(targetReps.Int64)
				point.TargetVolume = &tv
			}
		}

		points = append(points, point)
		totalSets += setsCount
		totalVolume += volume
	}
	if err := rows.Err(); err != nil {
		return models.ExerciseStats{}, fmt.Errorf("exercise stats rows: %w", err)
	}

	summary := models.ExerciseStatsSummary{
		SessionCount: len(points),
		TotalVolume:  totalVolume,
	}
	if len(points) > 0 {
		summary.AvgSets = math.Round(float64(totalSets)/float64(len(points))*10) / 10
	}

	out := models.ExerciseStats{
		ExerciseID:          exerciseID,
		Kind:                ex.Kind,
		Period:              period,
		CurrentTargetVolume: currentTargetVolume(ex),
		Summary:             summary,
		Points:              points,
	}
	return out, nil
}

func currentTargetVolume(ex models.Exercise) *int {
	if ex.Target == nil {
		return nil
	}
	var v int
	if ex.Kind == models.ExerciseKindHold {
		v = ex.Target.TargetSets * ex.Target.HoldSec
	} else {
		v = ex.Target.TargetSets * ex.Target.TargetReps
	}
	return &v
}

// ErrInvalidPeriod is returned when stats period is not supported.
var ErrInvalidPeriod = errors.New("invalid stats period")
