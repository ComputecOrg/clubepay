package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/repository"
)

// webhookPayload represents the Asaas webhook JSON body.
type webhookPayload struct {
	Event   string         `json:"event"`
	Payment webhookPayment `json:"payment"`
}

type webhookPayment struct {
	Subscription string `json:"subscription"`
	Status       string `json:"status"`
}

// PSPWebhook handles incoming PSP (Asaas) webhook events.
// POST /api/psp/webhook
func (h *Handler) PSPWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}
	defer r.Body.Close()

	// Validate HMAC signature if webhook secret is configured
	if h.Config.AsaasWebhookSecret != "" {
		signature := r.Header.Get("asaas-access-token")
		if !h.PSP.ValidateWebhook(body, signature) {
			writeError(w, http.StatusUnauthorized, "invalid webhook signature")
			return
		}
	}

	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	slog.Info("webhook received",
		"event", payload.Event,
		"psp_subscription_id", payload.Payment.Subscription,
		"payment_status", payload.Payment.Status,
	)

	if payload.Payment.Subscription == "" {
		// Not a subscription-related event, ignore
		w.WriteHeader(http.StatusOK)
		return
	}

	// Lookup subscription by PSP ID
	sub, err := h.Queries.GetSubscriptionByPSPID(r.Context(), pgtype.Text{
		String: payload.Payment.Subscription,
		Valid:  true,
	})
	if err != nil {
		slog.Error("webhook: subscription not found", "psp_id", payload.Payment.Subscription, "error", err)
		// Return 200 to avoid retries for unknown subscriptions
		w.WriteHeader(http.StatusOK)
		return
	}

	switch payload.Event {
	case "PAYMENT_CONFIRMED", "PAYMENT_RECEIVED":
		// Reactivate subscription if it was in grace or blocked
		if sub.Status == "grace" || sub.Status == "blocked" {
			if err := h.Queries.UpdateSubscriptionStatus(r.Context(), repository.UpdateSubscriptionStatusParams{
				ID:     sub.ID,
				Status: "active",
			}); err != nil {
				slog.Error("webhook: failed to reactivate subscription", "id", sub.ID, "error", err)
				writeError(w, http.StatusInternalServerError, "failed to update subscription")
				return
			}
			slog.Info("webhook processed", "event", payload.Event, "subscription_id", sub.ID, "action", "reactivated")
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			subscriber, subErr := h.Queries.GetUserByID(ctx, sub.SubscriberID)
			plan, planErr := h.Queries.GetPlanByID(ctx, sub.PlanID)
			if subErr == nil && planErr == nil {
				amount := fmt.Sprintf("R$ %.2f", float64(plan.PriceCents)/100)
				subject, body := email.PaymentConfirmedEmail(subscriber.Name, plan.Name, amount)
				if err := h.Email.Send(subscriber.Email, subject, body); err != nil {
					slog.Error("failed to send payment confirmed email", "error", err, "to", subscriber.Email)
				}
			}
		}()

	case "PAYMENT_OVERDUE":
		// Set subscription to grace with 3-day deadline
		graceDeadline := time.Now().AddDate(0, 0, domain.GracePeriodDays)
		if err := h.Queries.UpdateSubscriptionGrace(r.Context(), repository.UpdateSubscriptionGraceParams{
			ID:            sub.ID,
			GraceDeadline: pgtype.Timestamptz{Time: graceDeadline, Valid: true},
		}); err != nil {
			slog.Error("webhook: failed to set grace", "id", sub.ID, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to update subscription")
			return
		}
		slog.Info("webhook processed", "event", payload.Event, "subscription_id", sub.ID, "action", "grace_set")

	case "PAYMENT_DELETED", "PAYMENT_REFUNDED":
		// Cancel subscription
		if err := h.Queries.CancelSubscription(r.Context(), sub.ID); err != nil {
			slog.Error("webhook: failed to cancel subscription", "id", sub.ID, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to cancel subscription")
			return
		}
		slog.Info("webhook processed", "event", payload.Event, "subscription_id", sub.ID, "action", "cancelled")
	}

	w.WriteHeader(http.StatusOK)
}
