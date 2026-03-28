package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registerOwnerAndGetIDs registers a new owner via handler and returns (userID, businessID).
func registerOwnerAndGetIDs(t *testing.T, h interface {
	RegisterOwner(http.ResponseWriter, *http.Request)
}, email, bizName string) (int64, int64) {
	t.Helper()
	regBody := map[string]string{
		"email":         email,
		"password":      "password123",
		"name":          "Plan Test Owner",
		"business_name": bizName,
		"segment":       "cafeteria",
	}
	b, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RegisterOwner(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	userMap := resp["user"].(map[string]interface{})
	bizMap := resp["business"].(map[string]interface{})
	userID := int64(userMap["id"].(float64))
	bizID := int64(bizMap["id"].(float64))
	return userID, bizID
}

// ---- CreatePlan tests ----

func TestCreatePlan_Success(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "createplan@example.com", "Create Plan Café")

	planBody := map[string]interface{}{
		"name":        "Plano Café Diário",
		"description": "1 café por dia",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	b, _ := json.Marshal(planBody)
	req := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.CreatePlan(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Plano Café Diário", resp["name"])
	assert.Equal(t, float64(2990), resp["price_cents"])
	assert.Equal(t, "daily", resp["limit_type"])
	assert.Equal(t, float64(1), resp["limit_count"])
	assert.Equal(t, true, resp["active"])
}

func TestCreatePlan_FreeTierLimit(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "freetier@example.com", "Free Tier Café")

	// Create first plan — should succeed
	planBody := map[string]interface{}{
		"name":        "Plano 1",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	b, _ := json.Marshal(planBody)
	req1 := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	req1.Header.Set("Content-Type", "application/json")
	req1 = withAuth(req1, userID, "owner")
	rr1 := httptest.NewRecorder()
	h.CreatePlan(rr1, req1)
	require.Equal(t, http.StatusCreated, rr1.Code)

	// Create second plan — should fail with 403
	planBody2 := map[string]interface{}{
		"name":        "Plano 2",
		"price_cents": 4990,
		"limit_type":  "monthly",
		"limit_count": 4,
	}
	b2, _ := json.Marshal(planBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2 = withAuth(req2, userID, "owner")
	rr2 := httptest.NewRecorder()
	h.CreatePlan(rr2, req2)

	assert.Equal(t, http.StatusForbidden, rr2.Code)

	var errResp map[string]string
	require.NoError(t, json.NewDecoder(rr2.Body).Decode(&errResp))
	assert.Contains(t, errResp["error"], "plano")
}

func TestCreatePlan_InvalidInput(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "createplan-invalid@example.com", "Invalid Input Café")

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "missing name",
			body: map[string]interface{}{
				"price_cents": 2990,
				"limit_type":  "daily",
				"limit_count": 1,
			},
		},
		{
			name: "zero price",
			body: map[string]interface{}{
				"name":        "Plano Sem Preço",
				"price_cents": 0,
				"limit_type":  "daily",
				"limit_count": 1,
			},
		},
		{
			name: "zero limit_count",
			body: map[string]interface{}{
				"name":        "Plano Sem Limite",
				"price_cents": 2990,
				"limit_type":  "daily",
				"limit_count": 0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req = withAuth(req, userID, "owner")
			rr := httptest.NewRecorder()

			h.CreatePlan(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestCreatePlan_InvalidLimitType(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "createplan-limittype@example.com", "Limit Type Café")

	body := map[string]interface{}{
		"name":        "Plano Semanal",
		"price_cents": 2990,
		"limit_type":  "weekly",
		"limit_count": 1,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()

	h.CreatePlan(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreatePlan_NoBusiness(t *testing.T) {
	h := setupHandler(t)

	// Use a userID with no business
	body := map[string]interface{}{
		"name":        "Plano Sem Negócio",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 999999, "owner")
	rr := httptest.NewRecorder()

	h.CreatePlan(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---- ListPlans tests ----

func TestListPlans_Success(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "listplans@example.com", "List Plans Café")

	// Create 1 plan (free tier allows only 1)
	planBody := map[string]interface{}{
		"name":        "Plano Lista",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	b, _ := json.Marshal(planBody)
	req := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()
	h.CreatePlan(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	// List plans
	listReq := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	listReq = withAuth(listReq, userID, "owner")
	listRr := httptest.NewRecorder()

	h.ListPlans(listRr, listReq)

	require.Equal(t, http.StatusOK, listRr.Code)

	var resp []interface{}
	require.NoError(t, json.NewDecoder(listRr.Body).Decode(&resp))
	assert.Len(t, resp, 1)
}

// ---- UpdatePlan tests ----

func TestUpdatePlan_Success(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "updateplan@example.com", "Update Plan Café")

	// Create plan first
	planBody := map[string]interface{}{
		"name":        "Plano Original",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	b, _ := json.Marshal(planBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withAuth(createReq, userID, "owner")
	createRr := httptest.NewRecorder()
	h.CreatePlan(createRr, createReq)
	require.Equal(t, http.StatusCreated, createRr.Code)

	var createdPlan map[string]interface{}
	require.NoError(t, json.NewDecoder(createRr.Body).Decode(&createdPlan))
	planID := int64(createdPlan["id"].(float64))

	// Update plan via chi router to handle URL param
	updateBody := map[string]interface{}{
		"name":        "Plano Atualizado",
		"description": "Nova descrição",
		"price_cents": 3990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	ub, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/plans/%d", planID), bytes.NewReader(ub))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = withAuth(updateReq, userID, "owner")

	r := chi.NewRouter()
	r.Put("/api/plans/{id}", h.UpdatePlan)

	updateRr := httptest.NewRecorder()
	r.ServeHTTP(updateRr, updateReq)

	require.Equal(t, http.StatusOK, updateRr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(updateRr.Body).Decode(&resp))
	assert.Equal(t, "Plano Atualizado", resp["name"])
	assert.Equal(t, float64(3990), resp["price_cents"])
}

func TestListPlans_NoBusiness(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req = withAuth(req, 999999, "owner")
	rr := httptest.NewRecorder()

	h.ListPlans(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdatePlan_NotFound(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "updateplan-notfound@example.com", "Update Plan NotFound Café")

	updateBody := map[string]interface{}{
		"name":        "Plano Inexistente",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	ub, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/plans/999999", bytes.NewReader(ub))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = withAuth(updateReq, userID, "owner")

	r := chi.NewRouter()
	r.Put("/api/plans/{id}", h.UpdatePlan)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, updateReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdatePlan_WrongBusiness(t *testing.T) {
	h := setupHandler(t)

	// Create first owner+business+plan
	userID1, _ := registerOwnerAndGetIDs(t, h, "updateplan-wrong1@example.com", "Wrong Business 1 Café")
	planBody := map[string]interface{}{
		"name":        "Plano do Dono 1",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	pb, _ := json.Marshal(planBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(pb))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withAuth(createReq, userID1, "owner")
	createRr := httptest.NewRecorder()
	h.CreatePlan(createRr, createReq)
	require.Equal(t, http.StatusCreated, createRr.Code)

	var createdPlan map[string]interface{}
	require.NoError(t, json.NewDecoder(createRr.Body).Decode(&createdPlan))
	planID := int64(createdPlan["id"].(float64))

	// Create second owner+business — they try to update owner1's plan
	userID2, _ := registerOwnerAndGetIDs(t, h, "updateplan-wrong2@example.com", "Wrong Business 2 Café")

	updateBody := map[string]interface{}{
		"name":        "Tentativa de Roubo",
		"price_cents": 100,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	ub, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/plans/%d", planID), bytes.NewReader(ub))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = withAuth(updateReq, userID2, "owner")

	r := chi.NewRouter()
	r.Put("/api/plans/{id}", h.UpdatePlan)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, updateReq)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// ---- DeletePlan tests ----

func TestDeletePlan_Success(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "deleteplan@example.com", "Delete Plan Café")

	// Create plan first
	planBody := map[string]interface{}{
		"name":        "Plano Para Deletar",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	b, _ := json.Marshal(planBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(b))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withAuth(createReq, userID, "owner")
	createRr := httptest.NewRecorder()
	h.CreatePlan(createRr, createReq)
	require.Equal(t, http.StatusCreated, createRr.Code)

	var createdPlan map[string]interface{}
	require.NoError(t, json.NewDecoder(createRr.Body).Decode(&createdPlan))
	planID := int64(createdPlan["id"].(float64))

	// Delete plan via chi router
	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/plans/%d", planID), nil)
	deleteReq = withAuth(deleteReq, userID, "owner")

	r := chi.NewRouter()
	r.Delete("/api/plans/{id}", h.DeletePlan)

	deleteRr := httptest.NewRecorder()
	r.ServeHTTP(deleteRr, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteRr.Code)
}

func TestDeletePlan_NotFound(t *testing.T) {
	h := setupHandler(t)
	userID, _ := registerOwnerAndGetIDs(t, h, "deleteplan-notfound@example.com", "Delete Plan NotFound Café")

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/plans/999999", nil)
	deleteReq = withAuth(deleteReq, userID, "owner")

	r := chi.NewRouter()
	r.Delete("/api/plans/{id}", h.DeletePlan)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeletePlan_WrongBusiness(t *testing.T) {
	h := setupHandler(t)

	// Create first owner+business+plan
	userID1, _ := registerOwnerAndGetIDs(t, h, "deleteplan-wrong1@example.com", "Delete Wrong Biz 1 Café")
	planBody := map[string]interface{}{
		"name":        "Plano para Deletar",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	pb, _ := json.Marshal(planBody)
	createReq := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(pb))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withAuth(createReq, userID1, "owner")
	createRr := httptest.NewRecorder()
	h.CreatePlan(createRr, createReq)
	require.Equal(t, http.StatusCreated, createRr.Code)

	var createdPlan map[string]interface{}
	require.NoError(t, json.NewDecoder(createRr.Body).Decode(&createdPlan))
	planID := int64(createdPlan["id"].(float64))

	// Create second owner — tries to delete owner1's plan
	userID2, _ := registerOwnerAndGetIDs(t, h, "deleteplan-wrong2@example.com", "Delete Wrong Biz 2 Café")

	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/plans/%d", planID), nil)
	deleteReq = withAuth(deleteReq, userID2, "owner")

	r := chi.NewRouter()
	r.Delete("/api/plans/{id}", h.DeletePlan)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteReq)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
