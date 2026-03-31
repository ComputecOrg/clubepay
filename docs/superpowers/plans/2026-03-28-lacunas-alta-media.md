# Lacunas Alta + Media — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar todas as lacunas de prioridade alta e media do ClubePay: Docker Compose, logging middleware, validator/v10, busca de assinante, email dunning, QR code, pagina de planos, e testes frontend.

**Architecture:** 8 tarefas independentes que podem ser paralelizadas em 3 grupos: infra backend (docker-compose, logging, validator), features backend (busca, email dunning), features frontend (QR, planos, testes).

**Tech Stack:** Go 1.22+, chi, sqlc, validator/v10, slog, net/smtp, Next.js 16, Vitest, qrcode (npm), Docker Compose

---

### Task 1: Docker Compose para dev local

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: Criar docker-compose.yml**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: clubepay
      POSTGRES_PASSWORD: clubepay
      POSTGRES_DB: clubepay_dev
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://clubepay:clubepay@postgres:5432/clubepay_dev?sslmode=disable
      JWT_SECRET: dev-secret-change-in-production
      PORT: "8080"
      ASAAS_URL: https://sandbox.asaas.com/api/v3
      CRON_SECRET: dev-cron-secret
    depends_on:
      - postgres

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    depends_on:
      - backend

volumes:
  pgdata:
```

- [ ] **Step 2: Adicionar targets ao Makefile**

Adicionar ao final do Makefile:

```makefile
# Docker Compose
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
```

- [ ] **Step 3: Testar compose up e verificar conectividade**

Run: `docker compose config`
Expected: Valid YAML parsed without errors

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml Makefile
git commit -m "chore: docker-compose para dev local com postgres, backend e frontend"
```

---

### Task 2: Logging Middleware

**Files:**
- Create: `backend/internal/middleware/logging.go`
- Create: `backend/internal/middleware/logging_test.go`
- Modify: `backend/cmd/api/router.go`

- [ ] **Step 1: Escrever teste para logging middleware**

```go
package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLogging_LogsRequestDetails(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "GET"), "log should contain HTTP method")
	assert.True(t, strings.Contains(logOutput, "/api/health"), "log should contain path")
	assert.True(t, strings.Contains(logOutput, "200"), "log should contain status code")
}

func TestLogging_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "404"), "log should contain 404 status")
}
```

- [ ] **Step 2: Rodar teste e verificar que falha**

Run: `cd backend && go test ./internal/middleware/ -run TestLogging -v`
Expected: FAIL — `Logging` not defined

- [ ] **Step 3: Implementar logging middleware**

```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
```

- [ ] **Step 4: Rodar teste e verificar que passa**

Run: `cd backend && go test ./internal/middleware/ -run TestLogging -v`
Expected: PASS

- [ ] **Step 5: Adicionar middleware ao router**

Em `backend/cmd/api/router.go`, adicionar `r.Use(middleware.Logging)` logo após `r.Use(middleware.CORS)`.

- [ ] **Step 6: Rodar todos os testes do backend**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/middleware/logging.go backend/internal/middleware/logging_test.go backend/cmd/api/router.go
git commit -m "feat: logging middleware com method, path, status e duration"
```

---

### Task 3: validator/v10

**Files:**
- Create: `backend/internal/domain/validate.go`
- Create: `backend/internal/domain/validate_test.go`
- Modify: `backend/internal/handler/auth.go`
- Modify: `backend/go.mod` (go get)

- [ ] **Step 1: Adicionar dependencia**

Run: `cd backend && go get github.com/go-playground/validator/v10`

- [ ] **Step 2: Escrever teste para validate helper**

```go
package domain_test

import (
	"testing"

	"github.com/clubepay/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidate_RegisterOwnerRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.RegisterOwnerInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: domain.RegisterOwnerInput{
				Email:        "test@example.com",
				Password:     "password123",
				Name:         "Joao",
				BusinessName: "Cafe do Joao",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			input: domain.RegisterOwnerInput{
				Password:     "password123",
				Name:         "Joao",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			input: domain.RegisterOwnerInput{
				Email:        "not-an-email",
				Password:     "password123",
				Name:         "Joao",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "short password",
			input: domain.RegisterOwnerInput{
				Email:        "test@example.com",
				Password:     "short",
				Name:         "Joao",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			input: domain.RegisterOwnerInput{
				Email:        "test@example.com",
				Password:     "password123",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "missing business name",
			input: domain.RegisterOwnerInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Joao",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_LoginInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.LoginInput
		wantErr bool
	}{
		{
			name:    "valid",
			input:   domain.LoginInput{Email: "a@b.com", Password: "12345678"},
			wantErr: false,
		},
		{
			name:    "missing email",
			input:   domain.LoginInput{Password: "12345678"},
			wantErr: true,
		},
		{
			name:    "missing password",
			input:   domain.LoginInput{Email: "a@b.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_RegisterSubscriberInput(t *testing.T) {
	valid := domain.RegisterSubscriberInput{
		Email:    "sub@example.com",
		Password: "password123",
		Name:     "Maria",
	}
	assert.NoError(t, domain.Validate(valid))

	invalid := domain.RegisterSubscriberInput{Email: "bad"}
	assert.Error(t, domain.Validate(invalid))
}
```

- [ ] **Step 3: Rodar teste e verificar que falha**

Run: `cd backend && go test ./internal/domain/ -run TestValidate -v`
Expected: FAIL — types not defined

- [ ] **Step 4: Implementar validate.go**

```go
package domain

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RegisterOwnerInput struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	Name         string `json:"name" validate:"required"`
	BusinessName string `json:"business_name" validate:"required"`
	Segment      string `json:"segment"`
	Phone        string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterSubscriberInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone"`
}

type SubscribeInput struct {
	PlanID int64 `json:"plan_id" validate:"required,gt=0"`
}

type ValidateUsageInput struct {
	BusinessSlug string `json:"business_slug" validate:"required"`
}

type CreatePlanInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents" validate:"required,gt=0"`
	LimitType   string `json:"limit_type" validate:"required,oneof=daily monthly"`
	LimitCount  int32  `json:"limit_count" validate:"required,gt=0"`
}

type UpdatePlanInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents" validate:"required,gt=0"`
	LimitType   string `json:"limit_type" validate:"required,oneof=daily monthly"`
	LimitCount  int32  `json:"limit_count" validate:"required,gt=0"`
}

type UpdateBusinessInput struct {
	Name    string `json:"name" validate:"required"`
	Segment string `json:"segment"`
	Address string `json:"address"`
	LogoURL string `json:"logo_url"`
}

type ApplyReferralInput struct {
	Code string `json:"code" validate:"required"`
}

type CancelBySubscriberInput struct {
	BusinessSlug string `json:"business_slug" validate:"required"`
}

// Validate validates a struct using validator/v10 tags.
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if vErrors, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("validacao: campo '%s' falhou na regra '%s'", vErrors[0].Field(), vErrors[0].Tag())
		}
		return fmt.Errorf("validacao: %w", err)
	}
	return nil
}
```

- [ ] **Step 5: Rodar teste e verificar que passa**

Run: `cd backend && go test ./internal/domain/ -run TestValidate -v`
Expected: PASS

- [ ] **Step 6: Integrar validator nos handlers de auth**

Em `backend/internal/handler/auth.go`:

1. Substituir `registerOwnerRequest` por uso de `domain.RegisterOwnerInput`
2. Substituir `loginRequest` por uso de `domain.LoginInput`
3. Substituir `registerSubscriberRequest` por uso de `domain.RegisterSubscriberInput`
4. Substituir validacao manual por `domain.Validate()`

O handler RegisterOwner fica:

```go
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterOwnerInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := domain.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Segment == "" {
		req.Segment = "cafeteria"
	}

	// ... rest stays the same, but use req.Email, req.Password, etc.
```

Fazer o mesmo para Login e RegisterSubscriber. Remover os tipos locais `registerOwnerRequest`, `loginRequest`, `registerSubscriberRequest`.

- [ ] **Step 7: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 8: Commit**

```bash
git add backend/internal/domain/validate.go backend/internal/domain/validate_test.go backend/internal/handler/auth.go backend/go.mod backend/go.sum
git commit -m "feat: validator/v10 para validacao de input com struct tags"
```

---

### Task 4: Busca de assinante por nome/telefone (fallback)

**Files:**
- Create: `backend/queries/search.sql`
- Modify: `backend/internal/repository/` (sqlc regenerate)
- Create: `backend/internal/handler/search.go`
- Create: `backend/internal/handler/search_test.go`
- Modify: `backend/cmd/api/router.go`

- [ ] **Step 1: Criar query SQL para busca**

```sql
-- name: SearchSubscribersByBusiness :many
SELECT u.id, u.name, u.email, u.phone,
       s.id as subscription_id, s.status,
       p.name as plan_name
FROM users u
JOIN subscriptions s ON s.subscriber_id = u.id
JOIN plans p ON p.id = s.plan_id
WHERE s.business_id = $1
  AND s.status IN ('active', 'grace')
  AND (u.name ILIKE '%' || $2 || '%' OR u.phone ILIKE '%' || $2 || '%')
ORDER BY u.name
LIMIT 20;
```

- [ ] **Step 2: Regenerar sqlc**

Run: `cd backend && sqlc generate`
Expected: No errors, new file generated in `internal/repository/search.sql.go`

- [ ] **Step 3: Escrever teste para o handler de busca**

```go
package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSubscribers_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner@test.com", "Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Test", "cafe-test")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano Cafe", 2990, "daily", 1)

	sub := testutil.SeedSubscriber(t, h.Queries, "maria@test.com", "Maria Santos", "11999990000")
	h.Queries.CreateSubscription(t.Context(), repository.CreateSubscriptionParams{
		PlanID:       plan.ID,
		SubscriberID: sub.ID,
		BusinessID:   biz.ID,
		Status:       "active",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=Maria", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp.Results, 1)
	assert.Equal(t, "Maria Santos", resp.Results[0].Name)
}

func TestSearchSubscribers_EmptyQuery(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner2@test.com", "Owner2")

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSearchSubscribers_NoResults(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner3@test.com", "Owner3")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Empty", "cafe-empty")

	req := httptest.NewRequest(http.MethodGet, "/api/subscribers/search?q=Ninguem", nil)
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.SearchSubscribers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Results []interface{} `json:"results"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp.Results, 0)
}
```

- [ ] **Step 4: Rodar teste e verificar que falha**

Run: `cd backend && go test ./internal/handler/ -run TestSearchSubscribers -v`
Expected: FAIL — `SearchSubscribers` not defined

- [ ] **Step 5: Implementar handler de busca**

```go
package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// SearchSubscribers searches for active subscribers by name or phone.
// GET /api/subscribers/search?q=xxx
func (h *Handler) SearchSubscribers(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "parametro 'q' e obrigatorio")
		return
	}

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negocio nao encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negocio")
		return
	}

	results, err := h.Queries.SearchSubscribersByBusiness(r.Context(), repository.SearchSubscribersByBusinessParams{
		BusinessID: biz.ID,
		Column2:    query,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinantes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
	})
}
```

- [ ] **Step 6: Adicionar rota ao router**

Em `backend/cmd/api/router.go`, dentro do grupo Owner routes:

```go
r.Get("/api/subscribers/search", h.SearchSubscribers)
```

- [ ] **Step 7: Rodar testes**

Run: `cd backend && go test ./internal/handler/ -run TestSearchSubscribers -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add backend/queries/search.sql backend/internal/repository/ backend/internal/handler/search.go backend/internal/handler/search_test.go backend/cmd/api/router.go
git commit -m "feat: busca de assinantes por nome/telefone (fallback validacao)"
```

---

### Task 5: Validacao de uso pelo dono (owner-side fallback)

**Files:**
- Create: `backend/internal/handler/owner_validate.go`
- Create: `backend/internal/handler/owner_validate_test.go`
- Modify: `backend/cmd/api/router.go`

- [ ] **Step 1: Escrever teste**

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateUsageByOwner_Success(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner@val.com", "Owner")
	biz := testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Val", "cafe-val")
	plan := testutil.SeedPlan(t, h.Queries, biz.ID, "Plano", 2990, "daily", 1)

	sub := testutil.SeedSubscriber(t, h.Queries, "sub@val.com", "Sub User", "")
	h.Queries.CreateSubscription(t.Context(), repository.CreateSubscriptionParams{
		PlanID:       plan.ID,
		SubscriberID: sub.ID,
		BusinessID:   biz.ID,
		Status:       "active",
	})

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
}

func TestValidateUsageByOwner_SubscriberNotFound(t *testing.T) {
	h := setupHandler(t)

	owner := testutil.SeedOwner(t, h.Queries, "owner@val2.com", "Owner2")
	testutil.SeedBusiness(t, h.Queries, owner.ID, "Cafe Val2", "cafe-val2")

	body, _ := json.Marshal(map[string]int64{"subscriber_id": 99999})
	req := httptest.NewRequest(http.MethodPost, "/api/validate-usage-owner", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, owner.ID, "owner")
	rr := httptest.NewRecorder()

	h.ValidateUsageByOwner(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
```

- [ ] **Step 2: Rodar teste e verificar que falha**

Run: `cd backend && go test ./internal/handler/ -run TestValidateUsageByOwner -v`
Expected: FAIL

- [ ] **Step 3: Implementar handler**

```go
package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// ValidateUsageByOwner allows the owner to validate usage on behalf of a subscriber (fallback).
// POST /api/validate-usage-owner
func (h *Handler) ValidateUsageByOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	var input struct {
		SubscriberID int64 `json:"subscriber_id"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if input.SubscriberID <= 0 {
		writeError(w, http.StatusBadRequest, "subscriber_id e obrigatorio")
		return
	}

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negocio nao encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negocio")
		return
	}

	sub, err := h.Queries.GetActiveSubscription(r.Context(), repository.GetActiveSubscriptionParams{
		SubscriberID: input.SubscriberID,
		BusinessID:   biz.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "assinatura ativa nao encontrada para este assinante")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinatura")
		return
	}

	plan, err := h.Queries.GetPlanByID(r.Context(), sub.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar plano")
		return
	}

	now := time.Now()
	var currentCount int64

	if plan.LimitType == domain.LimitTypeDaily {
		start, end := domain.DailyRange(now)
		currentCount, err = h.Queries.CountDailyUsage(r.Context(), repository.CountDailyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgTimestamptz(start),
			Column3:        pgTimestamptz(end),
		})
	} else {
		start, end := domain.MonthlyRange(now)
		currentCount, err = h.Queries.CountMonthlyUsage(r.Context(), repository.CountMonthlyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgTimestamptz(start),
			Column3:        pgTimestamptz(end),
		})
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao contar usos")
		return
	}

	if err := domain.CheckUsageLimit(plan.LimitType, int(plan.LimitCount), currentCount); err != nil {
		writeError(w, http.StatusForbidden, "limite de uso atingido para este periodo")
		return
	}

	_, err = h.Queries.CreateUsage(r.Context(), sub.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao registrar uso")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "validated",
		"used":      currentCount + 1,
		"limit":     plan.LimitCount,
		"plan_name": plan.Name,
	})
}
```

- [ ] **Step 4: Adicionar rota ao router**

Em `backend/cmd/api/router.go`, dentro do grupo Owner routes:

```go
r.Post("/api/validate-usage-owner", h.ValidateUsageByOwner)
```

- [ ] **Step 5: Rodar testes**

Run: `cd backend && go test ./internal/handler/ -run TestValidateUsageByOwner -v`
Expected: PASS

- [ ] **Step 6: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/owner_validate.go backend/internal/handler/owner_validate_test.go backend/cmd/api/router.go
git commit -m "feat: validacao de uso pelo dono (fallback busca por nome/telefone)"
```

---

### Task 6: Email dunning

**Files:**
- Create: `backend/internal/email/email.go`
- Create: `backend/internal/email/email_test.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/cron.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Escrever teste para email sender**

```go
package email_test

import (
	"testing"

	"github.com/clubepay/backend/internal/email"
	"github.com/stretchr/testify/assert"
)

func TestMockSender_RecordsMessages(t *testing.T) {
	mock := &email.MockSender{}

	err := mock.Send("user@test.com", "Assinatura bloqueada", "Sua assinatura foi bloqueada por falta de pagamento.")
	assert.NoError(t, err)
	assert.Len(t, mock.Sent, 1)
	assert.Equal(t, "user@test.com", mock.Sent[0].To)
	assert.Equal(t, "Assinatura bloqueada", mock.Sent[0].Subject)
}

func TestNewSMTP_ReturnsNilSenderWhenNotConfigured(t *testing.T) {
	sender := email.NewSMTP("", "", "", "")
	assert.Nil(t, sender)
}

func TestNewSMTP_ReturnsSenderWhenConfigured(t *testing.T) {
	sender := email.NewSMTP("smtp.test.com", "587", "user", "pass")
	assert.NotNil(t, sender)
}
```

- [ ] **Step 2: Rodar teste e verificar que falha**

Run: `cd backend && go test ./internal/email/ -v`
Expected: FAIL — package not found

- [ ] **Step 3: Implementar email package**

```go
package email

import (
	"fmt"
	"net/smtp"
)

// Sender defines the interface for sending emails.
type Sender interface {
	Send(to, subject, body string) error
}

// SMTPSender sends emails via SMTP.
type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
}

// NewSMTP creates a new SMTP sender. Returns nil if host is empty.
func NewSMTP(host, port, username, password string) *SMTPSender {
	if host == "" {
		return nil
	}
	return &SMTPSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	from := s.Username
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	addr := s.Host + ":" + s.Port

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// MockSender records sent messages for testing.
type MockSender struct {
	Sent []Message
}

type Message struct {
	To      string
	Subject string
	Body    string
}

func (m *MockSender) Send(to, subject, body string) error {
	m.Sent = append(m.Sent, Message{To: to, Subject: subject, Body: body})
	return nil
}
```

- [ ] **Step 4: Rodar testes**

Run: `cd backend && go test ./internal/email/ -v`
Expected: PASS

- [ ] **Step 5: Adicionar config SMTP**

Em `backend/internal/config/config.go`, adicionar campos:

```go
SMTPHost     string
SMTPPort     string
SMTPUsername string
SMTPPassword string
```

E no Load():

```go
cfg.SMTPHost = os.Getenv("SMTP_HOST")
cfg.SMTPPort = getEnv("SMTP_PORT", "587")
cfg.SMTPUsername = os.Getenv("SMTP_USERNAME")
cfg.SMTPPassword = os.Getenv("SMTP_PASSWORD")
```

- [ ] **Step 6: Adicionar EmailSender ao Handler**

Em `backend/internal/handler/handler.go`, adicionar campo `Email email.Sender` ao struct Handler. Atualizar `New()`:

```go
func New(q *repository.Queries, cfg *config.Config, p psp.PSP, e email.Sender) *Handler {
	return &Handler{Queries: q, Config: cfg, PSP: p, Email: e}
}
```

- [ ] **Step 7: Atualizar main.go para injetar email sender**

Em `backend/cmd/api/main.go`:

```go
import "github.com/clubepay/backend/internal/email"

// After PSP client setup:
var emailSender email.Sender
smtpSender := email.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
if smtpSender != nil {
    emailSender = smtpSender
    slog.Info("using SMTP email sender")
} else {
    emailSender = &email.MockSender{}
    slog.Warn("SMTP not configured, using mock email sender")
}

h := handler.New(queries, cfg, pspClient, emailSender)
```

- [ ] **Step 8: Atualizar setupHandler nos testes**

Em todos os test files que usam `setupHandler`, atualizar para passar `&email.MockSender{}`:

```go
func setupHandler(t *testing.T) *handler.Handler {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	cfg := &config.Config{JWTSecret: "test-secret-key"}
	mockPSP := &psp.MockPSP{}
	mockEmail := &email.MockSender{}
	return handler.New(queries, cfg, mockPSP, mockEmail)
}
```

Nota: `setupHandler` esta definido em `auth_test.go`. Atualizar la e todos os testes vao herdar.

- [ ] **Step 9: Modificar cron.go para enviar email ao bloquear**

No handler Reconcile, apos bloquear uma subscription com sucesso, buscar o subscriber e enviar email:

```go
// Inside the graceExpired loop, after successful block:
if h.Email != nil {
    subscriber, err := h.Queries.GetUserByID(ctx, sub.SubscriberID)
    if err == nil {
        h.Email.Send(
            subscriber.Email,
            "ClubePay - Assinatura bloqueada",
            fmt.Sprintf("Ola %s,\n\nSua assinatura foi bloqueada por falta de pagamento.\nPor favor, regularize seu pagamento para continuar usando o servico.\n\nEquipe ClubePay", subscriber.Name),
        )
    }
}
```

- [ ] **Step 10: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 11: Commit**

```bash
git add backend/internal/email/ backend/internal/config/config.go backend/internal/handler/handler.go backend/internal/handler/cron.go backend/cmd/api/main.go backend/internal/handler/auth_test.go
git commit -m "feat: email dunning — notifica assinante quando assinatura e bloqueada"
```

---

### Task 7: QR Code no dashboard

**Files:**
- Modify: `frontend/package.json` (add qrcode.react)
- Create: `frontend/src/components/QRCode.tsx`
- Modify: `frontend/src/app/(auth)/dashboard/page.tsx`

- [ ] **Step 1: Instalar dependencia**

Run: `cd frontend && npm install qrcode.react`

- [ ] **Step 2: Criar componente QRCode**

```tsx
"use client";

import { QRCodeSVG } from "qrcode.react";

interface QRCodeProps {
  slug: string;
}

export default function BusinessQRCode({ slug }: QRCodeProps) {
  const url =
    typeof window !== "undefined"
      ? `${window.location.origin}/validar/${slug}`
      : `/validar/${slug}`;

  function handlePrint() {
    const printWindow = window.open("", "_blank");
    if (!printWindow) return;
    printWindow.document.write(`
      <html>
        <head><title>QR Code - ClubePay</title></head>
        <body style="display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:100vh;font-family:system-ui">
          <h2 style="color:#2a7d6e">Escaneie para validar seu uso</h2>
          <img src="https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=${encodeURIComponent(url)}" alt="QR Code" />
          <p style="margin-top:16px;color:#666">${url}</p>
          <script>window.print();window.close();</script>
        </body>
      </html>
    `);
    printWindow.document.close();
  }

  return (
    <div className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col items-center gap-4">
      <h2 className="font-semibold text-gray-800">QR Code do balcao</h2>
      <p className="text-sm text-gray-500 text-center">
        Imprima e coloque no balcao. Assinantes escaneiam para validar o uso.
      </p>
      <QRCodeSVG value={url} size={200} fgColor="#2a7d6e" />
      <p className="text-xs text-gray-400 break-all text-center">{url}</p>
      <button
        onClick={handlePrint}
        className="rounded-xl px-4 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90"
        style={{ backgroundColor: "#d4a853" }}
      >
        Imprimir QR Code
      </button>
    </div>
  );
}
```

- [ ] **Step 3: Adicionar QR Code ao dashboard**

Em `frontend/src/app/(auth)/dashboard/page.tsx`:

1. Importar: `import BusinessQRCode from "@/components/QRCode";`
2. Adicionar apos os StatsCards e antes da lista de assinantes:

```tsx
{business && <BusinessQRCode slug={business.slug} />}
```

- [ ] **Step 4: Verificar build**

Run: `cd frontend && npx next build`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/components/QRCode.tsx frontend/src/app/\(auth\)/dashboard/page.tsx
git commit -m "feat: QR code do balcao no dashboard com botao de imprimir"
```

---

### Task 8: Pagina de gestao de planos

**Files:**
- Create: `frontend/src/app/(auth)/planos/page.tsx`
- Modify: `frontend/src/app/(auth)/dashboard/page.tsx` (add link)

- [ ] **Step 1: Criar pagina de gestao de planos**

```tsx
"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";

interface Plan {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  limit_type: "daily" | "monthly";
  limit_count: number;
  active: boolean;
}

interface PlansResponse {
  plans: Plan[];
}

function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

export default function PlanosPage() {
  const router = useRouter();
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    price_cents: "",
    limit_type: "daily" as "daily" | "monthly",
    limit_count: "1",
  });
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState("");

  const fetchPlans = useCallback(async () => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    try {
      const data = await api.get<PlansResponse>("/api/plans", token);
      setPlans(data.plans ?? []);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.push("/login");
      } else {
        setError("Erro ao carregar planos.");
      }
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    fetchPlans();
  }, [fetchPlans]);

  function resetForm() {
    setFormData({
      name: "",
      description: "",
      price_cents: "",
      limit_type: "daily",
      limit_count: "1",
    });
    setEditingPlan(null);
    setShowForm(false);
    setFormError("");
  }

  function handleEdit(plan: Plan) {
    setEditingPlan(plan);
    setFormData({
      name: plan.name,
      description: plan.description || "",
      price_cents: (plan.price_cents / 100).toString(),
      limit_type: plan.limit_type,
      limit_count: plan.limit_count.toString(),
    });
    setShowForm(true);
    setFormError("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token) return;

    setSubmitting(true);
    setFormError("");

    const priceCents = Math.round(parseFloat(formData.price_cents) * 100);
    if (isNaN(priceCents) || priceCents <= 0) {
      setFormError("Preco invalido.");
      setSubmitting(false);
      return;
    }

    const limitCount = parseInt(formData.limit_count, 10);
    if (isNaN(limitCount) || limitCount <= 0) {
      setFormError("Limite invalido.");
      setSubmitting(false);
      return;
    }

    const payload = {
      name: formData.name,
      description: formData.description,
      price_cents: priceCents,
      limit_type: formData.limit_type,
      limit_count: limitCount,
    };

    try {
      if (editingPlan) {
        await api.put(`/api/plans/${editingPlan.id}`, payload, token);
      } else {
        await api.post("/api/plans", payload, token);
      }
      resetForm();
      fetchPlans();
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message);
      } else {
        setFormError("Erro ao salvar plano.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(planId: number) {
    const confirmed = window.confirm("Tem certeza que deseja desativar este plano?");
    if (!confirmed) return;

    const token = getToken();
    if (!token) return;

    try {
      await api.del(`/api/plans/${planId}`, token);
      fetchPlans();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      }
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-4">
        <div className="mx-auto max-w-2xl flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900">Meus Planos</h1>
          <Link
            href="/dashboard"
            className="text-sm font-medium transition-colors hover:opacity-80"
            style={{ color: "#2a7d6e" }}
          >
            Voltar ao Dashboard
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-2xl px-4 py-8 flex flex-col gap-6">
        {error && (
          <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {!showForm && (
          <button
            onClick={() => { resetForm(); setShowForm(true); }}
            className="w-full rounded-2xl py-4 text-lg font-bold text-white transition-opacity hover:opacity-90 shadow-md"
            style={{ backgroundColor: "#2a7d6e" }}
          >
            + Criar novo plano
          </button>
        )}

        {showForm && (
          <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col gap-4">
            <h2 className="font-semibold text-gray-800">
              {editingPlan ? "Editar plano" : "Novo plano"}
            </h2>

            {formError && (
              <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
                {formError}
              </div>
            )}

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700">Nome</label>
              <input
                type="text"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                placeholder="Ex: Cafe Diario"
              />
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700">Descricao (opcional)</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                placeholder="1 cafe por dia"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1">
                <label className="text-sm font-medium text-gray-700">Preco (R$)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  required
                  value={formData.price_cents}
                  onChange={(e) => setFormData({ ...formData, price_cents: e.target.value })}
                  className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                  placeholder="29.90"
                />
              </div>

              <div className="flex flex-col gap-1">
                <label className="text-sm font-medium text-gray-700">Tipo de limite</label>
                <select
                  value={formData.limit_type}
                  onChange={(e) => setFormData({ ...formData, limit_type: e.target.value as "daily" | "monthly" })}
                  className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                >
                  <option value="daily">Diario</option>
                  <option value="monthly">Mensal</option>
                </select>
              </div>
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700">
                Limite ({formData.limit_type === "daily" ? "usos por dia" : "usos por mes"})
              </label>
              <input
                type="number"
                min="1"
                required
                value={formData.limit_count}
                onChange={(e) => setFormData({ ...formData, limit_count: e.target.value })}
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
              />
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={submitting}
                className="flex-1 rounded-xl py-3 font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60"
                style={{ backgroundColor: "#2a7d6e" }}
              >
                {submitting ? "Salvando..." : editingPlan ? "Atualizar" : "Criar plano"}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="rounded-xl px-6 py-3 font-semibold text-gray-600 bg-gray-100 hover:bg-gray-200 transition-colors"
              >
                Cancelar
              </button>
            </div>
          </form>
        )}

        {plans.length === 0 && !showForm ? (
          <div className="text-center py-12 text-gray-500">
            Nenhum plano criado ainda. Crie seu primeiro plano!
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {plans.map((plan) => (
              <div
                key={plan.id}
                className="bg-white rounded-2xl border border-gray-200 p-6 flex items-center justify-between"
              >
                <div>
                  <h3 className="font-semibold text-gray-900">{plan.name}</h3>
                  {plan.description && (
                    <p className="text-sm text-gray-500">{plan.description}</p>
                  )}
                  <p className="text-sm text-gray-600 mt-1">
                    {formatPrice(plan.price_cents)}/mes — {plan.limit_count}x{" "}
                    {plan.limit_type === "daily" ? "por dia" : "por mes"}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleEdit(plan)}
                    className="rounded-lg px-3 py-1.5 text-sm font-medium border border-gray-300 text-gray-600 hover:bg-gray-50 transition-colors"
                  >
                    Editar
                  </button>
                  <button
                    onClick={() => handleDelete(plan.id)}
                    className="rounded-lg px-3 py-1.5 text-sm font-medium border border-red-200 text-red-600 hover:bg-red-50 transition-colors"
                  >
                    Desativar
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
```

- [ ] **Step 2: Adicionar link no dashboard**

Em `frontend/src/app/(auth)/dashboard/page.tsx`, adicionar um link para `/planos` no header ou abaixo dos stats:

```tsx
<Link
  href="/planos"
  className="block w-full rounded-2xl py-4 text-center text-lg font-bold border-2 transition-colors hover:bg-gray-50"
  style={{ borderColor: "#2a7d6e", color: "#2a7d6e" }}
>
  Gerenciar planos
</Link>
```

- [ ] **Step 3: Verificar build**

Run: `cd frontend && npx next build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/\(auth\)/planos/page.tsx frontend/src/app/\(auth\)/dashboard/page.tsx
git commit -m "feat: pagina de gestao de planos (criar, editar, desativar)"
```

---

### Task 9: Testes frontend (Vitest)

**Files:**
- Create: `frontend/src/lib/__tests__/api.test.ts`
- Create: `frontend/src/lib/__tests__/auth.test.ts`
- Create: `frontend/src/components/__tests__/StatsCard.test.tsx`
- Create: `frontend/src/components/__tests__/PlanCard.test.tsx`
- Create: `frontend/src/components/__tests__/UsageBar.test.tsx`

- [ ] **Step 1: Escrever testes para api.ts**

```typescript
import { describe, it, expect, vi, beforeEach } from "vitest";
import { api, ApiError } from "@/lib/api";

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

describe("api.get", () => {
  it("makes GET request and returns data", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: "test" }),
    });

    const result = await api.get("/api/test");
    expect(result).toEqual({ data: "test" });
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/test"),
      expect.objectContaining({ method: "GET" })
    );
  });

  it("sends authorization header when token provided", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.get("/api/test", "my-token");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].headers["Authorization"]).toBe("Bearer my-token");
  });

  it("throws ApiError on non-ok response", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      statusText: "Not Found",
      json: () => Promise.resolve({ message: "not found" }),
    });

    await expect(api.get("/api/missing")).rejects.toThrow(ApiError);
    try {
      await api.get("/api/missing");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(404);
    }
  });
});

describe("api.post", () => {
  it("makes POST request with body", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 1 }),
    });

    const result = await api.post("/api/items", { name: "test" });
    expect(result).toEqual({ id: 1 });

    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("POST");
    expect(callArgs[1].body).toBe(JSON.stringify({ name: "test" }));
  });
});

describe("api.put", () => {
  it("makes PUT request", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ updated: true }),
    });

    await api.put("/api/items/1", { name: "updated" }, "token");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("PUT");
  });
});

describe("api.del", () => {
  it("makes DELETE request", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.del("/api/items/1", "token");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("DELETE");
  });
});
```

- [ ] **Step 2: Escrever testes para auth.ts**

```typescript
import { describe, it, expect, beforeEach } from "vitest";
import { getToken, setToken, clearToken } from "@/lib/auth";

beforeEach(() => {
  localStorage.clear();
});

describe("auth", () => {
  it("setToken stores token in localStorage", () => {
    setToken("abc123");
    expect(localStorage.getItem("clubepay_token")).toBe("abc123");
  });

  it("getToken retrieves token from localStorage", () => {
    localStorage.setItem("clubepay_token", "xyz789");
    expect(getToken()).toBe("xyz789");
  });

  it("getToken returns null when no token", () => {
    expect(getToken()).toBeNull();
  });

  it("clearToken removes token from localStorage", () => {
    setToken("token");
    clearToken();
    expect(getToken()).toBeNull();
  });
});
```

- [ ] **Step 3: Rodar testes e verificar que passam**

Run: `cd frontend && npx vitest run src/lib/`
Expected: PASS

- [ ] **Step 4: Ler componentes existentes para escrever testes**

Ler: `frontend/src/components/StatsCard.tsx`, `PlanCard.tsx`, `UsageBar.tsx`

- [ ] **Step 5: Escrever testes para StatsCard**

```tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import StatsCard from "@/components/StatsCard";

describe("StatsCard", () => {
  it("renders label and value", () => {
    render(<StatsCard label="Assinantes ativos" value={42} />);
    expect(screen.getByText("Assinantes ativos")).toBeDefined();
    expect(screen.getByText("42")).toBeDefined();
  });

  it("renders string values", () => {
    render(<StatsCard label="MRR" value="R$ 150,00" />);
    expect(screen.getByText("MRR")).toBeDefined();
    expect(screen.getByText("R$ 150,00")).toBeDefined();
  });
});
```

- [ ] **Step 6: Escrever testes para UsageBar**

```tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import UsageBar from "@/components/UsageBar";

describe("UsageBar", () => {
  it("renders used and limit counts", () => {
    render(<UsageBar used={2} limit={4} period="monthly" />);
    expect(screen.getByText(/2/)).toBeDefined();
    expect(screen.getByText(/4/)).toBeDefined();
  });

  it("renders daily period text", () => {
    render(<UsageBar used={0} limit={1} period="daily" />);
    expect(screen.getByText(/hoje/i) || screen.getByText(/dia/i)).toBeDefined();
  });
});
```

- [ ] **Step 7: Escrever testes para PlanCard**

```tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import PlanCard from "@/components/PlanCard";

describe("PlanCard", () => {
  it("renders plan name and price", () => {
    render(
      <PlanCard
        name="Cafe Diario"
        priceCents={2990}
        limitType="daily"
        limitCount={1}
        onSubscribe={() => {}}
      />
    );
    expect(screen.getByText("Cafe Diario")).toBeDefined();
    expect(screen.getByText(/29,90/)).toBeDefined();
  });

  it("renders subscribe button", () => {
    render(
      <PlanCard
        name="Plano Mensal"
        priceCents={5000}
        limitType="monthly"
        limitCount={4}
        onSubscribe={() => {}}
      />
    );
    expect(screen.getByRole("button")).toBeDefined();
  });
});
```

- [ ] **Step 8: Rodar todos os testes frontend**

Run: `cd frontend && npx vitest run`
Expected: All PASS

- [ ] **Step 9: Commit**

```bash
git add frontend/src/lib/__tests__/ frontend/src/components/__tests__/
git commit -m "test: testes Vitest para api, auth, StatsCard, PlanCard, UsageBar"
```

---

## Parallelization Guide

Tasks that can run in parallel (independent of each other):

**Group 1 (backend infra):** Task 1, Task 2, Task 3
**Group 2 (backend features):** Task 4, Task 5, Task 6 (Task 6 depends on handler.go changes from earlier)
**Group 3 (frontend):** Task 7, Task 8, Task 9

Within each group, tasks are independent and can be parallelized. Between groups, Task 6 (email dunning) modifies handler.go which may conflict with other handler changes, so merge carefully.
