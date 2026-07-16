package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

// API holds HTTP handlers backed by the repository.
type API struct {
	repo *repository.Repository
}

// NewAPI wires handlers to a repository.
func NewAPI(repo *repository.Repository) *API {
	return &API{repo: repo}
}

type createExerciseRequest struct {
	Name           string `json:"name"`
	MuscleGroup    string `json:"muscleGroup"`
	SupportsAssist bool   `json:"supportsAssist"`
}

type setTargetRequest struct {
	Sets     int     `json:"sets"`
	Reps     int     `json:"reps"`
	WeightKg float64 `json:"weightKg"`
	AssistKg float64 `json:"assistKg"`
}

// CreateExercise handles POST /v1/exercises.
func (a *API) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var req createExerciseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ex, err := a.repo.CreateExercise(req.Name, strings.TrimSpace(req.MuscleGroup), req.SupportsAssist)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create exercise")
		return
	}
	writeJSON(w, http.StatusCreated, ex)
}

// ListExercises handles GET /v1/exercises.
func (a *API) ListExercises(w http.ResponseWriter, r *http.Request) {
	list, err := a.repo.ListExercises()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list exercises")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GetExercise handles GET /v1/exercises/{id}.
func (a *API) GetExercise(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ex, err := a.repo.GetExercise(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get exercise")
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

// SetTarget handles PUT /v1/exercises/{id}/target.
func (a *API) SetTarget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req setTargetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Sets < 1 || req.Reps < 1 {
		writeError(w, http.StatusBadRequest, "sets and reps must be >= 1")
		return
	}
	if req.WeightKg < 0 || req.AssistKg < 0 {
		writeError(w, http.StatusBadRequest, "weightKg and assistKg must be >= 0")
		return
	}

	target, err := a.repo.SetTarget(id, req.Sets, req.Reps, req.WeightKg, req.AssistKg)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set target")
		return
	}
	writeJSON(w, http.StatusOK, target)
}
