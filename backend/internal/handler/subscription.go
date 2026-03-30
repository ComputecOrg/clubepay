package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
)

const freeTierSubscriberLimit = 15

// Subscribe creates a new subscription for the authenticated subscriber.
// POST /api/subscribe
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	subscriberID := middleware.UserIDFromContext(r.Context())

	var input struct {
		PlanID int64 `json:"plan_id"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if input.PlanID <= 0 {
		writeError(w, http.StatusBadRequest, "plan_id é obrigatório")
		return
	}

	// Fetch plan and verify it exists and is active
	plan, err := h.Queries.GetPlanByID(r.Context(), input.PlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "plano não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar plano")
		return
	}
	if !plan.Active {
		writeError(w, http.StatusBadRequest, "plano não está ativo")
		return
	}

	// Check if subscriber already has an active subscription to this business
	_, err = h.Queries.GetActiveSubscription(r.Context(), repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   plan.BusinessID,
	})
	if err == nil {
		writeError(w, http.StatusConflict, "já possui assinatura ativa neste negócio")
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "erro ao verificar assinatura existente")
		return
	}

	// Check free tier subscriber limit
	count, err := h.Queries.CountActiveSubscriptionsByBusinessID(r.Context(), plan.BusinessID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao contar assinantes")
		return
	}
	if count >= int64(freeTierSubscriberLimit) {
		writeError(w, http.StatusForbidden, "limite de assinantes atingido no plano gratuito")
		return
	}

	// Get subscriber data for PSP
	subscriber, err := h.Queries.GetUserByID(r.Context(), subscriberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar dados do assinante")
		return
	}

	// Create PSP customer
	customer, err := h.PSP.CreateCustomer(r.Context(), psp.CreateCustomerRequest{
		Name:  subscriber.Name,
		Email: subscriber.Email,
		Phone: subscriber.Phone.String,
	})
	if err != nil {
		slog.Error("failed to create PSP customer", "error", err)
		writeError(w, http.StatusInternalServerError, "erro ao criar cliente no PSP")
		return
	}

	// Check for active referral discount
	discountPercent := int32(0)
	referral, refErr := h.Queries.GetReferralByReferredAndBusiness(r.Context(), repository.GetReferralByReferredAndBusinessParams{
		ReferredID: subscriberID,
		BusinessID: plan.BusinessID,
	})
	if refErr == nil {
		discountPercent = 10
	}

	// Calculate discounted price
	priceCents := plan.PriceCents
	if discountPercent > 0 {
		priceCents = priceCents * int64(100-discountPercent) / 100
	}

	// Create PSP subscription
	pspSub, err := h.PSP.CreateSubscription(r.Context(), psp.CreateSubscriptionRequest{
		CustomerID:  customer.ID,
		PriceCents:  priceCents,
		Description: plan.Name,
		Cycle:       "MONTHLY",
	})
	if err != nil {
		slog.Error("failed to create PSP subscription", "error", err)
		writeError(w, http.StatusInternalServerError, "erro ao criar assinatura no PSP")
		return
	}

	// Create DB subscription
	periodEnd := time.Now().AddDate(0, 1, 0)
	referredBy := pgtype.Int8{}
	if refErr == nil {
		referredBy = pgtype.Int8{Int64: referral.ReferrerID, Valid: true}
	}
	sub, err := h.Queries.CreateSubscriptionWithDiscount(r.Context(), repository.CreateSubscriptionWithDiscountParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriberID,
		BusinessID:        plan.BusinessID,
		PspSubscriptionID: pgText(pspSub.ID),
		Status:            "active",
		PeriodEnd:         pgTimestamptz(periodEnd),
		ReferredBy:        referredBy,
		DiscountPercent:   discountPercent,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar assinatura")
		return
	}

	// Send welcome email
	go func() {
		biz, bizErr := h.Queries.GetBusinessByID(context.Background(), plan.BusinessID)
		if bizErr == nil {
			subject, body := email.WelcomeEmail(subscriber.Name, plan.Name, biz.Name)
			h.Email.Send(subscriber.Email, subject, body)
		}
	}()

	writeJSON(w, http.StatusCreated, sub)
}

// ListSubscriptions lists all subscriptions for the authenticated owner's business.
// GET /api/subscriptions
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
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

	subs, err := h.Queries.ListSubscriptionsByBusinessID(r.Context(), biz.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar assinaturas")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"subscriptions": subs})
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

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	sub, err := h.Queries.GetSubscriptionByID(r.Context(), subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "assinatura não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinatura")
		return
	}

	if sub.BusinessID != biz.ID {
		writeError(w, http.StatusForbidden, "assinatura não pertence ao seu negócio")
		return
	}

	// Cancel in PSP (best effort)
	if sub.PspSubscriptionID.Valid && sub.PspSubscriptionID.String != "" {
		if err := h.PSP.CancelSubscription(r.Context(), sub.PspSubscriptionID.String); err != nil {
			slog.Error("failed to cancel PSP subscription", "error", err, "psp_id", sub.PspSubscriptionID.String)
		}
	}

	// Cancel in DB
	if err := h.Queries.CancelSubscription(r.Context(), subID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao cancelar assinatura")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
