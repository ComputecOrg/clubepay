package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmokeFlow_RegisterSubscribeValidateCancel verifica o fluxo completo do MVP:
// dono registra → cria plano → assinante registra → assina → valida uso → cancela → uso bloqueado.
func TestSmokeFlow_RegisterSubscribeValidateCancel(t *testing.T) {
	h := setupHandler(t)

	// ── 1. Registrar dono e negócio ──────────────────────────────────────────
	ownerID, _ := registerOwnerAndGetIDs(t, h, "smokeowner@test.com", "Smoke Café")

	// ── 2. Criar plano (autenticado como dono) ───────────────────────────────
	planBody := map[string]interface{}{
		"name":        "Plano Smoke",
		"price_cents": 2990,
		"limit_type":  "daily",
		"limit_count": 1,
	}
	pb, _ := json.Marshal(planBody)
	planReq := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(pb))
	planReq.Header.Set("Content-Type", "application/json")
	planReq = withAuth(planReq, ownerID, "owner")
	planRR := httptest.NewRecorder()
	h.CreatePlan(planRR, planReq)
	require.Equal(t, http.StatusCreated, planRR.Code, "criar plano deve retornar 201")

	var planResp map[string]interface{}
	require.NoError(t, json.NewDecoder(planRR.Body).Decode(&planResp))
	planID := int64(planResp["id"].(float64))

	// ── 3. Registrar assinante ───────────────────────────────────────────────
	subRegBody := map[string]string{
		"email":    "smokesub@test.com",
		"password": "password123",
		"name":     "Smoke Assinante",
	}
	srb, _ := json.Marshal(subRegBody)
	subRegReq := httptest.NewRequest(http.MethodPost, "/api/auth/register-subscriber", bytes.NewReader(srb))
	subRegReq.Header.Set("Content-Type", "application/json")
	subRegRR := httptest.NewRecorder()
	h.RegisterSubscriber(subRegRR, subRegReq)
	require.Equal(t, http.StatusCreated, subRegRR.Code, "registrar assinante deve retornar 201")

	var subRegResp map[string]interface{}
	require.NoError(t, json.NewDecoder(subRegRR.Body).Decode(&subRegResp))
	userMap := subRegResp["user"].(map[string]interface{})
	subscriberID := int64(userMap["id"].(float64))

	// ── 4. Assinar plano ─────────────────────────────────────────────────────
	subscribeBody := map[string]interface{}{"plan_id": planID}
	subb, _ := json.Marshal(subscribeBody)
	subReq := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(subb))
	subReq.Header.Set("Content-Type", "application/json")
	subReq = withAuth(subReq, subscriberID, "subscriber")
	subRR := httptest.NewRecorder()
	h.Subscribe(subRR, subReq)
	require.Equal(t, http.StatusCreated, subRR.Code, "assinar deve retornar 201")

	var subResp map[string]interface{}
	require.NoError(t, json.NewDecoder(subRR.Body).Decode(&subResp))
	assert.Equal(t, "active", subResp["status"], "assinatura deve estar ativa")

	// ── 5. Validar uso ───────────────────────────────────────────────────────
	validateBody := map[string]string{"business_slug": "smoke-cafe"}
	vb, _ := json.Marshal(validateBody)
	validateReq := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(vb))
	validateReq.Header.Set("Content-Type", "application/json")
	validateReq = withAuth(validateReq, subscriberID, "subscriber")
	validateRR := httptest.NewRecorder()
	h.ValidateUsage(validateRR, validateReq)
	require.Equal(t, http.StatusOK, validateRR.Code, "primeiro uso do dia deve ser validado")

	var validateResp map[string]interface{}
	require.NoError(t, json.NewDecoder(validateRR.Body).Decode(&validateResp))
	assert.Equal(t, "validated", validateResp["status"])
	assert.Equal(t, float64(1), validateResp["used"])

	// ── 6. Limite diário: segundo uso deve falhar ────────────────────────────
	validateReq2 := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(vb))
	validateReq2.Header.Set("Content-Type", "application/json")
	validateReq2 = withAuth(validateReq2, subscriberID, "subscriber")
	validateRR2 := httptest.NewRecorder()
	h.ValidateUsage(validateRR2, validateReq2)
	assert.Equal(t, http.StatusForbidden, validateRR2.Code, "segundo uso no mesmo dia deve ser bloqueado")

	// ── 7. Ver meu plano ─────────────────────────────────────────────────────
	myPlanReq := httptest.NewRequest(http.MethodGet, "/api/my-plan", nil)
	myPlanReq = withAuth(myPlanReq, subscriberID, "subscriber")
	myPlanRR := httptest.NewRecorder()
	h.MyPlan(myPlanRR, myPlanReq)
	require.Equal(t, http.StatusOK, myPlanRR.Code)

	// ── 8. Cancelar assinatura ───────────────────────────────────────────────
	cancelReq := httptest.NewRequest(http.MethodPost, "/api/cancel", nil)
	cancelReq = withAuth(cancelReq, subscriberID, "subscriber")
	cancelRR := httptest.NewRecorder()
	h.CancelBySubscriber(cancelRR, cancelReq)
	require.Equal(t, http.StatusOK, cancelRR.Code, "cancelamento deve retornar 200")

	// ── 9. Após cancelar: uso deve retornar 404 (sem assinatura ativa) ───────
	validateReq3 := httptest.NewRequest(http.MethodPost, "/api/validate-usage", bytes.NewReader(vb))
	validateReq3.Header.Set("Content-Type", "application/json")
	validateReq3 = withAuth(validateReq3, subscriberID, "subscriber")
	validateRR3 := httptest.NewRecorder()
	h.ValidateUsage(validateRR3, validateReq3)
	assert.Equal(t, http.StatusNotFound, validateRR3.Code, "após cancelamento uso deve ser negado")
}
