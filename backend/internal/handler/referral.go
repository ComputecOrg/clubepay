package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// MyReferralCode returns or creates the subscriber's referral code for a business.
// GET /api/my-referral-code
func (h *Handler) MyReferralCode(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	sub, err := h.Queries.GetActiveSubscriptionBySubscriber(r.Context(), subscriberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "assinatura não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinatura")
		return
	}

	biz, err := h.Queries.GetBusinessByID(r.Context(), sub.BusinessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	// Check if referral code already exists
	code, err := h.Queries.GetReferralCodeBySubscriberAndBusiness(r.Context(), repository.GetReferralCodeBySubscriberAndBusinessParams{
		ReferrerID: subscriberID,
		BusinessID: biz.ID,
	})
	if err == nil {
		// Code already exists
		writeJSON(w, http.StatusOK, map[string]string{"code": code})
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "erro ao buscar código de indicação")
		return
	}

	// Generate new 8-char hex code
	codeBytes := make([]byte, 4)
	if _, err := rand.Read(codeBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar código")
		return
	}
	newCode := hex.EncodeToString(codeBytes)

	// Create referral record with self-reference (template record)
	_, err = h.Queries.CreateReferral(r.Context(), repository.CreateReferralParams{
		ReferrerID: subscriberID,
		ReferredID: subscriberID, // self-reference for the template record
		BusinessID: biz.ID,
		Code:       newCode,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar código de indicação")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"code": newCode})
}

// ApplyReferral applies a referral code for the authenticated subscriber.
// POST /api/referrals/apply
func (h *Handler) ApplyReferral(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	var input struct {
		Code string `json:"code"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if input.Code == "" {
		writeError(w, http.StatusBadRequest, "code é obrigatório")
		return
	}

	// Find referral by code
	referral, err := h.Queries.GetReferralByCode(r.Context(), input.Code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "código de indicação não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar indicação")
		return
	}

	// Self-referral check
	if referral.ReferrerID == subscriberID {
		writeError(w, http.StatusBadRequest, "não é possível usar seu próprio código de indicação")
		return
	}

	// Check referral limit (max 3 per referrer per business)
	count, err := h.Queries.CountActiveReferralsByReferrer(r.Context(), repository.CountActiveReferralsByReferrerParams{
		ReferrerID: referral.ReferrerID,
		BusinessID: referral.BusinessID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao contar indicações")
		return
	}
	// The template record counts as 1, so real referrals are count - 1
	// But we count all active referrals including the template, so limit is referralLimit + 1
	if count > int64(domain.ReferralLimit) {
		writeError(w, http.StatusForbidden, "limite de indicações atingido (máximo 3)")
		return
	}

	// Create referral record for the referred subscriber
	_, err = h.Queries.CreateReferral(r.Context(), repository.CreateReferralParams{
		ReferrerID: referral.ReferrerID,
		ReferredID: subscriberID,
		BusinessID: referral.BusinessID,
		Code:       input.Code,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao aplicar indicação")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "applied",
		"message": "Indicação aplicada com sucesso! 10% de desconto para ambos.",
	})
}
