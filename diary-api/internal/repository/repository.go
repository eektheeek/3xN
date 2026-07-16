package repository

import "database/sql"

// Repository is the SQLite-backed data access layer.
type Repository struct {
	db *sql.DB
}

// New wraps an open database connection.
func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}
