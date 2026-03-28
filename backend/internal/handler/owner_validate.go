package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// ValidateUsageByOwner allows the owner to validate usage on behalf of a subscriber (fallback).
// POST /api/validate-usage-owner
func (h *Handler) ValidateUsageByOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	var input struct {
		SubscriberID int64 `json:"subscriber_id"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if input.SubscriberID <= 0 {
		writeError(w, http.StatusBadRequest, "subscriber_id e obrigatorio")
		return
	}

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negocio nao encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negocio")
		return
	}

	sub, err := h.Queries.GetActiveSubscription(r.Context(), repository.GetActiveSubscriptionParams{
		SubscriberID: input.SubscriberID,
		BusinessID:   biz.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "assinatura ativa nao encontrada para este assinante")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinatura")
		return
	}

	plan, err := h.Queries.GetPlanByID(r.Context(), sub.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar plano")
		return
	}

	now := time.Now()
	var currentCount int64

	if plan.LimitType == domain.LimitTypeDaily {
		start, end := domain.DailyRange(now)
		currentCount, err = h.Queries.CountDailyUsage(r.Context(), repository.CountDailyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgTimestamptz(start),
			Column3:        pgTimestamptz(end),
		})
	} else {
		start, end := domain.MonthlyRange(now)
		currentCount, err = h.Queries.CountMonthlyUsage(r.Context(), repository.CountMonthlyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgTimestamptz(start),
			Column3:        pgTimestamptz(end),
		})
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao contar usos")
		return
	}

	if err := domain.CheckUsageLimit(plan.LimitType, int(plan.LimitCount), currentCount); err != nil {
		writeError(w, http.StatusForbidden, "limite de uso atingido para este periodo")
		return
	}

	_, err = h.Queries.CreateUsage(r.Context(), sub.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao registrar uso")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "validated",
		"used":      currentCount + 1,
		"limit":     plan.LimitCount,
		"plan_name": plan.Name,
	})
}
