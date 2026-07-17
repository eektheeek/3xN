package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
	"github.com/google/uuid"
)

// CreateIntervalProtocolInput defines a new saved Tabata template.
type CreateIntervalProtocolInput struct {
	Name        string
	PrepareSec  int
	WorkSec     int
	RestSec     int
	WarmupExtra bool
}

// CreateIntervalProtocol stores a reusable interval protocol.
func (r *Repository) CreateIntervalProtocol(in CreateIntervalProtocolInput) (models.IntervalProtocol, error) {
	if in.Name == "" {
		return models.IntervalProtocol{}, fmt.Errorf("name is required")
	}
	if in.PrepareSec < 0 {
		return models.IntervalProtocol{}, fmt.Errorf("prepareSec must be >= 0")
	}
	if in.WorkSec < 1 {
		return models.IntervalProtocol{}, fmt.Errorf("workSec must be >= 1")
	}
	if in.RestSec < 0 {
		return models.IntervalProtocol{}, fmt.Errorf("restSec must be >= 0")
	}

	p := models.IntervalProtocol{
		ID:          uuid.NewString(),
		Name:        in.Name,
		PrepareSec:  in.PrepareSec,
		WorkSec:     in.WorkSec,
		RestSec:     in.RestSec,
		WarmupExtra: in.WarmupExtra,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	_, err := r.db.Exec(
		`INSERT INTO interval_protocols (id, name, work_sec, rest_sec, warmup_extra, created_at, prepare_sec)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.WorkSec, p.RestSec, boolToInt(p.WarmupExtra), p.CreatedAt, p.PrepareSec,
	)
	if err != nil {
		return models.IntervalProtocol{}, fmt.Errorf("insert interval protocol: %w", err)
	}
	return p, nil
}

// ListIntervalProtocols returns all saved protocols, newest first.
func (r *Repository) ListIntervalProtocols() ([]models.IntervalProtocol, error) {
	rows, err := r.db.Query(
		`SELECT id, name, work_sec, rest_sec, warmup_extra, created_at, prepare_sec
		 FROM interval_protocols
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list interval protocols: %w", err)
	}
	defer rows.Close()

	var out []models.IntervalProtocol
	for rows.Next() {
		p, err := scanProtocol(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list interval protocols rows: %w", err)
	}
	if out == nil {
		out = []models.IntervalProtocol{}
	}
	return out, nil
}

// GetIntervalProtocol returns one protocol by id.
func (r *Repository) GetIntervalProtocol(id string) (models.IntervalProtocol, error) {
	p, err := scanProtocol(r.db.QueryRow(
		`SELECT id, name, work_sec, rest_sec, warmup_extra, created_at, prepare_sec
		 FROM interval_protocols WHERE id = ?`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return models.IntervalProtocol{}, ErrNotFound
	}
	if err != nil {
		return models.IntervalProtocol{}, err
	}
	return p, nil
}

type protocolScanner interface {
	Scan(dest ...any) error
}

func scanProtocol(row protocolScanner) (models.IntervalProtocol, error) {
	var p models.IntervalProtocol
	var warmup int
	err := row.Scan(&p.ID, &p.Name, &p.WorkSec, &p.RestSec, &warmup, &p.CreatedAt, &p.PrepareSec)
	if err != nil {
		return models.IntervalProtocol{}, err
	}
	p.WarmupExtra = warmup == 1
	return p, nil
}

// UpdateIntervalProtocol updates a saved protocol.
func (r *Repository) UpdateIntervalProtocol(id string, in CreateIntervalProtocolInput) (models.IntervalProtocol, error) {
	if in.Name == "" {
		return models.IntervalProtocol{}, fmt.Errorf("name is required")
	}
	if in.PrepareSec < 0 {
		return models.IntervalProtocol{}, fmt.Errorf("prepareSec must be >= 0")
	}
	if in.WorkSec < 1 {
		return models.IntervalProtocol{}, fmt.Errorf("workSec must be >= 1")
	}
	if in.RestSec < 0 {
		return models.IntervalProtocol{}, fmt.Errorf("restSec must be >= 0")
	}

	res, err := r.db.Exec(
		`UPDATE interval_protocols
		 SET name = ?, work_sec = ?, rest_sec = ?, warmup_extra = ?, prepare_sec = ?
		 WHERE id = ?`,
		in.Name, in.WorkSec, in.RestSec, boolToInt(in.WarmupExtra), in.PrepareSec, id,
	)
	if err != nil {
		return models.IntervalProtocol{}, fmt.Errorf("update interval protocol: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return models.IntervalProtocol{}, fmt.Errorf("update interval protocol rows: %w", err)
	}
	if n == 0 {
		return models.IntervalProtocol{}, ErrNotFound
	}
	return r.GetIntervalProtocol(id)
}

// DeleteIntervalProtocol removes a protocol (exercises.protocol_id becomes NULL via FK).
func (r *Repository) DeleteIntervalProtocol(id string) error {
	res, err := r.db.Exec(`DELETE FROM interval_protocols WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete interval protocol: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete interval protocol rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
