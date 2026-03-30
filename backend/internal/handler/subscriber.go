package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// MyPlan returns the plan information for the authenticated subscriber.
// GET /api/my-plan?business_slug=xxx
func (h *Handler) MyPlan(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plan": map[string]interface{}{
			"id":          plan.ID,
			"name":        plan.Name,
			"description": plan.Description,
			"price_cents": plan.PriceCents,
			"limit_type":  plan.LimitType,
			"limit_count": plan.LimitCount,
		},
		"business": map[string]interface{}{
			"id":   biz.ID,
			"name": biz.Name,
			"slug": biz.Slug,
		},
		"subscription": map[string]interface{}{
			"id":         sub.ID,
			"status":     sub.Status,
			"period_end": sub.PeriodEnd,
		},
	})
}

// CancelBySubscriber allows the subscriber to cancel their own subscription.
// POST /api/cancel
func (h *Handler) CancelBySubscriber(w http.ResponseWriter, r *http.Request) {
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

	biz, err := h.Queries.GetBusinessBySlug(r.Context(), input.BusinessSlug)
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

	// Cancel in PSP (log error but don't fail)
	if sub.PspSubscriptionID.Valid && sub.PspSubscriptionID.String != "" {
		if err := h.PSP.CancelSubscription(r.Context(), sub.PspSubscriptionID.String); err != nil {
			slog.Error("failed to cancel PSP subscription", "error", err, "psp_id", sub.PspSubscriptionID.String)
		}
	}

	// Cancel in DB
	if err := h.Queries.CancelSubscription(r.Context(), sub.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao cancelar assinatura")
		return
	}

	go func() {
		user, userErr := h.Queries.GetUserByID(context.Background(), subscriberID)
		plan, planErr := h.Queries.GetPlanByID(context.Background(), sub.PlanID)
		if userErr == nil && planErr == nil {
			validUntil := ""
			if sub.PeriodEnd.Valid {
				validUntil = sub.PeriodEnd.Time.Format("02/01/2006")
			}
			subject, body := email.SubscriptionCancelledEmail(user.Name, plan.Name, validUntil)
			h.Email.Send(user.Email, subject, body)
		}
	}()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "cancelled",
		"period_end": sub.PeriodEnd,
		"message":    "Assinatura cancelada. Acesso continua até o fim do período pago.",
	})
}
