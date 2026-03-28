package handler_test

import (
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

func TestReconcile_BlocksExpiredGrace(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{}
	h := handler.New(queries, cfg, mockPSP)

	// Seed owner, business, plan, subscriber
	owner := testutil.SeedOwner(t, queries, "cronowner@test.com", "Cron Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Cron Cafe", "cron-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Cron Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, queries, "cronsub@test.com", "Cron Sub", "11999990002")

	// Create subscription in grace with past deadline
	sub, err := queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_cron_123", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Set to grace with a past deadline
	pastDeadline := time.Now().Add(-24 * time.Hour)
	err = queries.UpdateSubscriptionGrace(context.Background(), repository.UpdateSubscriptionGraceParams{
		ID:            sub.ID,
		GraceDeadline: pgtype.Timestamptz{Time: pastDeadline, Valid: true},
	})
	require.NoError(t, err)

	// Run reconcile
	req := httptest.NewRequest(http.MethodPost, "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "test-cron-secret")
	rr := httptest.NewRecorder()

	h.Reconcile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, float64(1), resp["blocked"])

	// Verify subscription is blocked
	updated, err := queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "blocked", updated.Status)
}

func TestReconcile_InvalidSecret(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{}
	h := handler.New(queries, cfg, mockPSP)

	req := httptest.NewRequest(http.MethodPost, "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "wrong-secret")
	rr := httptest.NewRecorder()

	h.Reconcile(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
