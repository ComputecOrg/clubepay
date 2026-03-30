package handler

import (
	"net/http"

	"github.com/clubepay/backend/internal/domain"
)

// RegisterOwner creates an owner user and a business, then returns a JWT.
// POST /api/auth/register
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterOwnerInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := domain.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Auth.RegisterOwner(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login authenticates a user and returns a JWT.
// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := domain.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Auth.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// RegisterSubscriber creates a subscriber user and returns a long-lived JWT (30 days).
// POST /api/auth/register-subscriber
func (h *Handler) RegisterSubscriber(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterSubscriberInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := domain.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Auth.RegisterSubscriber(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// RequestPasswordReset sends a password reset email.
// POST /api/auth/request-password-reset
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req domain.RequestPasswordResetInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always return 200 to avoid leaking user existence.
	writeJSON(w, http.StatusOK, map[string]string{"message": "if an account with that email exists, a reset link has been sent"})

	h.Auth.RequestPasswordReset(r.Context(), req.Email)
}

// ConfirmPasswordReset resets the user's password using a valid token.
// POST /api/auth/confirm-password-reset
func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req domain.ConfirmPasswordResetInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := domain.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Auth.ConfirmPasswordReset(r.Context(), req); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
}
