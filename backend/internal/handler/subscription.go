package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/service"
)

// Subscribe creates a new subscription for the authenticated subscriber.
// POST /api/subscribe
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	var input domain.SubscribeInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := domain.Validate(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Sync PSP reference so that test overrides of h.PSP propagate to the service.
	h.Subscriptions.PSP = h.PSP

	sub, err := h.Subscriptions.Subscribe(r.Context(), subscriberID, service.SubscribeInput{
		PlanID: input.PlanID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

// ListSubscriptions lists all subscriptions for the authenticated owner's business.
// GET /api/subscriptions
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	resp, err := h.Subscriptions.ListByBusiness(r.Context(), ownerID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CancelSubscriptionByOwner cancels a subscription by the owner.
// DELETE /api/subscriptions/:id
func (h *Handler) CancelSubscriptionByOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := h.Subscriptions.CancelByOwner(r.Context(), ownerID, subID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
