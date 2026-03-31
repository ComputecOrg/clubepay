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
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/provider"
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
	mockEmail := &email.MockSender{}
	h := handler.New(queries, cfg, mockPSP, mockEmail)

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

func TestReconcile_SyncsPendingSubscription(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{
		GetSubscriptionFn: func(ctx context.Context, subscriptionID string) (*psp.Subscription, error) {
			return &psp.Subscription{ID: subscriptionID, Status: "ACTIVE"}, nil
		},
	}
	mockEmail := &email.MockSender{}
	h := handler.New(queries, cfg, mockPSP, mockEmail)

	owner := testutil.SeedOwner(t, queries, "cronpending@test.com", "Cron Pending Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Cron Pending Cafe", "cron-pending-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Cron Pending Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, queries, "cronpendingsub@test.com", "Cron Pending Sub", "11999990004")

	// Create subscription with status "pending"
	sub, err := queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_pending_sync", Valid: true},
		Status:            "pending",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
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
	assert.Equal(t, float64(1), resp["synced"])

	// Verify subscription status was changed to "active"
	updated, err := queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", updated.Status)
}

func TestReconcile_NoExpiredGrace(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{}
	mockEmail := &email.MockSender{}
	h := handler.New(queries, cfg, mockPSP, mockEmail)

	// Run reconcile with no expired grace subscriptions
	req := httptest.NewRequest(http.MethodPost, "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "test-cron-secret")
	rr := httptest.NewRecorder()

	h.Reconcile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, float64(0), resp["blocked"])
}

func TestReconcile_SyncsPendingToInactive(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{
		GetSubscriptionFn: func(ctx context.Context, subscriptionID string) (*psp.Subscription, error) {
			return &psp.Subscription{ID: subscriptionID, Status: "INACTIVE"}, nil
		},
	}
	mockEmail := &email.MockSender{}
	h := handler.New(queries, cfg, mockPSP, mockEmail)

	owner := testutil.SeedOwner(t, queries, "croninactive@test.com", "Cron Inactive Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Cron Inactive Cafe", "cron-inactive-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Cron Inactive Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, queries, "croninactivesub@test.com", "Cron Inactive Sub", "11999990005")

	// Create subscription with status "pending"
	sub, err := queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_pending_inactive", Valid: true},
		Status:            "pending",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Run reconcile — PSP returns INACTIVE → should map to "cancelled"
	req := httptest.NewRequest(http.MethodPost, "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "test-cron-secret")
	rr := httptest.NewRecorder()

	h.Reconcile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, float64(1), resp["synced"])

	// Verify subscription status was changed to "cancelled"
	updated, err := queries.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", updated.Status)
}

func TestReconcile_InvalidSecret(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{}
	mockEmail := &email.MockSender{}
	h := handler.New(queries, cfg, mockPSP, mockEmail)

	req := httptest.NewRequest(http.MethodPost, "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "wrong-secret")
	rr := httptest.NewRecorder()

	h.Reconcile(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// MockCostProvider for testing
type mockCostProvider struct {
	cost int64
}

func (m *mockCostProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	return provider.Cost{
		CostCents:   m.cost,
		Provider:    "mock",
		Description: "Mock provider",
	}, nil
}

func TestReconcile_UpdatesCosts(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:  "test-secret-key",
		CronSecret: "test-cron-secret",
	}
	mockPSP := &psp.MockPSP{}
	mockEmail := &email.MockSender{}
	h := handler.New(queries, cfg, mockPSP, mockEmail)

	// Setup cost providers (total: 28495 cents = Hostinger + Claude + Brevo)
	providers := []provider.CostProvider{
		&mockCostProvider{cost: 10000}, // Hostinger
		&mockCostProvider{cost: 12495}, // Claude API
		&mockCostProvider{cost: 6000},  // Brevo
	}
	agg := provider.NewAggregator(providers)
	h.CostAggregator = agg

	// Seed owner, business, plan, subscriber
	owner := testutil.SeedOwner(t, queries, "costowner@test.com", "Cost Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Cost Cafe", "cost-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Cost Plan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, queries, "costsub@test.com", "Cost Sub", "11999990010")

	// Create subscription
	sub, err := queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_cost_123", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)
	_ = sub // Prevent unused variable warning

	// Run reconcile
	req := httptest.NewRequest(http.MethodPost, "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "test-cron-secret")
	rr := httptest.NewRecorder()

	h.Reconcile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	// Verify monthly cost was updated
	currentMonth := time.Now()
	pgDate := pgtype.Date{
		Time:  currentMonth,
		Valid: true,
	}
	cost, err := queries.GetOrCreateMonthlyCost(context.Background(), repository.GetOrCreateMonthlyCostParams{
		BusinessID: biz.ID,
		Month:      pgDate,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(28495), cost.InfrastructureCostCents)
}
