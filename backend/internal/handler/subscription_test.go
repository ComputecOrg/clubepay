package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
)

func TestSubscribe_Success(t *testing.T) {
	h := setupHandler(t)

	// Seed owner + business + plan
	owner := testutil.SeedOwner(t, h.Queries, "subowner@test.com", "Owner Sub")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Sub", "cafe-sub")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Café", 2990, "daily", 1)

	// Seed subscriber
	subscriber := testutil.SeedSubscriber(t, h.Queries, "subscriber@test.com", "Sub User", "11999990000")

	body := map[string]interface{}{
		"plan_id": plan.ID,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.Subscribe(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "active", resp["status"])
	assert.Equal(t, float64(plan.ID), resp["plan_id"])
	assert.Equal(t, float64(subscriber.ID), resp["subscriber_id"])
}

func TestSubscribe_AlreadySubscribed(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "subdup@test.com", "Owner Dup")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Dup", "cafe-dup")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Dup", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "dupsubscriber@test.com", "Dup Sub", "11999990001")

	// First subscription
	body := map[string]interface{}{"plan_id": plan.ID}
	b, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	req1.Header.Set("Content-Type", "application/json")
	req1 = withAuth(req1, subscriber.ID, "subscriber")
	rr1 := httptest.NewRecorder()
	h.Subscribe(rr1, req1)
	require.Equal(t, http.StatusCreated, rr1.Code)

	// Second subscription — should fail with 409
	b2, _ := json.Marshal(body)
	req2 := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2 = withAuth(req2, subscriber.ID, "subscriber")
	rr2 := httptest.NewRecorder()
	h.Subscribe(rr2, req2)

	assert.Equal(t, http.StatusConflict, rr2.Code)
}

func TestSubscribe_FreeTierLimit(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "freelimit@test.com", "Owner Free")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Free", "cafe-free")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Free", 2990, "daily", 1)

	// Create 15 subscriptions directly in DB
	for i := 0; i < 15; i++ {
		sub := testutil.SeedSubscriber(t, h.Queries, fmt.Sprintf("freesub%d@test.com", i), fmt.Sprintf("Sub %d", i), fmt.Sprintf("1199999%04d", i))
		_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
			PlanID:            plan.ID,
			SubscriberID:      sub.ID,
			BusinessID:        biz.ID,
			PspSubscriptionID: pgtype.Text{String: fmt.Sprintf("psp_%d", i), Valid: true},
			Status:            "active",
			PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
		})
		require.NoError(t, err)
	}

	// 16th subscriber should be rejected
	sub16 := testutil.SeedSubscriber(t, h.Queries, "freesub16@test.com", "Sub 16", "11999991600")
	body := map[string]interface{}{"plan_id": plan.ID}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, sub16.ID, "subscriber")
	rr := httptest.NewRecorder()
	h.Subscribe(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestListSubscriptions_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "listsubs@test.com", "Owner List")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café List", "cafe-list")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano List", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "listsub1@test.com", "List Sub", "11999992000")

	// Subscribe
	body := map[string]interface{}{"plan_id": plan.ID}
	b, _ := json.Marshal(body)
	subReq := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	subReq.Header.Set("Content-Type", "application/json")
	subReq = withAuth(subReq, subscriber.ID, "subscriber")
	subRr := httptest.NewRecorder()
	h.Subscribe(subRr, subReq)
	require.Equal(t, http.StatusCreated, subRr.Code)

	// List as owner
	listReq := httptest.NewRequest(http.MethodGet, "/api/subscriptions", nil)
	listReq = withAuth(listReq, owner.ID, "owner")
	listRr := httptest.NewRecorder()
	h.ListSubscriptions(listRr, listReq)

	require.Equal(t, http.StatusOK, listRr.Code)

	var resp []map[string]interface{}
	require.NoError(t, json.NewDecoder(listRr.Body).Decode(&resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "List Sub", resp[0]["subscriber_name"])
	assert.Equal(t, "Plano List", resp[0]["plan_name"])
}

func TestCancelSubscriptionByOwner_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "cancelsub@test.com", "Owner Cancel")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Cancel", "cafe-cancel")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Cancel", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "cancelsub1@test.com", "Cancel Sub", "11999993000")

	// Subscribe first
	body := map[string]interface{}{"plan_id": plan.ID}
	b, _ := json.Marshal(body)
	subReq := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	subReq.Header.Set("Content-Type", "application/json")
	subReq = withAuth(subReq, subscriber.ID, "subscriber")
	subRr := httptest.NewRecorder()
	h.Subscribe(subRr, subReq)
	require.Equal(t, http.StatusCreated, subRr.Code)

	var subResp map[string]interface{}
	require.NoError(t, json.NewDecoder(subRr.Body).Decode(&subResp))
	subID := int64(subResp["id"].(float64))

	// Cancel via chi router (DELETE /api/subscriptions/{id})
	cancelReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/subscriptions/%d", subID), nil)
	cancelReq = withAuth(cancelReq, owner.ID, "owner")

	r := chi.NewRouter()
	r.Delete("/api/subscriptions/{id}", h.CancelSubscriptionByOwner)

	cancelRr := httptest.NewRecorder()
	r.ServeHTTP(cancelRr, cancelReq)

	assert.Equal(t, http.StatusNoContent, cancelRr.Code)
}

func TestSubscribe_InvalidJSON(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "subjson@test.com", "JSON Sub", "11999993100")

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.Subscribe(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubscribe_MissingPlanID(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "subnoplan@test.com", "NoPlan Sub", "11999993200")

	body := map[string]interface{}{"plan_id": 0}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.Subscribe(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubscribe_PlanNotFound(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "subnotfound@test.com", "NotFound Sub", "11999993300")

	body := map[string]interface{}{"plan_id": 999999}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.Subscribe(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSubscribe_PlanInactive(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "inactiveplanowner@test.com", "Inactive Plan Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Inactive", "cafe-inactive-plan")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Inativo", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "inactiveplansub@test.com", "Inactive Plan Sub", "11999993400")

	// Deactivate the plan
	err := h.Queries.DeactivatePlan(context.Background(), plan.ID)
	require.NoError(t, err)

	body := map[string]interface{}{"plan_id": plan.ID}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.Subscribe(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCancelSubscriptionByOwner_NotFound(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "cancelnotfound@test.com", "Cancel NotFound Owner")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Café NotFound Cancel", "cafe-cancel-notfound")

	r := chi.NewRouter()
	r.Delete("/api/subscriptions/{id}", h.CancelSubscriptionByOwner)

	cancelReq := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/999999", nil)
	cancelReq = withAuth(cancelReq, owner.ID, "owner")
	cancelRr := httptest.NewRecorder()
	r.ServeHTTP(cancelRr, cancelReq)

	assert.Equal(t, http.StatusNotFound, cancelRr.Code)
}

func TestCancelSubscriptionByOwner_WrongBusiness(t *testing.T) {
	h := setupHandler(t)

	// Owner1 creates a business and a subscriber
	owner1 := testutil.SeedOwner(t, h.Queries, "cancelowner1@test.com", "Owner1 Cancel")
	biz1 := testutil.SeedBusiness(t, h.Queries, owner1.ID, "Café Owner1", "cafe-owner1-cancel")
	plan1 := testutil.SeedPlan(t, h.Queries, biz1.ID, "Plano Owner1", 2990, "daily", 1)
	subscriber1 := testutil.SeedSubscriber(t, h.Queries, "sub1cancel@test.com", "Sub1 Cancel", "11999993500")

	// Create a subscription for owner1's business
	sub, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan1.ID,
		SubscriberID:      subscriber1.ID,
		BusinessID:        biz1.ID,
		PspSubscriptionID: pgtype.Text{String: "sub_cancel_wrong", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// Owner2 tries to cancel owner1's subscriber's subscription
	owner2 := testutil.SeedOwner(t, h.Queries, "cancelowner2@test.com", "Owner2 Cancel")
	testutil.SeedBusiness(t, h.Queries, owner2.ID, "Café Owner2", "cafe-owner2-cancel")

	r := chi.NewRouter()
	r.Delete("/api/subscriptions/{id}", h.CancelSubscriptionByOwner)

	cancelReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/subscriptions/%d", sub.ID), nil)
	cancelReq = withAuth(cancelReq, owner2.ID, "owner")
	cancelRr := httptest.NewRecorder()
	r.ServeHTTP(cancelRr, cancelReq)

	assert.Equal(t, http.StatusForbidden, cancelRr.Code)
}

func TestListSubscriptions_NoBusiness(t *testing.T) {
	h := setupHandler(t)

	// Owner with no business registered
	owner := testutil.SeedOwner(t, h.Queries, "nobizowner@test.com", "NoBiz Owner")

	listReq := httptest.NewRequest(http.MethodGet, "/api/subscriptions", nil)
	listReq = withAuth(listReq, owner.ID, "owner")
	listRr := httptest.NewRecorder()

	h.ListSubscriptions(listRr, listReq)

	assert.Equal(t, http.StatusNotFound, listRr.Code)
}
