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

func TestValidateUsageByOwner_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-oval@test.com", "Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Val", "cafe-val")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano", 2990, "daily", 1)

	sub := testutil.SeedSubscriber(t, h.Queries, "sub-oval@test.com", "Sub User", "")
	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      sub.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_oval_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]int64{"subscriber_id": sub.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.ValidateUsageByOwner(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "validated", resp["status"])
	assert.Equal(t, float64(1), resp["used"])
	assert.Equal(t, "Plano", resp["plan_name"])
}

func TestValidateUsageByOwner_LimitExceeded(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-oval-limit@test.com", "Owner Limit")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Limit Oval", "cafe-limit-oval")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Limit", 2990, "daily", 1)

	sub := testutil.SeedSubscriber(t, h.Queries, "sub-oval-limit@test.com", "Sub Limit", "")
	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      sub.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_oval_limit", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	// First validation should succeed
	body, _ := json.Marshal(map[string]int64{"subscriber_id": sub.ID})
	req1 := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1 = withAuth(req1, owner.ID, "owner")
	rr1 := httptest.NewRecorder()
	h.ValidateUsageByOwner(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	// Second validation should fail (daily limit = 1)
	body2, _ := json.Marshal(map[string]int64{"subscriber_id": sub.ID})
	req2 := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2 = withAuth(req2, owner.ID, "owner")
	rr2 := httptest.NewRecorder()
	h.ValidateUsageByOwner(rr2, req2)
	assert.Equal(t, http.StatusForbidden, rr2.Code)
}

func TestValidateUsageByOwner_SubscriberNotFound(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-oval2@test.com", "Owner2")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Val2", "cafe-val2")

	body, _ := json.Marshal(map[string]int64{"subscriber_id": 99999})
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.ValidateUsageByOwner(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestValidateUsageByOwner_MissingSubscriberID(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-oval3@test.com", "Owner3")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Val3", "cafe-val3")

	body, _ := json.Marshal(map[string]int64{"subscriber_id": 0})
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.ValidateUsageByOwner(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateUsageByOwner_NoBusiness(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-oval-nobiz@test.com", "Owner NoBiz")

	body, _ := json.Marshal(map[string]int64{"subscriber_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.ValidateUsageByOwner(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
