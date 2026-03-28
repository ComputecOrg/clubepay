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

// ValidateUsage validates a usage for the authenticated subscriber.
// POST /api/validate-usage
func (h *Handler) ValidateUsage(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	var input struct {
		BusinessSlug string `json:"business_slug"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if input.BusinessSlug == "" {
		writeError(w, http.StatusBadRequest, "business_slug é obrigatório")
		return
	}

	// Find business by slug
	biz, err := h.Queries.GetBusinessBySlug(r.Context(), input.BusinessSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	// Find active subscription
	sub, err := h.Queries.GetActiveSubscription(r.Context(), repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   biz.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "assinatura não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinatura")
		return
	}

	// Get plan
	plan, err := h.Queries.GetPlanByID(r.Context(), sub.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar plano")
		return
	}

	// Count current usage for the period
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

	// Check usage limit
	if err := domain.CheckUsageLimit(plan.LimitType, int(plan.LimitCount), currentCount); err != nil {
		writeError(w, http.StatusForbidden, "limite de uso atingido para este período")
		return
	}

	// Create usage record
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

// MyUsage returns usage information for the authenticated subscriber.
// GET /api/my-usage?business_slug=xxx
func (h *Handler) MyUsage(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	businessSlug := r.URL.Query().Get("business_slug")
	if businessSlug == "" {
		writeError(w, http.StatusBadRequest, "business_slug é obrigatório")
		return
	}

	biz, err := h.Queries.GetBusinessBySlug(r.Context(), businessSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	sub, err := h.Queries.GetActiveSubscription(r.Context(), repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   biz.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "assinatura não encontrada")
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
	var start, end time.Time
	if plan.LimitType == domain.LimitTypeDaily {
		start, end = domain.DailyRange(now)
	} else {
		start, end = domain.MonthlyRange(now)
	}

	usages, err := h.Queries.ListUsagesBySubscription(r.Context(), repository.ListUsagesBySubscriptionParams{
		SubscriptionID: sub.ID,
		Column2:        pgTimestamptz(start),
		Column3:        pgTimestamptz(end),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar usos")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"used":      len(usages),
		"limit":     plan.LimitCount,
		"plan_name": plan.Name,
		"usages":    usages,
	})
}
