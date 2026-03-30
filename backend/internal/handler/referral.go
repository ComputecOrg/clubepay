package handler

import (
	"net/http"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
)

// MyReferralCode returns or creates the subscriber's referral code for a business.
// GET /api/my-referral-code
func (h *Handler) MyReferralCode(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	resp, err := h.Referrals.GetOrCreateCode(r.Context(), subscriberID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ApplyReferral applies a referral code for the authenticated subscriber.
// POST /api/referrals/apply
func (h *Handler) ApplyReferral(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	var input domain.ApplyReferralInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := domain.Validate(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Referrals.Apply(r.Context(), subscriberID, input); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "applied",
		"message": "Indicação aplicada com sucesso! 10% de desconto para ambos.",
	})
}
