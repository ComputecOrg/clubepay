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

func TestMyPlan_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "myplanowner@test.com", "MyPlan Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café MyPlan", "cafe-myplan")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano MyPlan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myplansub@test.com", "MyPlan Sub", "11999998000")

	// Create subscription directly
	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_myplan_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/my-plan?business_slug=cafe-myplan", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyPlan(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	planData := resp["plan"].(map[string]interface{})
	assert.Equal(t, "Plano MyPlan", planData["name"])
	assert.Equal(t, float64(2990), planData["price_cents"])

	bizData := resp["business"].(map[string]interface{})
	assert.Equal(t, "Café MyPlan", bizData["name"])
	assert.Equal(t, "cafe-myplan", bizData["slug"])

	subData := resp["subscription"].(map[string]interface{})
	assert.Equal(t, "active", subData["status"])
}

func TestMyPlan_NoSubscription(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "noplnowner@test.com", "NoPlan Owner")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Café NoPlan", "cafe-noplan")
	subscriber := testutil.SeedSubscriber(t, h.Queries, "noplnsub@test.com", "NoPlan Sub", "11999998100")

	req := httptest.NewRequest(http.MethodGet, "/api/my-plan?business_slug=cafe-noplan", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyPlan(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestMyPlan_MissingSlug(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myplannoslug@test.com", "MyPlan NoSlug Sub", "11999998300")

	req := httptest.NewRequest(http.MethodGet, "/api/my-plan", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyPlan(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMyPlan_BusinessNotFound(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myplandnobiz@test.com", "MyPlan NoBiz Sub", "11999998400")

	req := httptest.NewRequest(http.MethodGet, "/api/my-plan?business_slug=nonexistent", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyPlan(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCancelBySubscriber_BusinessNotFound(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "selfcancelnobiz@test.com", "SelfCancel NoBiz Sub", "11999998700")

	body := map[string]string{"business_slug": "nonexistent-biz-slug"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/cancel", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.CancelBySubscriber(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCancelBySubscriber_InvalidJSON(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "selfcanceljson@test.com", "SelfCancel JSON Sub", "11999998500")

	req := httptest.NewRequest(http.MethodPost, "/api/cancel", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.CancelBySubscriber(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCancelBySubscriber_NoSubscription(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "selfcancelnosub@test.com", "SelfCancel NoSub Owner")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Café SelfCancel NoSub", "cafe-selfcancel-nosub")
	subscriber := testutil.SeedSubscriber(t, h.Queries, "selfcancelnosubscriber@test.com", "SelfCancel NoSub Sub", "11999998600")

	body := map[string]string{"business_slug": "cafe-selfcancel-nosub"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/cancel", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.CancelBySubscriber(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCancelBySubscriber_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "selfcancelowner@test.com", "SelfCancel Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café SelfCancel", "cafe-selfcancel")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano SelfCancel", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "selfcancelsub@test.com", "SelfCancel Sub", "11999998200")

	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_selfcancel_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	body := map[string]string{"business_slug": "cafe-selfcancel"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/cancel", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.CancelBySubscriber(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "cancelled", resp["status"])
	assert.NotEmpty(t, resp["message"])
}
