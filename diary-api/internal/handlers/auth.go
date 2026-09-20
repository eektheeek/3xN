package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/auth"
)

type ctxSessionKey int

const sessionIDKey ctxSessionKey = 1

func withSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

func sessionIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionIDKey).(string)
	return v
}

func userIDFromRequest(r *http.Request) string {
	return auth.UserIDFromContext(r.Context())
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, prefix))
}

// RequireAuth rejects requests without a valid Bearer session.
func (a *API) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}
		userID, sessionID, err := a.repo.UserIDByTokenHash(auth.HashToken(token))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		ctx := auth.WithUserID(r.Context(), userID)
		ctx = withSessionID(ctx, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) issueSession(w http.ResponseWriter, status int, userID, email, createdAt string) {
	raw, hash, err := auth.NewSessionToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	expires := time.Now().UTC().Add(time.Duration(auth.SessionTTLDays) * 24 * time.Hour)
	if _, err := a.repo.CreateSession(userID, hash, expires); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	writeJSON(w, status, authResponse{
		Token: raw,
		User: authUser{
			ID:        userID,
			Email:     email,
			CreatedAt: createdAt,
		},
	})
}
