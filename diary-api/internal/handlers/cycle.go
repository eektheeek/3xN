package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

// ListCycles handles GET /v1/cycles.
func (a *API) ListCycles(w http.ResponseWriter, r *http.Request) {
	cycles, err := a.repo.ListCycles(userIDFromRequest(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list cycles")
		return
	}
	writeJSON(w, http.StatusOK, cycles)
}

// GetCycle handles GET /v1/cycles/{id}.
func (a *API) GetCycle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	cycle, err := a.repo.GetCycle(userIDFromRequest(r), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "cycle not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load cycle")
		return
	}
	writeJSON(w, http.StatusOK, cycle)
}

type createCycleRequest struct {
	Name string `json:"name"`
}

// CreateCycle handles POST /v1/cycles.
func (a *API) CreateCycle(w http.ResponseWriter, r *http.Request) {
	var req createCycleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	cycle, err := a.repo.CreateCycle(userIDFromRequest(r), req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cycle)
}

type replaceCycleStepsRequest struct {
	WorkoutPlanIDs []string `json:"workoutPlanIds"`
}

// ReplaceCycleSteps handles PUT /v1/cycles/{id}/steps.
func (a *API) ReplaceCycleSteps(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	var req replaceCycleStepsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.WorkoutPlanIDs == nil {
		req.WorkoutPlanIDs = []string{}
	}

	cycle, err := a.repo.ReplaceCycleSteps(userIDFromRequest(r), id, req.WorkoutPlanIDs)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "cycle or workout plan not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cycle)
}

// AdvanceCycle handles POST /v1/cycles/{id}/advance.
func (a *API) AdvanceCycle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	var req struct {
		SessionID string `json:"sessionId"`
	}
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
	}
	cycle, err := a.repo.AdvanceCycle(userIDFromRequest(r), id, strings.TrimSpace(req.SessionID))
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "cycle not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cycle)
}

// RestartCycle handles POST /v1/cycles/{id}/restart.
func (a *API) RestartCycle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	cycle, err := a.repo.RestartCycle(userIDFromRequest(r), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "cycle not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to restart cycle")
		return
	}
	writeJSON(w, http.StatusOK, cycle)
}

type setCycleOnHomeRequest struct {
	OnHome bool `json:"onHome"`
}

// SetCycleOnHome handles PUT /v1/cycles/{id}/on-home.
func (a *API) SetCycleOnHome(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	var req setCycleOnHomeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cycle, err := a.repo.SetCycleOnHome(userIDFromRequest(r), id, req.OnHome)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "cycle not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cycle)
}

// RepeatCycle handles POST /v1/cycles/{id}/repeat.
func (a *API) RepeatCycle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	cycle, err := a.repo.RepeatCycle(userIDFromRequest(r), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "cycle not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cycle)
}
