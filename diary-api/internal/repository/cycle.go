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

func finalizeCycle(c *models.Cycle) {
	n := len(c.Steps)
	if n == 0 {
		c.CurrentStep = 1
		c.Completed = false
		return
	}
	if c.CurrentStep < 1 {
		c.CurrentStep = 1
	}
	if c.CurrentStep > n {
		c.CurrentStep = n + 1
		c.Completed = true
		return
	}
	c.Completed = false
}

func scanCycleRow(scanner interface {
	Scan(dest ...any) error
}) (models.Cycle, error) {
	var c models.Cycle
	var onHome int
	err := scanner.Scan(&c.ID, &c.Name, &c.CurrentStep, &onHome, &c.CreatedAt)
	if err != nil {
		return models.Cycle{}, err
	}
	c.OnHome = onHome == 1
	return c, nil
}

// ListCycles returns training cycles owned by userID (newest first).
func (r *Repository) ListCycles(userID string) ([]models.Cycle, error) {
	rows, err := r.db.Query(
		`SELECT id, name, current_step, on_home, created_at FROM cycles WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list cycles: %w", err)
	}
	defer rows.Close()

	var out []models.Cycle
	for rows.Next() {
		c, err := scanCycleRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cycle: %w", err)
		}
		steps, err := r.listCycleSteps(c.ID)
		if err != nil {
			return nil, err
		}
		c.Steps = steps
		finalizeCycle(&c)
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list cycles rows: %w", err)
	}
	if out == nil {
		out = []models.Cycle{}
	}
	return out, nil
}

// GetCycle returns one cycle with steps owned by userID.
func (r *Repository) GetCycle(userID, id string) (models.Cycle, error) {
	c, err := scanCycleRow(r.db.QueryRow(
		`SELECT id, name, current_step, on_home, created_at FROM cycles WHERE id = ? AND user_id = ?`,
		id, userID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return models.Cycle{}, ErrNotFound
	}
	if err != nil {
		return models.Cycle{}, fmt.Errorf("get cycle: %w", err)
	}
	steps, err := r.listCycleSteps(c.ID)
	if err != nil {
		return models.Cycle{}, err
	}
	c.Steps = steps
	finalizeCycle(&c)
	return c, nil
}

// CreateCycle creates an empty named cycle for userID.
func (r *Repository) CreateCycle(userID, name string) (models.Cycle, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Cycle{}, fmt.Errorf("name is required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	c := models.Cycle{
		ID:          uuid.NewString(),
		Name:        name,
		CurrentStep: 1,
		OnHome:      false,
		CreatedAt:   now,
		Steps:       []models.CycleStep{},
	}
	_, err := r.db.Exec(
		`INSERT INTO cycles (id, name, current_step, on_home, created_at, user_id) VALUES (?, ?, 1, 0, ?, ?)`,
		c.ID, c.Name, c.CreatedAt, userID,
	)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("insert cycle: %w", err)
	}
	return c, nil
}

func (r *Repository) listCycleSteps(cycleID string) ([]models.CycleStep, error) {
	rows, err := r.db.Query(
		`SELECT cs.id, cs.cycle_id, cs.position, cs.workout_plan_id, wp.name,
		        cs.completed_session_id, ws.performed_at
		 FROM cycle_steps cs
		 INNER JOIN workout_plans wp ON wp.id = cs.workout_plan_id
		 LEFT JOIN workout_sessions ws ON ws.id = cs.completed_session_id
		 WHERE cs.cycle_id = ?
		 ORDER BY cs.position ASC`,
		cycleID,
	)
	if err != nil {
		return nil, fmt.Errorf("list cycle steps: %w", err)
	}
	defer rows.Close()

	var out []models.CycleStep
	for rows.Next() {
		var s models.CycleStep
		var sessionID, performedAt sql.NullString
		if err := rows.Scan(
			&s.ID, &s.CycleID, &s.Position, &s.WorkoutPlanID, &s.WorkoutPlanName,
			&sessionID, &performedAt,
		); err != nil {
			return nil, fmt.Errorf("scan cycle step: %w", err)
		}
		if sessionID.Valid {
			s.CompletedSessionID = sessionID.String
		}
		if performedAt.Valid {
			s.CompletedPerformedAt = performedAt.String
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cycle steps rows: %w", err)
	}
	if out == nil {
		out = []models.CycleStep{}
	}
	return out, nil
}

// ReplaceCycleSteps replaces the full ordered list of workout_plan IDs for a cycle.
func (r *Repository) ReplaceCycleSteps(userID, cycleID string, workoutPlanIDs []string) (models.Cycle, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.Cycle{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var exists int
	err = tx.QueryRow(`SELECT 1 FROM cycles WHERE id = ? AND user_id = ?`, cycleID, userID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Cycle{}, ErrNotFound
	}
	if err != nil {
		return models.Cycle{}, fmt.Errorf("check cycle: %w", err)
	}

	for i, planID := range workoutPlanIDs {
		if planID == "" {
			return models.Cycle{}, fmt.Errorf("workoutPlanId at index %d is empty", i)
		}
		var planExists int
		err := tx.QueryRow(`SELECT 1 FROM workout_plans WHERE id = ? AND user_id = ?`, planID, userID).Scan(&planExists)
		if errors.Is(err, sql.ErrNoRows) {
			return models.Cycle{}, fmt.Errorf("workout plan %s: %w", planID, ErrNotFound)
		}
		if err != nil {
			return models.Cycle{}, fmt.Errorf("check workout plan %s: %w", planID, err)
		}
	}

	if _, err := tx.Exec(`DELETE FROM cycle_steps WHERE cycle_id = ?`, cycleID); err != nil {
		return models.Cycle{}, fmt.Errorf("clear cycle steps: %w", err)
	}

	for i, planID := range workoutPlanIDs {
		_, err = tx.Exec(
			`INSERT INTO cycle_steps (id, cycle_id, position, workout_plan_id)
			 VALUES (?, ?, ?, ?)`,
			uuid.NewString(), cycleID, i+1, planID,
		)
		if err != nil {
			return models.Cycle{}, fmt.Errorf("insert cycle step: %w", err)
		}
	}

	var currentStep int
	err = tx.QueryRow(`SELECT current_step FROM cycles WHERE id = ?`, cycleID).Scan(&currentStep)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("read current_step: %w", err)
	}
	n := len(workoutPlanIDs)
	newStep := currentStep
	if n == 0 {
		newStep = 1
	} else if currentStep < 1 {
		newStep = 1
	} else if currentStep > n {
		newStep = n + 1 // keep completed if stack still non-empty
	}
	if newStep != currentStep {
		_, err = tx.Exec(`UPDATE cycles SET current_step = ? WHERE id = ?`, newStep, cycleID)
		if err != nil {
			return models.Cycle{}, fmt.Errorf("normalize current_step: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Cycle{}, fmt.Errorf("commit: %w", err)
	}
	return r.GetCycle(userID, cycleID)
}

// AdvanceCycle moves to the next step; finishing the last step marks the cycle completed.
// Optional sessionID links the step being finished to a diary workout session.
func (r *Repository) AdvanceCycle(userID, cycleID string, sessionID string) (models.Cycle, error) {
	c, err := r.GetCycle(userID, cycleID)
	if err != nil {
		return models.Cycle{}, err
	}
	n := len(c.Steps)
	if n == 0 {
		return c, nil
	}
	if c.Completed {
		return c, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return models.Cycle{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		var exists int
		err = tx.QueryRow(`SELECT 1 FROM workout_sessions WHERE id = ? AND user_id = ?`, sessionID, userID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return models.Cycle{}, fmt.Errorf("workout session not found")
		}
		if err != nil {
			return models.Cycle{}, fmt.Errorf("check session: %w", err)
		}
		_, err = tx.Exec(
			`UPDATE cycle_steps SET completed_session_id = ?
			 WHERE cycle_id = ? AND position = ?`,
			sessionID, cycleID, c.CurrentStep,
		)
		if err != nil {
			return models.Cycle{}, fmt.Errorf("link step session: %w", err)
		}
	}

	next := c.CurrentStep + 1
	onHome := 1
	if next > n {
		onHome = 0 // completed runs leave the home screen; stay in history
	}
	_, err = tx.Exec(
		`UPDATE cycles SET current_step = ?, on_home = ? WHERE id = ?`,
		next, onHome, cycleID,
	)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("advance cycle: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return models.Cycle{}, fmt.Errorf("commit: %w", err)
	}
	return r.GetCycle(userID, cycleID)
}

// SetCycleOnHome toggles whether the cycle appears on the home screen.
func (r *Repository) SetCycleOnHome(userID, cycleID string, onHome bool) (models.Cycle, error) {
	c, err := r.GetCycle(userID, cycleID)
	if err != nil {
		return models.Cycle{}, err
	}
	if c.Completed && onHome {
		return models.Cycle{}, fmt.Errorf("completed cycle cannot be pinned to home; use Repeat")
	}
	res, err := r.db.Exec(
		`UPDATE cycles SET on_home = ? WHERE id = ? AND user_id = ?`,
		boolToInt(onHome), cycleID, userID,
	)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("set cycle on_home: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return models.Cycle{}, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return models.Cycle{}, ErrNotFound
	}
	return r.GetCycle(userID, cycleID)
}

// RepeatCycle clones a cycle (usually completed) into a fresh active run with the same steps.
func (r *Repository) RepeatCycle(userID, cycleID string) (models.Cycle, error) {
	src, err := r.GetCycle(userID, cycleID)
	if err != nil {
		return models.Cycle{}, err
	}
	if len(src.Steps) == 0 {
		return models.Cycle{}, fmt.Errorf("cycle has no steps to repeat")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return models.Cycle{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	newID := uuid.NewString()
	_, err = tx.Exec(
		`INSERT INTO cycles (id, name, current_step, on_home, created_at, user_id) VALUES (?, ?, 1, 1, ?, ?)`,
		newID, src.Name, now, userID,
	)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("insert repeated cycle: %w", err)
	}
	for _, step := range src.Steps {
		_, err = tx.Exec(
			`INSERT INTO cycle_steps (id, cycle_id, position, workout_plan_id)
			 VALUES (?, ?, ?, ?)`,
			uuid.NewString(), newID, step.Position, step.WorkoutPlanID,
		)
		if err != nil {
			return models.Cycle{}, fmt.Errorf("clone cycle step: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return models.Cycle{}, fmt.Errorf("commit: %w", err)
	}
	return r.GetCycle(userID, newID)
}

// RestartCycle resets the cursor to step 1 (keeps on_home as-is).
// Prefer RepeatCycle for completed runs so history is preserved.
func (r *Repository) RestartCycle(userID, cycleID string) (models.Cycle, error) {
	var exists int
	err := r.db.QueryRow(`SELECT 1 FROM cycles WHERE id = ? AND user_id = ?`, cycleID, userID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Cycle{}, ErrNotFound
	}
	if err != nil {
		return models.Cycle{}, fmt.Errorf("check cycle: %w", err)
	}
	_, err = r.db.Exec(`UPDATE cycles SET current_step = 1 WHERE id = ? AND user_id = ?`, cycleID, userID)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("restart cycle: %w", err)
	}
	return r.GetCycle(userID, cycleID)
}
