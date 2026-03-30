package handler

import (
	"net/http"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
)

// ValidateUsage validates a usage for the authenticated subscriber.
// POST /api/validate-usage
func (h *Handler) ValidateUsage(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	var input domain.ValidateUsageInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := domain.Validate(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Usage.Validate(r.Context(), subscriberID, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// MyUsage returns usage information for the authenticated subscriber.
// GET /api/my-usage?business_slug=xxx
func (h *Handler) MyUsage(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	businessSlug := r.URL.Query().Get("business_slug")
	if businessSlug == "" {
		writeError(w, http.StatusBadRequest, "business_slug é obrigatório")
		return
	}

	resp, err := h.Usage.GetMyUsage(r.Context(), subscriberID, businessSlug)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
