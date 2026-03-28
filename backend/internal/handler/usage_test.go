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

	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
)

func TestValidateUsage_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "usageowner@test.com", "Usage Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Usage", "cafe-usage")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Usage", 2990, "daily", 2)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "usagesub@test.com", "Usage Sub", "11999994000")

	// Create subscription directly
	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_usage_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	body := map[string]string{"business_slug": "cafe-usage"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ValidateUsage(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "validated", resp["status"])
	assert.Equal(t, float64(1), resp["used"])
	assert.Equal(t, float64(2), resp["limit"])
	assert.Equal(t, "Plano Usage", resp["plan_name"])
}

func TestValidateUsage_LimitExceeded(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "limitowner@test.com", "Limit Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Limit", "cafe-limit")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Limit", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "limitsub@test.com", "Limit Sub", "11999995000")

	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_limit_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// First validation — should succeed
	body := map[string]string{"business_slug": "cafe-limit"}
	b, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b))
	req1.Header.Set("Content-Type", "application/json")
	req1 = withAuth(req1, subscriber.ID, "subscriber")
	rr1 := httptest.NewRecorder()
	h.ValidateUsage(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	// Second validation — should fail (daily limit=1)
	b2, _ := json.Marshal(body)
	req2 := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2 = withAuth(req2, subscriber.ID, "subscriber")
	rr2 := httptest.NewRecorder()
	h.ValidateUsage(rr2, req2)

	assert.Equal(t, http.StatusForbidden, rr2.Code)
}

func TestValidateUsage_BlockedSubscription(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "blockedowner@test.com", "Blocked Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Blocked", "cafe-blocked")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Blocked", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "blockedsub@test.com", "Blocked Sub", "11999996000")

	// Create subscription and then block it
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_blocked_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Block the subscription
	err = h.Queries.UpdateSubscriptionStatus(context.Background(), repository.UpdateSubscriptionStatusParams{
		ID:     sub.ID,
		Status: "blocked",
	})
	require.NoError(t, err)

	body := map[string]string{"business_slug": "cafe-blocked"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ValidateUsage(rr, req)

	// GetActiveSubscription only returns active/grace, so a blocked subscription will be "not found"
	// The handler should return 404 or 403 depending on how it's handled
	// Since GetActiveSubscription filters by status IN ('active', 'grace'), blocked won't be found
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestMyUsage_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "myusageowner@test.com", "My Usage Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café MyUsage", "cafe-myusage")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano MyUsage", 2990, "daily", 5)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myusagesub@test.com", "MyUsage Sub", "11999997000")

	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_myusage_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Validate once
	body := map[string]string{"business_slug": "cafe-myusage"}
	b, _ := json.Marshal(body)
	valReq := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b))
	valReq.Header.Set("Content-Type", "application/json")
	valReq = withAuth(valReq, subscriber.ID, "subscriber")
	valRr := httptest.NewRecorder()
	h.ValidateUsage(valRr, valReq)
	require.Equal(t, http.StatusOK, valRr.Code)

	// Get my-usage
	myReq := httptest.NewRequest(http.MethodGet, "/api/my-usage?business_slug=cafe-myusage", nil)
	myReq = withAuth(myReq, subscriber.ID, "subscriber")
	myRr := httptest.NewRecorder()

	h.MyUsage(myRr, myReq)

	require.Equal(t, http.StatusOK, myRr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(myRr.Body).Decode(&resp))
	assert.Equal(t, float64(1), resp["used"])
	assert.Equal(t, float64(5), resp["limit"])
	assert.Equal(t, "Plano MyUsage", resp["plan_name"])
}
