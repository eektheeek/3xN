package handlers

import "net/http"

func authWrap(api *API, h http.HandlerFunc) http.Handler {
	return api.RequireAuth(h)
}

// NewMux registers all HTTP routes.
func NewMux(api *API) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("POST /v1/auth/register", api.Register)
	mux.HandleFunc("POST /v1/auth/login", api.Login)
	mux.Handle("POST /v1/auth/logout", authWrap(api, api.Logout))
	mux.Handle("GET /v1/auth/me", authWrap(api, api.Me))
	mux.Handle("PUT /v1/auth/password", authWrap(api, api.ChangePassword))

	mux.Handle("POST /v1/exercises", authWrap(api, api.CreateExercise))
	mux.Handle("GET /v1/exercises", authWrap(api, api.ListExercises))
	mux.Handle("GET /v1/exercises/{id}", authWrap(api, api.GetExercise))
	mux.Handle("GET /v1/exercises/{id}/stats", authWrap(api, api.GetExerciseStats))
	mux.Handle("PUT /v1/exercises/{id}", authWrap(api, api.UpdateExercise))
	mux.Handle("PUT /v1/exercises/{id}/target", authWrap(api, api.SetTarget))

	mux.Handle("POST /v1/interval-protocols", authWrap(api, api.CreateIntervalProtocol))
	mux.Handle("GET /v1/interval-protocols", authWrap(api, api.ListIntervalProtocols))
	mux.Handle("GET /v1/interval-protocols/{id}", authWrap(api, api.GetIntervalProtocol))
	mux.Handle("PUT /v1/interval-protocols/{id}", authWrap(api, api.UpdateIntervalProtocol))
	mux.Handle("DELETE /v1/interval-protocols/{id}", authWrap(api, api.DeleteIntervalProtocol))

	mux.Handle("POST /v1/workout-sessions", authWrap(api, api.CreateWorkoutSession))
	mux.Handle("POST /v1/workout-sessions/start", authWrap(api, api.StartWorkoutSession))
	mux.Handle("GET /v1/workout-sessions", authWrap(api, api.ListWorkoutSessions))
	mux.Handle("GET /v1/workout-sessions/{id}", authWrap(api, api.GetWorkoutSession))
	mux.Handle("POST /v1/workout-sessions/{id}/finish", authWrap(api, api.FinishWorkoutSession))
	mux.Handle("PUT /v1/workout-sessions/{id}/exercises/{exerciseId}", authWrap(api, api.SaveSessionExercise))

	mux.Handle("POST /v1/workout-plans", authWrap(api, api.CreateWorkoutPlan))
	mux.Handle("GET /v1/workout-plans", authWrap(api, api.ListWorkoutPlans))
	mux.Handle("GET /v1/workout-plans/{id}", authWrap(api, api.GetWorkoutPlan))

	mux.Handle("GET /v1/cycles", authWrap(api, api.ListCycles))
	mux.Handle("POST /v1/cycles", authWrap(api, api.CreateCycle))
	mux.Handle("GET /v1/cycles/{id}", authWrap(api, api.GetCycle))
	mux.Handle("PUT /v1/cycles/{id}/steps", authWrap(api, api.ReplaceCycleSteps))
	mux.Handle("POST /v1/cycles/{id}/advance", authWrap(api, api.AdvanceCycle))
	mux.Handle("POST /v1/cycles/{id}/restart", authWrap(api, api.RestartCycle))
	mux.Handle("POST /v1/cycles/{id}/repeat", authWrap(api, api.RepeatCycle))
	mux.Handle("PUT /v1/cycles/{id}/on-home", authWrap(api, api.SetCycleOnHome))

	return mux
}
