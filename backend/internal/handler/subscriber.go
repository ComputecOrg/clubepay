package handler

import (
	"net/http"

	"github.com/clubepay/backend/internal/middleware"
)

// MyPlan returns the plan information for the authenticated subscriber.
// GET /api/my-plan
func (h *Handler) MyPlan(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	resp, err := h.Subscriptions.GetMyPlan(r.Context(), subscriberID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CancelBySubscriber allows the subscriber to cancel their own subscription.
// POST /api/cancel
func (h *Handler) CancelBySubscriber(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	resp, err := h.Subscriptions.CancelBySubscriber(r.Context(), subscriberID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
