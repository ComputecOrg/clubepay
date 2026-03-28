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

func TestValidateUsage_InvalidJSON(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "usagejson@test.com", "Usage JSON Sub", "11999994100")

	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ValidateUsage(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateUsage_BusinessNotFound(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "usagenobiz@test.com", "Usage NoBiz Sub", "11999994200")

	body := map[string]string{"business_slug": "nonexistent"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ValidateUsage(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestValidateUsage_NoSubscription(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "usagenosub@test.com", "Usage NoSub Owner")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Café NoSub Usage", "cafe-nosub-usage")
	subscriber := testutil.SeedSubscriber(t, h.Queries, "usagenosubscriber@test.com", "Usage NoSub Sub", "11999994300")

	body := map[string]string{"business_slug": "cafe-nosub-usage"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ValidateUsage(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestValidateUsage_MonthlyLimit(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "monthlylimitowner@test.com", "Monthly Limit Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Monthly", "cafe-monthly-limit")
	// Monthly plan with limit 2
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Monthly", 2990, "monthly", 2)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "monthlylimitsub@test.com", "Monthly Limit Sub", "11999994400")

	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_monthly_limit", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	slug := "cafe-monthly-limit"

	// First usage — should succeed
	for i := 0; i < 2; i++ {
		b2, _ := json.Marshal(map[string]string{"business_slug": slug})
		req := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b2))
		req.Header.Set("Content-Type", "application/json")
		req = withAuth(req, subscriber.ID, "subscriber")
		rr := httptest.NewRecorder()
		h.ValidateUsage(rr, req)
		require.Equal(t, http.StatusOK, rr.Code, "usage %d should succeed", i+1)
	}

	// Third usage — should fail (monthly limit=2)
	b3, _ := json.Marshal(map[string]string{"business_slug": slug})
	req3 := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(b3))
	req3.Header.Set("Content-Type", "application/json")
	req3 = withAuth(req3, subscriber.ID, "subscriber")
	rr3 := httptest.NewRecorder()
	h.ValidateUsage(rr3, req3)

	assert.Equal(t, http.StatusForbidden, rr3.Code)
}

func TestMyUsage_MissingSlug(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myusagenoslug@test.com", "MyUsage NoSlug Sub", "11999994500")

	req := httptest.NewRequest(http.MethodGet, "/api/my-usage", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyUsage(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMyUsage_BusinessNotFound(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myusagenobiz@test.com", "MyUsage NoBiz Sub", "11999994600")

	req := httptest.NewRequest(http.MethodGet, "/api/my-usage?business_slug=nonexistent", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyUsage(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestMyUsage_NoSubscription(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "myusagenosub@test.com", "MyUsage NoSub Owner")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Café MyUsage NoSub", "cafe-myusage-nosub")
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myusagenosubscriber@test.com", "MyUsage NoSub Sub", "11999994700")

	req := httptest.NewRequest(http.MethodGet, "/api/my-usage?business_slug=cafe-myusage-nosub", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyUsage(rr, req)

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
