package handlers

import "net/http"

// NewMux registers all HTTP routes.
func NewMux(api *API) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("POST /v1/exercises", api.CreateExercise)
	mux.HandleFunc("GET /v1/exercises", api.ListExercises)
	mux.HandleFunc("GET /v1/exercises/{id}", api.GetExercise)
	mux.HandleFunc("PUT /v1/exercises/{id}", api.UpdateExercise)
	mux.HandleFunc("PUT /v1/exercises/{id}/target", api.SetTarget)

	mux.HandleFunc("POST /v1/workout-sessions", api.CreateWorkoutSession)
	mux.HandleFunc("POST /v1/workout-sessions/start", api.StartWorkoutSession)
	mux.HandleFunc("GET /v1/workout-sessions", api.ListWorkoutSessions)
	mux.HandleFunc("GET /v1/workout-sessions/{id}", api.GetWorkoutSession)
	mux.HandleFunc("POST /v1/workout-sessions/{id}/finish", api.FinishWorkoutSession)
	mux.HandleFunc("PUT /v1/workout-sessions/{id}/exercises/{exerciseId}", api.SaveSessionExercise)

	mux.HandleFunc("POST /v1/workout-plans", api.CreateWorkoutPlan)
	mux.HandleFunc("GET /v1/workout-plans", api.ListWorkoutPlans)
	mux.HandleFunc("GET /v1/workout-plans/{id}", api.GetWorkoutPlan)

	return mux
}
