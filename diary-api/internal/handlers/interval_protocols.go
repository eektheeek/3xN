package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

type createIntervalProtocolRequest struct {
	Name        string `json:"name"`
	PrepareSec  int    `json:"prepareSec"`
	WorkSec     int    `json:"workSec"`
	RestSec     int    `json:"restSec"`
	WarmupExtra bool   `json:"warmupExtra"`
}

// CreateIntervalProtocol handles POST /v1/interval-protocols.
func (a *API) CreateIntervalProtocol(w http.ResponseWriter, r *http.Request) {
	var req createIntervalProtocolRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.PrepareSec < 0 || req.WorkSec < 1 || req.RestSec < 0 {
		writeError(w, http.StatusBadRequest, "invalid prepareSec/workSec/restSec")
		return
	}

	p, err := a.repo.CreateIntervalProtocol(repository.CreateIntervalProtocolInput{
		Name:        req.Name,
		PrepareSec:  req.PrepareSec,
		WorkSec:     req.WorkSec,
		RestSec:     req.RestSec,
		WarmupExtra: req.WarmupExtra,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create interval protocol")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// ListIntervalProtocols handles GET /v1/interval-protocols.
func (a *API) ListIntervalProtocols(w http.ResponseWriter, r *http.Request) {
	list, err := a.repo.ListIntervalProtocols()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list interval protocols")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GetIntervalProtocol handles GET /v1/interval-protocols/{id}.
func (a *API) GetIntervalProtocol(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	p, err := a.repo.GetIntervalProtocol(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "interval protocol not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get interval protocol")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// UpdateIntervalProtocol handles PUT /v1/interval-protocols/{id}.
func (a *API) UpdateIntervalProtocol(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	var req createIntervalProtocolRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.PrepareSec < 0 || req.WorkSec < 1 || req.RestSec < 0 {
		writeError(w, http.StatusBadRequest, "invalid prepareSec/workSec/restSec")
		return
	}

	p, err := a.repo.UpdateIntervalProtocol(id, repository.CreateIntervalProtocolInput{
		Name:        req.Name,
		PrepareSec:  req.PrepareSec,
		WorkSec:     req.WorkSec,
		RestSec:     req.RestSec,
		WarmupExtra: req.WarmupExtra,
	})
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "interval protocol not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update interval protocol")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// DeleteIntervalProtocol handles DELETE /v1/interval-protocols/{id}.
func (a *API) DeleteIntervalProtocol(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	err := a.repo.DeleteIntervalProtocol(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "interval protocol not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete interval protocol")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
