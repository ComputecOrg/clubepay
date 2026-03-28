package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
)

func TestWebhook_PaymentConfirmed(t *testing.T) {
	h := setupHandler(t)

	// Seed owner, business, plan, subscriber
	owner := testutil.SeedOwner(t, h.Queries, "whowner@test.com", "WH Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "WH Cafe", "wh-cafe")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "WH Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "whsub@test.com", "WH Sub", "11999990000")

	// Create subscription in grace status directly
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_wh_123", Valid: true},
		Status:            "grace",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Send PAYMENT_CONFIRMED webhook
	payload := map[string]interface{}{
		"event": "PAYMENT_CONFIRMED",
		"payment": map[string]interface{}{
			"subscription": "sub_wh_123",
			"status":       "CONFIRMED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify subscription was reactivated
	updated, err := h.Queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", updated.Status)
}

func TestWebhook_PaymentOverdue(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "whoverdue@test.com", "WH Overdue Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "WH Overdue Cafe", "wh-overdue-cafe")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "WH Overdue Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "whoverduesub@test.com", "WH Overdue Sub", "11999990001")

	// Create active subscription
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_overdue_123", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Send PAYMENT_OVERDUE webhook
	payload := map[string]interface{}{
		"event": "PAYMENT_OVERDUE",
		"payment": map[string]interface{}{
			"subscription": "sub_overdue_123",
			"status":       "OVERDUE",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify subscription is now in grace with a deadline
	updated, err := h.Queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "grace", updated.Status)
	assert.True(t, updated.GraceDeadline.Valid)
	// Grace deadline should be approximately 3 days from now
	assert.WithinDuration(t, time.Now().AddDate(0, 0, 3), updated.GraceDeadline.Time, 10*time.Second)
}

func TestWebhook_InvalidJSON(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestWebhook_UnknownSubscription(t *testing.T) {
	h := setupHandler(t)

	// Payload with a PSP subscription ID that doesn't exist in the DB
	payload := map[string]interface{}{
		"event": "PAYMENT_CONFIRMED",
		"payment": map[string]interface{}{
			"subscription": "sub_unknown_99999",
			"status":       "CONFIRMED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	// Should return 200 — don't retry unknown subscriptions
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWebhook_PaymentDeleted(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "whdel@test.com", "WH Delete Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "WH Delete Cafe", "wh-delete-cafe")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "WH Delete Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "whdelete@test.com", "WH Delete Sub", "11999990003")

	// Create active subscription
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_delete_123", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Send PAYMENT_DELETED webhook
	payload := map[string]interface{}{
		"event": "PAYMENT_DELETED",
		"payment": map[string]interface{}{
			"subscription": "sub_delete_123",
			"status":       "DELETED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify subscription was cancelled
	updated, err := h.Queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", updated.Status)
}

func TestWebhook_NoSubscriptionField(t *testing.T) {
	h := setupHandler(t)

	// Payload without a subscription field (e.g. a payment not linked to a subscription)
	payload := map[string]interface{}{
		"event": "PAYMENT_CONFIRMED",
		"payment": map[string]interface{}{
			"subscription": "",
			"status":       "CONFIRMED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	// Should return 200 — ignore non-subscription events
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWebhook_PaymentRefunded(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "whrefund@test.com", "WH Refund Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "WH Refund Cafe", "wh-refund-cafe")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "WH Refund Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "whrefundsub@test.com", "WH Refund Sub", "11999990006")

	// Create active subscription
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_refund_123", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Send PAYMENT_REFUNDED webhook
	payload := map[string]interface{}{
		"event": "PAYMENT_REFUNDED",
		"payment": map[string]interface{}{
			"subscription": "sub_refund_123",
			"status":       "REFUNDED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify subscription was cancelled
	updated, err := h.Queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", updated.Status)
}

func TestWebhook_PaymentReceived(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "whreceived@test.com", "WH Received Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "WH Received Cafe", "wh-received-cafe")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "WH Received Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "whreceivedsub@test.com", "WH Received Sub", "11999990007")

	// Create subscription in blocked status
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_received_123", Valid: true},
		Status:            "blocked",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Send PAYMENT_RECEIVED webhook — should reactivate if blocked
	payload := map[string]interface{}{
		"event": "PAYMENT_RECEIVED",
		"payment": map[string]interface{}{
			"subscription": "sub_received_123",
			"status":       "RECEIVED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify subscription was reactivated
	updated, err := h.Queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", updated.Status)
}

func TestWebhook_InvalidSignature(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:          "test-secret-key",
		AsaasWebhookSecret: "my-webhook-secret",
	}
	mockPSP := &psp.MockPSP{
		ValidateWebhookFn: func(payload []byte, signature string) bool {
			return false // Always reject
		},
	}
	h := handler.New(queries, cfg, mockPSP)

	payload := map[string]interface{}{
		"event": "PAYMENT_CONFIRMED",
		"payment": map[string]interface{}{
			"subscription": "sub_invalid",
			"status":       "CONFIRMED",
		},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/psp/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("asaas-access-token", "wrong-signature")
	rr := httptest.NewRecorder()

	h.PSPWebhook(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
