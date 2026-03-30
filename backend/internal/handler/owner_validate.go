package handler

import (
	"net/http"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
)

// ValidateUsageByOwner allows the owner to validate usage on behalf of a subscriber (fallback).
// POST /api/validate-usage-owner
func (h *Handler) ValidateUsageByOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	var input domain.ValidateUsageOwnerInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if err := domain.Validate(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Usage.ValidateByOwner(r.Context(), ownerID, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
