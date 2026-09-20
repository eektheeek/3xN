package repository

import "errors"

// ErrNotFound indicates that the requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict indicates a uniqueness clash (e.g. email already registered).
var ErrConflict = errors.New("conflict")

// ErrInvalidPeriod indicates an unsupported stats period query value.
var ErrInvalidPeriod = errors.New("invalid period")

