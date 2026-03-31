package handler

import (
	"log/slog"
	"net/http"

	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/repository"
)

// Reconcile performs daily reconciliation: blocks grace-expired subscriptions and syncs pending ones with PSP.
// POST /api/cron/reconcile
func (h *Handler) Reconcile(w http.ResponseWriter, r *http.Request) {
	// Validate cron secret
	secret := r.Header.Get("X-Cron-Secret")
	if h.Config.CronSecret == "" || secret != h.Config.CronSecret {
		writeError(w, http.StatusUnauthorized, "invalid cron secret")
		return
	}

	ctx := r.Context()
	blockedCount := 0
	syncedCount := 0

	// 1. Block grace-expired subscriptions
	graceExpired, err := h.Queries.ListGraceExpiredSubscriptions(ctx)
	if err != nil {
		slog.Error("reconcile: failed to list grace-expired subscriptions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list grace-expired subscriptions")
		return
	}

	for _, sub := range graceExpired {
		if err := h.Queries.UpdateSubscriptionStatus(ctx, repository.UpdateSubscriptionStatusParams{
			ID:     sub.ID,
			Status: "blocked",
		}); err != nil {
			slog.Error("reconcile: failed to block subscription", "id", sub.ID, "error", err)
			continue
		}
		blockedCount++

		if h.Email != nil {
			subscriber, err := h.Queries.GetUserByID(ctx, sub.SubscriberID)
			if err == nil {
				subject, body := email.GraceBlockedEmail(subscriber.Name)
				if err := h.Email.Send(subscriber.Email, subject, body); err != nil {
					slog.Error("failed to send reconcile email", "error", err, "to", subscriber.Email)
				}
			}
		}
	}

	// 2. Sync pending subscriptions with PSP
	pending, err := h.Queries.ListPendingSubscriptions(ctx)
	if err != nil {
		slog.Error("reconcile: failed to list pending subscriptions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list pending subscriptions")
		return
	}

	for _, sub := range pending {
		if !sub.PspSubscriptionID.Valid || sub.PspSubscriptionID.String == "" {
			continue
		}

		pspSub, err := h.PSP.GetSubscription(ctx, sub.PspSubscriptionID.String)
		if err != nil {
			slog.Error("reconcile: failed to get PSP subscription", "psp_id", sub.PspSubscriptionID.String, "error", err)
			continue
		}

		var newStatus string
		switch pspSub.Status {
		case "ACTIVE":
			newStatus = "active"
		case "EXPIRED", "INACTIVE":
			newStatus = "cancelled"
		default:
			continue
		}

		if err := h.Queries.UpdateSubscriptionStatus(ctx, repository.UpdateSubscriptionStatusParams{
			ID:     sub.ID,
			Status: newStatus,
		}); err != nil {
			slog.Error("reconcile: failed to update subscription status", "id", sub.ID, "error", err)
			continue
		}
		syncedCount++
	}

	// 3. Send spending alerts
	if err := h.SendSpendingAlerts(ctx); err != nil {
		slog.Error("reconcile: failed to send spending alerts", "error", err)
		// Don't fail the whole reconcile, just log the error
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"blocked": blockedCount,
		"synced":  syncedCount,
	})
}
