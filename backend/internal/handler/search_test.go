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

func TestSearchSubscribers_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-search@test.com", "Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Search", "cafe-search")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Cafe", 2990, "daily", 1)

	sub := testutil.SeedSubscriber(t, h.Queries, "maria-search@test.com", "Maria Santos", "11999990000")
	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      sub.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_search_1", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=Maria", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	results := resp["results"].([]interface{})
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestSearchSubscribers_ByPhone(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner-search-phone@test.com", "Owner Phone")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Phone", "cafe-phone")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Phone", 2990, "daily", 1)

	sub := testutil.SeedSubscriber(t, h.Queries, "joao-phone@test.com", "Joao da Silva", "11888880000")
	_, err := h.Queries.CreateSubscription(context.Background(), repository.CreateSubscriptionParams{
		PlanID:            plan.ID,
		SubscriberID:      sub.ID,
		BusinessID:        biz.ID,
		PspSubscriptionID: pgtype.Text{String: "psp_search_phone", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=11888", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	results := resp["results"].([]interface{})
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestSearchSubscribers_EmptyQuery(t *testing.T) {
	h := setupHandler(t)
	owner := testutil.SeedOwner(t, h.Queries, "owner-search2@test.com", "Owner2")

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSearchSubscribers_NoResults(t *testing.T) {
	h := setupHandler(t)
	owner := testutil.SeedOwner(t, h.Queries, "owner-search3@test.com", "Owner3")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Empty", "cafe-empty")

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=Ninguem", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	results := resp["results"].([]interface{})
	assert.Len(t, results, 0)
}
