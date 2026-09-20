package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/auth"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
	"github.com/google/uuid"
)

// User is the persisted account row (password hash never exported on API models).
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    string
}

// PublicUser is the JSON-safe user projection.
func (u User) Public() models.User {
	return models.User{ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt}
}

// EnsureBootstrapOwner creates the legacy owner account if needed and assigns
// all rows with empty user_id to that account.
// DIARY_BOOTSTRAP_PASSWORD is required when creating the user for the first time.
func (r *Repository) EnsureBootstrapOwner() error {
	email := auth.NormalizeEmail(auth.BootstrapEmail)

	user, err := r.GetUserByEmail(email)
	if errors.Is(err, ErrNotFound) {
		password := os.Getenv("DIARY_BOOTSTRAP_PASSWORD")
		if password == "" {
			return fmt.Errorf("DIARY_BOOTSTRAP_PASSWORD is required to create bootstrap user %s", email)
		}
		if err := auth.ValidatePassword(password); err != nil {
			return err
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		user, err = r.CreateUser(email, hash)
		if err != nil {
			return fmt.Errorf("create bootstrap user: %w", err)
		}
	} else if err != nil {
		return err
	}

	if err := r.assignEmptyUserIDs(user.ID); err != nil {
		return fmt.Errorf("backfill user_id: %w", err)
	}
	return nil
}

func (r *Repository) assignEmptyUserIDs(userID string) error {
	tables := []string{
		"exercises",
		"interval_protocols",
		"workout_plans",
		"cycles",
		"workout_sessions",
	}
	for _, table := range tables {
		q := fmt.Sprintf(`UPDATE %s SET user_id = ? WHERE user_id = ''`, table)
		if _, err := r.db.Exec(q, userID); err != nil {
			return fmt.Errorf("update %s: %w", table, err)
		}
	}
	return nil
}

// CreateUser inserts a new user with a precomputed password hash.
func (r *Repository) CreateUser(email, passwordHash string) (User, error) {
	email = auth.NormalizeEmail(email)
	if email == "" {
		return User{}, fmt.Errorf("email is required")
	}
	u := User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, created_at) VALUES (?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return User{}, ErrConflict
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

// GetUserByEmail loads a user by normalized email.
func (r *Repository) GetUserByEmail(email string) (User, error) {
	email = auth.NormalizeEmail(email)
	var u User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetUserByID loads a user by id.
func (r *Repository) GetUserByID(id string) (User, error) {
	var u User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// UpdateUserPasswordHash sets a new password hash for the user.
func (r *Repository) UpdateUserPasswordHash(userID, passwordHash string) error {
	res, err := r.db.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSession stores a hashed session token and returns its id.
func (r *Repository) CreateSession(userID, tokenHash string, expiresAt time.Time) (string, error) {
	id := uuid.NewString()
	created := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, userID, tokenHash, expiresAt.UTC().Format(time.RFC3339), created,
	)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return id, nil
}

// UserIDByTokenHash resolves a non-expired session to a user id.
func (r *Repository) UserIDByTokenHash(tokenHash string) (userID string, sessionID string, err error) {
	now := time.Now().UTC().Format(time.RFC3339)
	err = r.db.QueryRow(
		`SELECT user_id, id FROM sessions WHERE token_hash = ? AND expires_at > ?`,
		tokenHash, now,
	).Scan(&userID, &sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("lookup session: %w", err)
	}
	return userID, sessionID, nil
}

// DeleteSession removes one session by id.
func (r *Repository) DeleteSession(sessionID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteOtherSessionsForUser revokes all sessions except keepSessionID.
func (r *Repository) DeleteOtherSessionsForUser(userID, keepSessionID string) error {
	_, err := r.db.Exec(
		`DELETE FROM sessions WHERE user_id = ? AND id != ?`,
		userID, keepSessionID,
	)
	if err != nil {
		return fmt.Errorf("delete other sessions: %w", err)
	}
	return nil
}
