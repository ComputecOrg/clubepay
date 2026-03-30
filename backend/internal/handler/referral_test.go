package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/testutil"
)

func TestMyReferralCode_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "refowner@test.com", "Ref Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café Ref", "cafe-ref")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Ref", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "refsub@test.com", "Ref Sub", "11999990100")
	testutil.SeedSubscription(t, h.Queries, plan.ID, subscriber.ID, biz.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/my-referral-code", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyReferralCode(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["code"])
	assert.Len(t, resp["code"], 8)
}

func TestMyReferralCode_SameCodeTwice(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "reftwice@test.com", "Ref Twice Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café RefTwice", "cafe-reftwice")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano RefTwice", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "reftwicesub@test.com", "Ref Twice Sub", "11999990200")
	testutil.SeedSubscription(t, h.Queries, plan.ID, subscriber.ID, biz.ID)

	// First call
	req1 := httptest.NewRequest(http.MethodGet, "/api/my-referral-code", nil)
	req1 = withAuth(req1, subscriber.ID, "subscriber")
	rr1 := httptest.NewRecorder()
	h.MyReferralCode(rr1, req1)
	require.Equal(t, http.StatusOK, rr1.Code)

	var resp1 map[string]string
	require.NoError(t, json.NewDecoder(rr1.Body).Decode(&resp1))

	// Second call
	req2 := httptest.NewRequest(http.MethodGet, "/api/my-referral-code", nil)
	req2 = withAuth(req2, subscriber.ID, "subscriber")
	rr2 := httptest.NewRecorder()
	h.MyReferralCode(rr2, req2)
	require.Equal(t, http.StatusOK, rr2.Code)

	var resp2 map[string]string
	require.NoError(t, json.NewDecoder(rr2.Body).Decode(&resp2))

	// Same code
	assert.Equal(t, resp1["code"], resp2["code"])
}

func TestMyReferralCode_NoSubscription(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "refnoslug@test.com", "Ref NoSub Sub", "11999990400")

	req := httptest.NewRequest(http.MethodGet, "/api/my-referral-code", nil)
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.MyReferralCode(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestApplyReferral_InvalidJSON(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "applyjson@test.com", "Apply JSON Sub", "11999990600")

	req := httptest.NewRequest(http.MethodPost, "/api/referrals/apply", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ApplyReferral(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestApplyReferral_EmptyCode(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "applyempty@test.com", "Apply Empty Sub", "11999990700")

	body := map[string]string{"code": ""}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/referrals/apply", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ApplyReferral(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestApplyReferral_CodeNotFound(t *testing.T) {
	h := setupHandler(t)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "applynotfound@test.com", "Apply NotFound Sub", "11999990800")

	body := map[string]string{"code": "nonexistent"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/referrals/apply", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, subscriber.ID, "subscriber")
	rr := httptest.NewRecorder()

	h.ApplyReferral(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestApplyReferral_ReferralLimit(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "reflimitowner@test.com", "Ref Limit Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café RefLimit", "cafe-reflimit")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano RefLimit", 2990, "daily", 1)

	// Referrer gets their code
	referrer := testutil.SeedSubscriber(t, h.Queries, "reflimiter@test.com", "Ref Limiter", "11999990900")
	testutil.SeedSubscription(t, h.Queries, plan.ID, referrer.ID, biz.ID)

	codeReq := httptest.NewRequest(http.MethodGet, "/api/my-referral-code", nil)
	codeReq = withAuth(codeReq, referrer.ID, "subscriber")
	codeRr := httptest.NewRecorder()
	h.MyReferralCode(codeRr, codeReq)
	require.Equal(t, http.StatusOK, codeRr.Code)

	var codeResp map[string]string
	require.NoError(t, json.NewDecoder(codeRr.Body).Decode(&codeResp))
	referralCode := codeResp["code"]

	// Apply the code 3 times
	for i := 0; i < 3; i++ {
		referred := testutil.SeedSubscriber(t, h.Queries, fmt.Sprintf("reflimitreferred%d@test.com", i), fmt.Sprintf("Referred %d", i), fmt.Sprintf("1199999100%d", i))
		ab, _ := json.Marshal(map[string]string{"code": referralCode})
		applyReq := httptest.NewRequest(http.MethodPost, "/api/referrals/apply", bytes.NewReader(ab))
		applyReq.Header.Set("Content-Type", "application/json")
		applyReq = withAuth(applyReq, referred.ID, "subscriber")
		applyRr := httptest.NewRecorder()
		h.ApplyReferral(applyRr, applyReq)
		require.Equal(t, http.StatusOK, applyRr.Code, "referral %d should succeed", i+1)
	}

	// 4th should fail
	referred4 := testutil.SeedSubscriber(t, h.Queries, "reflimitreferred4@test.com", "Referred 4", "11999991004")
	ab4, _ := json.Marshal(map[string]string{"code": referralCode})
	applyReq4 := httptest.NewRequest(http.MethodPost, "/api/referrals/apply", bytes.NewReader(ab4))
	applyReq4.Header.Set("Content-Type", "application/json")
	applyReq4 = withAuth(applyReq4, referred4.ID, "subscriber")
	applyRr4 := httptest.NewRecorder()
	h.ApplyReferral(applyRr4, applyReq4)

	assert.Equal(t, http.StatusForbidden, applyRr4.Code)
}

func TestApplyReferral_SelfReferral(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "selfrefowner@test.com", "SelfRef Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Café SelfRef", "cafe-selfref")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano SelfRef", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, h.Queries, "selfrefsub@test.com", "SelfRef Sub", "11999990300")
	testutil.SeedSubscription(t, h.Queries, plan.ID, subscriber.ID, biz.ID)

	// Get own referral code
	codeReq := httptest.NewRequest(http.MethodGet, "/api/my-referral-code", nil)
	codeReq = withAuth(codeReq, subscriber.ID, "subscriber")
	codeRr := httptest.NewRecorder()
	h.MyReferralCode(codeRr, codeReq)
	require.Equal(t, http.StatusOK, codeRr.Code)

	var codeResp map[string]string
	require.NoError(t, json.NewDecoder(codeRr.Body).Decode(&codeResp))

	// Try to apply own code
	ab, _ := json.Marshal(map[string]string{"code": codeResp["code"]})
	applyReq := httptest.NewRequest(http.MethodPost, "/api/referrals/apply", bytes.NewReader(ab))
	applyReq.Header.Set("Content-Type", "application/json")
	applyReq = withAuth(applyReq, subscriber.ID, "subscriber")
	applyRr := httptest.NewRecorder()

	h.ApplyReferral(applyRr, applyReq)

	assert.Equal(t, http.StatusBadRequest, applyRr.Code)
}
