package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

type createWorkoutPlanRequest struct {
	Name        string   `json:"name"`
	ExerciseIDs []string `json:"exerciseIds"`
}

// CreateWorkoutPlan handles POST /v1/workout-plans.
func (a *API) CreateWorkoutPlan(w http.ResponseWriter, r *http.Request) {
	var req createWorkoutPlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректное тело запроса")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "укажите название тренировки")
		return
	}
	if len(req.ExerciseIDs) == 0 {
		writeError(w, http.StatusBadRequest, "добавьте хотя бы одно упражнение")
		return
	}

	plan, err := a.repo.CreateWorkoutPlan(repository.CreateWorkoutPlanInput{
		Name:        req.Name,
		ExerciseIDs: req.ExerciseIDs,
	})
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "упражнение не найдено")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

// ListWorkoutPlans handles GET /v1/workout-plans.
func (a *API) ListWorkoutPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := a.repo.ListWorkoutPlans()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось загрузить тренировки")
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

// GetWorkoutPlan handles GET /v1/workout-plans/{id}.
func (a *API) GetWorkoutPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "не указан id")
		return
	}
	plan, err := a.repo.GetWorkoutPlan(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "тренировка не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось загрузить тренировку")
		return
	}
	writeJSON(w, http.StatusOK, plan)
}
