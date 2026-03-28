package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// GetBusiness returns the business owned by the authenticated user.
func (h *Handler) GetBusiness(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	writeJSON(w, http.StatusOK, biz)
}

// UpdateBusiness updates editable fields of the authenticated owner's business.
func (h *Handler) UpdateBusiness(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	// Retrieve existing business to get its ID
	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	var input struct {
		Name    string `json:"name"`
		Segment string `json:"segment"`
		Address string `json:"address"`
		LogoUrl string `json:"logo_url"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}

	updated, err := h.Queries.UpdateBusiness(r.Context(), repository.UpdateBusinessParams{
		ID:      biz.ID,
		Name:    input.Name,
		Segment: input.Segment,
		Address: pgText(input.Address),
		LogoUrl: pgText(input.LogoUrl),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar negócio")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}
