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

	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
)

func TestMyPlan_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "myplanowner@test.com", "MyPlan Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café MyPlan", "cafe-myplan")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano MyPlan", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "myplansub@test.com", "MyPlan Sub", "11999998000")

	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriber.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_myplan_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/my-plan", nil)
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

	subscriber := testutil.SeedSubscriber(t, h.Queries, "noplnsub@test.com", "NoPlan Sub", "11999998100")

	req := httptest.NewRequest(http.MethodGet, "/api/my-plan", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyPlan(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCancelBySubscriber_NoSubscription(t *testing.T) {
	h := setupHandler(t)

	subscriber := testutil.SeedSubscriber(t, h.Queries, "selfcancelnosubscriber@test.com", "SelfCancel NoSub Sub", "11999998600")

	req := httptest.NewRequest(http.MethodPost, "/api/cancel", nil)
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

	req := httptest.NewRequest(http.MethodPost, "/api/cancel", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.CancelBySubscriber(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "cancelled", resp["status"])
	assert.NotEmpty(t, resp["message"])
}
