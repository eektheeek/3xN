package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/models"
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
	Kind           string `json:"kind"`
	SupportsAssist bool   `json:"supportsAssist"`
	ProtocolID     string `json:"protocolId"`
}

type setTargetRequest struct {
	Sets     int     `json:"sets"`
	Reps     int     `json:"reps"`
	HoldSec  int     `json:"holdSec"`
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

	ex, err := a.repo.CreateExercise(
		userIDFromRequest(r),
		req.Name,
		strings.TrimSpace(req.MuscleGroup),
		strings.TrimSpace(req.Kind),
		req.SupportsAssist,
		strings.TrimSpace(req.ProtocolID),
	)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "interval protocol not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create exercise")
		return
	}
	writeJSON(w, http.StatusCreated, ex)
}

type updateExerciseRequest struct {
	Name           string `json:"name"`
	MuscleGroup    string `json:"muscleGroup"`
	Kind           string `json:"kind"`
	SupportsAssist bool   `json:"supportsAssist"`
	ProtocolID     string `json:"protocolId"`
}

// UpdateExercise handles PUT /v1/exercises/{id}.
func (a *API) UpdateExercise(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req updateExerciseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ex, err := a.repo.UpdateExercise(
		userIDFromRequest(r),
		id,
		req.Name,
		strings.TrimSpace(req.MuscleGroup),
		strings.TrimSpace(req.Kind),
		req.SupportsAssist,
		strings.TrimSpace(req.ProtocolID),
	)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise or interval protocol not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update exercise")
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

// ListExercises handles GET /v1/exercises.
func (a *API) ListExercises(w http.ResponseWriter, r *http.Request) {
	list, err := a.repo.ListExercises(userIDFromRequest(r))
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

	ex, err := a.repo.GetExercise(userIDFromRequest(r), id)
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

// GetExerciseStats handles GET /v1/exercises/{id}/stats?period=30d|all.
func (a *API) GetExerciseStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = models.StatsPeriod30d
	}

	stats, err := a.repo.GetExerciseStats(userIDFromRequest(r), id, period)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if errors.Is(err, repository.ErrInvalidPeriod) {
		writeError(w, http.StatusBadRequest, "period must be 30d or all")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get exercise stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
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
	if req.Sets < 1 {
		writeError(w, http.StatusBadRequest, "sets must be >= 1")
		return
	}
	if req.WeightKg < 0 || req.AssistKg < 0 {
		writeError(w, http.StatusBadRequest, "weightKg and assistKg must be >= 0")
		return
	}
	if req.Reps < 0 || req.HoldSec < 0 {
		writeError(w, http.StatusBadRequest, "reps and holdSec must be >= 0")
		return
	}

	ex, err := a.repo.GetExercise(userIDFromRequest(r), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get exercise")
		return
	}
	if ex.Kind == models.ExerciseKindHold {
		if req.HoldSec < 1 {
			writeError(w, http.StatusBadRequest, "holdSec must be >= 1 for hold exercises")
			return
		}
	} else if req.Reps < 1 {
		writeError(w, http.StatusBadRequest, "reps must be >= 1 for reps exercises")
		return
	}

	target, err := a.repo.SetTarget(userIDFromRequest(r), id, req.Sets, req.Reps, req.HoldSec, req.WeightKg, req.AssistKg)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, target)
}
