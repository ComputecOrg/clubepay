package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// GetProfile returns the authenticated user's profile.
// GET /api/profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "usuario nao encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar perfil")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone.String,
			"role":  user.Role,
		},
	})
}

// UpdateProfile updates name and phone.
// PUT /api/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var input struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "nome e obrigatorio")
		return
	}

	user, err := h.Queries.UpdateUserProfile(r.Context(), repository.UpdateUserProfileParams{
		ID:    userID,
		Name:  input.Name,
		Phone: pgText(input.Phone),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar perfil")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone.String,
			"role":  user.Role,
		},
	})
}

// ChangePassword changes the user's password.
// POST /api/profile/change-password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if len(input.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "nova senha deve ter no minimo 8 caracteres")
		return
	}

	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar usuario")
		return
	}

	if !domain.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "senha atual incorreta")
		return
	}

	hash, err := domain.HashPassword(input.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}

	if err := h.Queries.UpdateUserPassword(r.Context(), repository.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hash,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar senha")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "senha atualizada com sucesso"})
}
