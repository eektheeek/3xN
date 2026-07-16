package handlers

import (
	"errors"
	"net/http"

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
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	WeightKg  float64 `json:"weightKg"`
	AssistKg  float64 `json:"assistKg"`
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
			if s.SetNumber < 1 || s.Reps < 0 || s.WeightKg < 0 || s.AssistKg < 0 {
				writeError(w, http.StatusBadRequest, "invalid set fields")
				return
			}
			sets = append(sets, repository.CreateSetInput{
				SetNumber: s.SetNumber,
				Reps:      s.Reps,
				WeightKg:  s.WeightKg,
				AssistKg:  s.AssistKg,
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
