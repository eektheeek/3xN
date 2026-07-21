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

// ListCycles returns all training cycles with steps (newest first).
func (r *Repository) ListCycles() ([]models.Cycle, error) {
	rows, err := r.db.Query(
		`SELECT id, name, current_step, on_home, created_at FROM cycles ORDER BY created_at DESC`,
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

// GetCycle returns one cycle with steps.
func (r *Repository) GetCycle(id string) (models.Cycle, error) {
	c, err := scanCycleRow(r.db.QueryRow(
		`SELECT id, name, current_step, on_home, created_at FROM cycles WHERE id = ?`,
		id,
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

// CreateCycle creates an empty named cycle.
func (r *Repository) CreateCycle(name string) (models.Cycle, error) {
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
		`INSERT INTO cycles (id, name, current_step, on_home, created_at) VALUES (?, ?, 1, 0, ?)`,
		c.ID, c.Name, c.CreatedAt,
	)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("insert cycle: %w", err)
	}
	return c, nil
}

func (r *Repository) listCycleSteps(cycleID string) ([]models.CycleStep, error) {
	rows, err := r.db.Query(
		`SELECT cs.id, cs.cycle_id, cs.position, cs.workout_plan_id, wp.name
		 FROM cycle_steps cs
		 INNER JOIN workout_plans wp ON wp.id = cs.workout_plan_id
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
		if err := rows.Scan(&s.ID, &s.CycleID, &s.Position, &s.WorkoutPlanID, &s.WorkoutPlanName); err != nil {
			return nil, fmt.Errorf("scan cycle step: %w", err)
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
func (r *Repository) ReplaceCycleSteps(cycleID string, workoutPlanIDs []string) (models.Cycle, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.Cycle{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var exists int
	err = tx.QueryRow(`SELECT 1 FROM cycles WHERE id = ?`, cycleID).Scan(&exists)
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
		err := tx.QueryRow(`SELECT 1 FROM workout_plans WHERE id = ?`, planID).Scan(&planExists)
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
	return r.GetCycle(cycleID)
}

// AdvanceCycle moves to the next step; finishing the last step marks the cycle completed.
// Any progress pins the cycle to the home screen.
func (r *Repository) AdvanceCycle(cycleID string) (models.Cycle, error) {
	c, err := r.GetCycle(cycleID)
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
	next := c.CurrentStep + 1
	_, err = r.db.Exec(
		`UPDATE cycles SET current_step = ?, on_home = 1 WHERE id = ?`,
		next, cycleID,
	)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("advance cycle: %w", err)
	}
	return r.GetCycle(cycleID)
}

// SetCycleOnHome toggles whether the cycle appears on the home screen.
func (r *Repository) SetCycleOnHome(cycleID string, onHome bool) (models.Cycle, error) {
	res, err := r.db.Exec(
		`UPDATE cycles SET on_home = ? WHERE id = ?`,
		boolToInt(onHome), cycleID,
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
	return r.GetCycle(cycleID)
}

// RestartCycle resets the cursor to step 1 (keeps on_home as-is).
func (r *Repository) RestartCycle(cycleID string) (models.Cycle, error) {
	var exists int
	err := r.db.QueryRow(`SELECT 1 FROM cycles WHERE id = ?`, cycleID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Cycle{}, ErrNotFound
	}
	if err != nil {
		return models.Cycle{}, fmt.Errorf("check cycle: %w", err)
	}
	_, err = r.db.Exec(`UPDATE cycles SET current_step = 1 WHERE id = ?`, cycleID)
	if err != nil {
		return models.Cycle{}, fmt.Errorf("restart cycle: %w", err)
	}
	return r.GetCycle(cycleID)
}
