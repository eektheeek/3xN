package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/auth"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

type authCredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type authUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type authResponse struct {
	Token string   `json:"token"`
	User  authUser `json:"user"`
}

// Register handles POST /v1/auth/register.
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var req authCredentialsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	email := auth.NormalizeEmail(req.Email)
	if email == "" || !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	user, err := a.repo.CreateUser(email, hash)
	if errors.Is(err, repository.ErrConflict) {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}
	a.issueSession(w, http.StatusCreated, user.ID, user.Email, user.CreatedAt)
}

// Login handles POST /v1/auth/login.
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req authCredentialsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	email := auth.NormalizeEmail(req.Email)
	user, err := a.repo.GetUserByEmail(email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	a.issueSession(w, http.StatusOK, user.ID, user.Email, user.CreatedAt)
}

// Logout handles POST /v1/auth/logout.
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := sessionIDFromContext(r.Context())
	if sessionID != "" {
		_ = a.repo.DeleteSession(sessionID)
	}
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /v1/auth/me.
func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	user, err := a.repo.GetUserByID(userIDFromRequest(r))
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	writeJSON(w, http.StatusOK, user.Public())
}

// ChangePassword handles PUT /v1/auth/password.
func (a *API) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID := userIDFromRequest(r)
	user, err := a.repo.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		writeError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	if err := a.repo.UpdateUserPasswordHash(userID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update password")
		return
	}
	if keep := sessionIDFromContext(r.Context()); keep != "" {
		_ = a.repo.DeleteOtherSessionsForUser(userID, keep)
	}
	w.WriteHeader(http.StatusNoContent)
}
