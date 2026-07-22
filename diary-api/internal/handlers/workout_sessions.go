package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

type createWorkoutSessionRequest struct {
	PerformedAt string                                `json:"performedAt"`
	IsDeload    bool                                  `json:"isDeload"`
	Exercises   []createWorkoutSessionExerciseRequest `json:"exercises"`
}

type createWorkoutSessionExerciseRequest struct {
	ExerciseID string             `json:"exerciseId"`
	Sets       []createSetRequest `json:"sets"`
}

type createSetRequest struct {
	SetNumber   int     `json:"setNumber"`
	Reps        int     `json:"reps"`
	DurationSec int     `json:"durationSec"`
	WeightKg    float64 `json:"weightKg"`
	AssistKg    float64 `json:"assistKg"`
}

// CreateWorkoutSession handles POST /v1/workout-sessions.
func (a *API) CreateWorkoutSession(w http.ResponseWriter, r *http.Request) {
	var req createWorkoutSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(req.Exercises) == 0 {
		writeError(w, http.StatusBadRequest, "exercises is required")
		return
	}

	in := repository.CreateWorkoutSessionInput{
		PerformedAt: req.PerformedAt,
		IsDeload:    req.IsDeload,
		Exercises:   make([]repository.CreateWorkoutSessionExerciseInput, 0, len(req.Exercises)),
	}
	for _, ex := range req.Exercises {
		if ex.ExerciseID == "" {
			writeError(w, http.StatusBadRequest, "exerciseId is required")
			return
		}
		if len(ex.Sets) == 0 {
			writeError(w, http.StatusBadRequest, "each exercise needs at least one set")
			return
		}
		sets := make([]repository.CreateSetInput, 0, len(ex.Sets))
		for _, s := range ex.Sets {
			if s.SetNumber < 1 || s.Reps < 0 || s.DurationSec < 0 || s.WeightKg < 0 || s.AssistKg < 0 {
				writeError(w, http.StatusBadRequest, "invalid set fields")
				return
			}
			sets = append(sets, repository.CreateSetInput{
				SetNumber:   s.SetNumber,
				Reps:        s.Reps,
				DurationSec: s.DurationSec,
				WeightKg:    s.WeightKg,
				AssistKg:    s.AssistKg,
			})
		}
		in.Exercises = append(in.Exercises, repository.CreateWorkoutSessionExerciseInput{
			ExerciseID: ex.ExerciseID,
			Sets:       sets,
		})
	}

	session, err := a.repo.CreateWorkoutSession(in)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workout session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// ListWorkoutSessions handles GET /v1/workout-sessions.
func (a *API) ListWorkoutSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := a.repo.ListWorkoutSessions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workout sessions")
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

// GetWorkoutSession handles GET /v1/workout-sessions/{id}.
func (a *API) GetWorkoutSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	session, err := a.repo.GetWorkoutSession(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get workout session")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

type startWorkoutSessionRequest struct {
	ID            string `json:"id"`
	PerformedAt   string `json:"performedAt"`
	IsDeload      bool   `json:"isDeload"`
	WorkoutPlanID string `json:"workoutPlanId"`
}

// StartWorkoutSession handles POST /v1/workout-sessions/start.
// Client must supply id (UUID). Replaying the same id is idempotent (200).
func (a *API) StartWorkoutSession(w http.ResponseWriter, r *http.Request) {
	var req startWorkoutSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if strings.TrimSpace(req.ID) == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	session, created, err := a.repo.StartWorkoutSession(req.ID, req.PerformedAt, req.IsDeload, req.WorkoutPlanID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout plan not found")
		return
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "id is required") || strings.Contains(msg, "id must be a valid UUID") {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to start workout session")
		return
	}
	if created {
		writeJSON(w, http.StatusCreated, session)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

type finishWorkoutSessionRequest struct {
	DurationSec int    `json:"durationSec"`
	CycleID     string `json:"cycleId"`
}

// FinishWorkoutSession handles POST /v1/workout-sessions/{id}/finish.
func (a *API) FinishWorkoutSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req finishWorkoutSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.DurationSec < 0 {
		writeError(w, http.StatusBadRequest, "durationSec must be >= 0")
		return
	}

	session, err := a.repo.FinishWorkoutSession(id, req.DurationSec, strings.TrimSpace(req.CycleID))
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, session)
}

type saveSessionExerciseRequest struct {
	Position int                `json:"position"`
	Sets     []createSetRequest `json:"sets"`
}

// SaveSessionExercise handles PUT /v1/workout-sessions/{id}/exercises/{exerciseId}.
func (a *API) SaveSessionExercise(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	exerciseID := r.PathValue("exerciseId")
	if sessionID == "" || exerciseID == "" {
		writeError(w, http.StatusBadRequest, "id and exerciseId are required")
		return
	}

	var req saveSessionExerciseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Position < 1 {
		writeError(w, http.StatusBadRequest, "position must be >= 1")
		return
	}
	if len(req.Sets) == 0 {
		writeError(w, http.StatusBadRequest, "sets is required")
		return
	}

	sets := make([]repository.CreateSetInput, 0, len(req.Sets))
	for _, s := range req.Sets {
		if s.SetNumber < 1 || s.Reps < 0 || s.DurationSec < 0 || s.WeightKg < 0 || s.AssistKg < 0 {
			writeError(w, http.StatusBadRequest, "invalid set fields")
			return
		}
		sets = append(sets, repository.CreateSetInput{
			SetNumber:   s.SetNumber,
			Reps:        s.Reps,
			DurationSec: s.DurationSec,
			WeightKg:    s.WeightKg,
			AssistKg:    s.AssistKg,
		})
	}

	block, err := a.repo.SaveSessionExercise(sessionID, req.Position, exerciseID, sets)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout session or exercise not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save exercise result")
		return
	}
	writeJSON(w, http.StatusOK, block)
}
