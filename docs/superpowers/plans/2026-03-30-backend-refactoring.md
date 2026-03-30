# Backend Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refatorar o backend Go do ClubePay para introduzir service layer, eliminar duplicacao, tipar respostas JSON e unificar validacao — em 4 fases incrementais onde cada fase compila e passa nos testes.

**Architecture:** Introduzir `internal/service/` entre handlers e repository. Handlers ficam finos (parse request, chama service, write response). Services encapsulam logica de negocio, PSP, email. Domain centraliza constantes, inputs tipados, responses e erros de servico.

**Tech Stack:** Go 1.22+, chi, sqlc, pgx, testify, testcontainers-go, validator/v10

**Spec:** `docs/superpowers/specs/2026-03-30-backend-refactoring-design.md`

---

## Fase 1: Fundacoes

Novos arquivos que nao mudam comportamento existente. Tudo compila, testes passam.

---

### Task 1: Constantes de negocio centralizadas

**Files:**
- Create: `internal/domain/constants.go`
- Modify: `internal/handler/plan.go:15` (remove `freeTierPlanLimit`)
- Modify: `internal/handler/subscription.go:21` (remove `freeTierSubscriberLimit`)
- Modify: `internal/handler/referral.go:15` (remove `referralLimit`)

- [ ] **Step 1: Criar `internal/domain/constants.go`**

```go
package domain

import "time"

// Business tier limits (free tier).
const (
	FreeTierPlanLimit       = 1
	FreeTierSubscriberLimit = 15
)

// Referral limits and discounts.
const (
	ReferralLimit           = 3
	ReferralDiscountPercent = 10
)

// Grace period for overdue payments.
const GracePeriodDays = 3

// JWT expiry durations.
const (
	OwnerJWTExpiry      = 24 * time.Hour
	SubscriberJWTExpiry = 30 * 24 * time.Hour
)
```

- [ ] **Step 2: Atualizar `handler/plan.go` para usar `domain.FreeTierPlanLimit`**

Remove a linha `const freeTierPlanLimit = 1` (linha 15) e substitui a referencia na linha 67:
```go
// antes: if count >= int64(freeTierPlanLimit) {
// depois:
if count >= int64(domain.FreeTierPlanLimit) {
```

Adicionar import `"github.com/clubepay/backend/internal/domain"` se nao existir.

- [ ] **Step 3: Atualizar `handler/subscription.go` para usar `domain.FreeTierSubscriberLimit`**

Remove `const freeTierSubscriberLimit = 15` (linha 21) e atualiza linha 76:
```go
if count >= int64(domain.FreeTierSubscriberLimit) {
```

- [ ] **Step 4: Atualizar `handler/referral.go` para usar `domain.ReferralLimit`**

Remove `const referralLimit = 3` (linha 15) e atualiza linha 122:
```go
if count > int64(domain.ReferralLimit) {
```

- [ ] **Step 5: Atualizar `handler/auth.go` para usar `domain.OwnerJWTExpiry` e `domain.SubscriberJWTExpiry`**

Linha 71:
```go
// antes: token, err := domain.GenerateJWT(user.ID, domain.RoleOwner, h.Config.JWTSecret, 24*time.Hour)
// depois:
token, err := domain.GenerateJWT(user.ID, domain.RoleOwner, h.Config.JWTSecret, domain.OwnerJWTExpiry)
```

Linha 123-126:
```go
// antes:
// expiry := 24 * time.Hour
// if user.Role == domain.RoleSubscriber {
//     expiry = 30 * 24 * time.Hour
// }
// depois:
expiry := domain.OwnerJWTExpiry
if user.Role == domain.RoleSubscriber {
    expiry = domain.SubscriberJWTExpiry
}
```

Linha 181:
```go
// antes: token, err := domain.GenerateJWT(user.ID, domain.RoleSubscriber, h.Config.JWTSecret, 30*24*time.Hour)
// depois:
token, err := domain.GenerateJWT(user.ID, domain.RoleSubscriber, h.Config.JWTSecret, domain.SubscriberJWTExpiry)
```

- [ ] **Step 6: Atualizar `handler/webhook.go` para usar `domain.GracePeriodDays`**

Linha 105:
```go
// antes: graceDeadline := time.Now().AddDate(0, 0, 3)
// depois:
graceDeadline := time.Now().AddDate(0, 0, domain.GracePeriodDays)
```

- [ ] **Step 7: Atualizar `handler/subscription.go` para usar `domain.ReferralDiscountPercent`**

Linha 107:
```go
// antes: discountPercent = 10
// depois:
discountPercent = int32(domain.ReferralDiscountPercent)
```

- [ ] **Step 8: Verificar compilacao e testes**

Run: `cd backend && go build ./... && go test ./...`
Expected: BUILD OK, ALL TESTS PASS

- [ ] **Step 9: Commit**

```bash
cd backend && git add internal/domain/constants.go internal/handler/plan.go internal/handler/subscription.go internal/handler/referral.go internal/handler/auth.go internal/handler/webhook.go
git commit -m "refactor: centralizar constantes de negocio em domain/constants.go"
```

---

### Task 2: Input structs adicionais com validator tags

**Files:**
- Modify: `internal/domain/validate.go` (adicionar novos input types)

Nota: `validate.go` ja tem RegisterOwnerInput, LoginInput, RegisterSubscriberInput, SubscribeInput, ValidateUsageInput, CreatePlanInput, UpdatePlanInput, UpdateBusinessInput, ApplyReferralInput. Faltam os inputs usados como structs anonimas nos handlers.

- [ ] **Step 1: Escrever teste para os novos input types**

Adicionar ao arquivo `internal/domain/validate_test.go`:

```go
func TestValidate_UpdateProfileInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.UpdateProfileInput
		wantErr bool
	}{
		{"valid", domain.UpdateProfileInput{Name: "João"}, false},
		{"valid with phone", domain.UpdateProfileInput{Name: "João", Phone: "11999999999"}, false},
		{"missing name", domain.UpdateProfileInput{Name: ""}, true},
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

func TestValidate_ChangePasswordInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.ChangePasswordInput
		wantErr bool
	}{
		{"valid", domain.ChangePasswordInput{CurrentPassword: "old12345", NewPassword: "new12345"}, false},
		{"missing current", domain.ChangePasswordInput{CurrentPassword: "", NewPassword: "new12345"}, true},
		{"short new password", domain.ChangePasswordInput{CurrentPassword: "old12345", NewPassword: "short"}, true},
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

func TestValidate_RequestPasswordResetInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.RequestPasswordResetInput
		wantErr bool
	}{
		{"valid", domain.RequestPasswordResetInput{Email: "a@b.com"}, false},
		{"missing email", domain.RequestPasswordResetInput{Email: ""}, true},
		{"invalid email", domain.RequestPasswordResetInput{Email: "notanemail"}, true},
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

func TestValidate_ConfirmPasswordResetInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.ConfirmPasswordResetInput
		wantErr bool
	}{
		{"valid", domain.ConfirmPasswordResetInput{Token: "abc123", Password: "newpass12"}, false},
		{"missing token", domain.ConfirmPasswordResetInput{Token: "", Password: "newpass12"}, true},
		{"short password", domain.ConfirmPasswordResetInput{Token: "abc123", Password: "short"}, true},
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

func TestValidate_ValidateUsageOwnerInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.ValidateUsageOwnerInput
		wantErr bool
	}{
		{"valid", domain.ValidateUsageOwnerInput{SubscriberID: 1}, false},
		{"zero subscriber_id", domain.ValidateUsageOwnerInput{SubscriberID: 0}, true},
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
```

- [ ] **Step 2: Rodar testes para ver falhar**

Run: `cd backend && go test ./internal/domain/ -run "TestValidate_(UpdateProfile|ChangePassword|RequestPasswordReset|ConfirmPasswordReset|ValidateUsageOwner)" -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Adicionar input types em `domain/validate.go`**

Adicionar ao final do arquivo, antes da funcao `Validate`:

```go
type UpdateProfileInput struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type RequestPasswordResetInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ConfirmPasswordResetInput struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type ValidateUsageOwnerInput struct {
	SubscriberID int64 `json:"subscriber_id" validate:"gt=0"`
}
```

- [ ] **Step 4: Rodar testes para ver passar**

Run: `cd backend && go test ./internal/domain/ -run "TestValidate_(UpdateProfile|ChangePassword|RequestPasswordReset|ConfirmPasswordReset|ValidateUsageOwner)" -v`
Expected: ALL PASS

- [ ] **Step 5: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
cd backend && git add internal/domain/validate.go internal/domain/validate_test.go
git commit -m "feat: input structs adicionais com validator tags"
```

---

### Task 3: Response structs tipadas

**Files:**
- Create: `internal/domain/responses.go`

- [ ] **Step 1: Criar `internal/domain/responses.go`**

```go
package domain

import "time"

// AuthResponse is returned by register and login endpoints.
type AuthResponse struct {
	Token    string            `json:"token"`
	User     UserResponse      `json:"user"`
	Business *BusinessResponse `json:"business,omitempty"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
	Role  string `json:"role"`
}

// BusinessResponse is the public representation of a business.
type BusinessResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Segment string `json:"segment"`
	Address string `json:"address,omitempty"`
	LogoURL string `json:"logo_url,omitempty"`
}

// PlanDetailResponse includes full plan info.
type PlanDetailResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	LimitType   string `json:"limit_type"`
	LimitCount  int32  `json:"limit_count"`
	Active      bool   `json:"active"`
}

// PlanListResponse wraps a list of plans.
type PlanListResponse struct {
	Plans []PlanDetailResponse `json:"plans"`
}

// SubscriptionListResponse wraps a list of subscriptions.
type SubscriptionListResponse struct {
	Subscriptions interface{} `json:"subscriptions"`
}

// SubscriptionInfo is a lightweight subscription summary.
type SubscriptionInfo struct {
	ID        int64      `json:"id"`
	Status    string     `json:"status"`
	PeriodEnd *time.Time `json:"period_end,omitempty"`
}

// MyPlanResponse combines plan, business, and subscription info for the subscriber.
type MyPlanResponse struct {
	Plan         PlanDetailResponse `json:"plan"`
	Business     BusinessResponse   `json:"business"`
	Subscription SubscriptionInfo   `json:"subscription"`
}

// ValidateUsageResponse is returned after validating a usage.
type ValidateUsageResponse struct {
	Status   string `json:"status"`
	Used     int64  `json:"used"`
	Limit    int32  `json:"limit"`
	PlanName string `json:"plan_name"`
}

// UsageListResponse is returned by my-usage.
type UsageListResponse struct {
	Used     int         `json:"used"`
	Limit    int32       `json:"limit"`
	PlanName string      `json:"plan_name"`
	Usages   interface{} `json:"usages"`
}

// CancelResponse is returned after cancelling a subscription.
type CancelResponse struct {
	Status    string     `json:"status"`
	PeriodEnd *time.Time `json:"period_end,omitempty"`
	Message   string     `json:"message"`
}

// ReconcileResponse is returned by the cron reconcile endpoint.
type ReconcileResponse struct {
	Blocked int `json:"blocked"`
	Synced  int `json:"synced"`
}

// ProfileResponse wraps a user profile.
type ProfileResponse struct {
	User UserResponse `json:"user"`
}

// SearchResponse wraps search results.
type SearchResponse struct {
	Results interface{} `json:"results"`
}

// ReferralCodeResponse returns the referral code.
type ReferralCodeResponse struct {
	Code string `json:"code"`
}

// PublicBusinessResponse includes subscriber count.
type PublicBusinessResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Segment         string `json:"segment"`
	SubscriberCount int64  `json:"subscriber_count"`
}
```

- [ ] **Step 2: Verificar compilacao**

Run: `cd backend && go build ./...`
Expected: BUILD OK

- [ ] **Step 3: Commit**

```bash
cd backend && git add internal/domain/responses.go
git commit -m "feat: response structs tipadas em domain/responses.go"
```

---

### Task 4: Service errors tipados

**Files:**
- Create: `internal/domain/service_errors.go`
- Create: `internal/domain/service_errors_test.go`

- [ ] **Step 1: Escrever teste**

Criar `internal/domain/service_errors_test.go`:

```go
package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clubepay/backend/internal/domain"
)

func TestServiceError_Error(t *testing.T) {
	err := domain.NewServiceError(400, "bad input", nil)
	assert.Equal(t, "bad input", err.Error())
}

func TestServiceError_Unwrap(t *testing.T) {
	cause := errors.New("db connection failed")
	err := domain.NewServiceError(500, "internal error", cause)
	assert.ErrorIs(t, err, cause)
}

func TestServiceError_ErrorsAs(t *testing.T) {
	err := domain.ErrNotFound("nao encontrado")
	var svcErr *domain.ServiceError
	assert.True(t, errors.As(err, &svcErr))
	assert.Equal(t, 404, svcErr.Code)
}

func TestServiceError_Constructors(t *testing.T) {
	tests := []struct {
		name string
		err  *domain.ServiceError
		code int
	}{
		{"not found", domain.ErrNotFound("x"), 404},
		{"conflict", domain.ErrConflict("x"), 409},
		{"forbidden", domain.ErrForbidden("x"), 403},
		{"bad request", domain.ErrBadRequest("x"), 400},
		{"internal", domain.ErrInternal("x", nil), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, "x", tt.err.Message)
		})
	}
}
```

- [ ] **Step 2: Rodar teste para ver falhar**

Run: `cd backend && go test ./internal/domain/ -run TestServiceError -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Criar `internal/domain/service_errors.go`**

```go
package domain

import "net/http"

// ServiceError is a typed error that carries an HTTP status code.
type ServiceError struct {
	Code    int
	Message string
	Err     error
}

func NewServiceError(code int, message string, err error) *ServiceError {
	return &ServiceError{Code: code, Message: message, Err: err}
}

func (e *ServiceError) Error() string { return e.Message }

func (e *ServiceError) Unwrap() error { return e.Err }

// Constructors for common error types.

func ErrNotFound(msg string) *ServiceError {
	return NewServiceError(http.StatusNotFound, msg, nil)
}

func ErrConflict(msg string) *ServiceError {
	return NewServiceError(http.StatusConflict, msg, nil)
}

func ErrForbidden(msg string) *ServiceError {
	return NewServiceError(http.StatusForbidden, msg, nil)
}

func ErrBadRequest(msg string) *ServiceError {
	return NewServiceError(http.StatusBadRequest, msg, nil)
}

func ErrInternal(msg string, err error) *ServiceError {
	return NewServiceError(http.StatusInternalServerError, msg, err)
}
```

- [ ] **Step 4: Rodar testes para ver passar**

Run: `cd backend && go test ./internal/domain/ -run TestServiceError -v`
Expected: ALL PASS

- [ ] **Step 5: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
cd backend && git add internal/domain/service_errors.go internal/domain/service_errors_test.go
git commit -m "feat: service errors tipados com HTTP status codes"
```

---

### Task 5: handleServiceError helper no handler

**Files:**
- Modify: `internal/handler/handler.go`

- [ ] **Step 1: Adicionar `handleServiceError` em `handler.go`**

Adicionar ao final do arquivo:

```go
func handleServiceError(w http.ResponseWriter, err error) {
	var svcErr *domain.ServiceError
	if errors.As(err, &svcErr) {
		writeError(w, svcErr.Code, svcErr.Message)
		return
	}
	slog.Error("unhandled service error", "error", err)
	writeError(w, http.StatusInternalServerError, "erro interno")
}
```

Adicionar imports: `"errors"` e `"github.com/clubepay/backend/internal/domain"`.

- [ ] **Step 2: Verificar compilacao e testes**

Run: `cd backend && go build ./... && go test ./...`
Expected: BUILD OK, ALL PASS

- [ ] **Step 3: Commit**

```bash
cd backend && git add internal/handler/handler.go
git commit -m "feat: handleServiceError helper para mapear ServiceError -> HTTP"
```

---

## Fase 2: Service Layer

Cria `internal/service/` e migra logica de negocio dos handlers.

---

### Task 6: BusinessService

**Files:**
- Create: `internal/service/business.go`
- Create: `internal/service/business_test.go`

- [ ] **Step 1: Escrever teste do BusinessService**

Criar `internal/service/business_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
	"github.com/clubepay/backend/internal/testutil"
)

func TestBusinessService_GetByOwner(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewBusinessService(queries)

	owner := testutil.SeedOwner(t, queries, "biz-svc@test.com", "Owner")
	testutil.SeedBusiness(t, queries, owner.ID, "My Cafe", "my-cafe")

	t.Run("success", func(t *testing.T) {
		resp, err := svc.GetByOwner(context.Background(), owner.ID)
		require.NoError(t, err)
		assert.Equal(t, "My Cafe", resp.Name)
		assert.Equal(t, "my-cafe", resp.Slug)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetByOwner(context.Background(), 999999)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.ErrorAs(t, err, &svcErr)
		assert.Equal(t, 404, svcErr.Code)
	})
}

func TestBusinessService_Update(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewBusinessService(queries)

	owner := testutil.SeedOwner(t, queries, "biz-upd@test.com", "Owner")
	testutil.SeedBusiness(t, queries, owner.ID, "Old Name", "old-name")

	resp, err := svc.Update(context.Background(), owner.ID, domain.UpdateBusinessInput{
		Name:    "New Name",
		Segment: "padaria",
		Address: "Rua Nova 123",
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", resp.Name)
	assert.Equal(t, "padaria", resp.Segment)
}
```

- [ ] **Step 2: Rodar teste para ver falhar**

Run: `cd backend && go test ./internal/service/ -run TestBusinessService -v`
Expected: FAIL — package not found

- [ ] **Step 3: Criar `internal/service/business.go`**

```go
package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type BusinessService struct {
	Queries *repository.Queries
}

func NewBusinessService(q *repository.Queries) *BusinessService {
	return &BusinessService{Queries: q}
}

func (s *BusinessService) GetByOwner(ctx context.Context, ownerID int64) (*domain.BusinessResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}
	return toBizResponse(&biz), nil
}

func (s *BusinessService) Update(ctx context.Context, ownerID int64, input domain.UpdateBusinessInput) (*domain.BusinessResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}

	updated, err := s.Queries.UpdateBusiness(ctx, repository.UpdateBusinessParams{
		ID:      biz.ID,
		Name:    input.Name,
		Segment: input.Segment,
		Address: pgText(input.Address),
		LogoUrl: pgText(input.LogoURL),
	})
	if err != nil {
		return nil, domain.ErrInternal("erro ao atualizar negócio", err)
	}
	return toBizResponse(&updated), nil
}

func toBizResponse(b *repository.Business) *domain.BusinessResponse {
	return &domain.BusinessResponse{
		ID:      b.ID,
		Name:    b.Name,
		Slug:    b.Slug,
		Segment: b.Segment,
		Address: b.Address.String,
		LogoURL: b.LogoUrl.String,
	}
}

// pgText converts a string to pgtype.Text.
func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
```

- [ ] **Step 4: Rodar testes para ver passar**

Run: `cd backend && go test ./internal/service/ -run TestBusinessService -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
cd backend && git add internal/service/business.go internal/service/business_test.go
git commit -m "feat: BusinessService com GetByOwner e Update"
```

---

### Task 7: PlanService

**Files:**
- Create: `internal/service/plan.go`
- Create: `internal/service/plan_test.go`

- [ ] **Step 1: Escrever teste do PlanService**

Criar `internal/service/plan_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
	"github.com/clubepay/backend/internal/testutil"
)

func TestPlanService_Create(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewPlanService(queries)

	owner := testutil.SeedOwner(t, queries, "plan-svc@test.com", "Owner")
	testutil.SeedBusiness(t, queries, owner.ID, "Plan Cafe", "plan-cafe")

	t.Run("success", func(t *testing.T) {
		resp, err := svc.Create(context.Background(), owner.ID, domain.CreatePlanInput{
			Name:       "Cafe Diario",
			PriceCents: 2990,
			LimitType:  "daily",
			LimitCount: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, "Cafe Diario", resp.Name)
		assert.Equal(t, int64(2990), resp.PriceCents)
	})

	t.Run("free tier limit", func(t *testing.T) {
		// Already created 1 plan above, so a second should fail
		_, err := svc.Create(context.Background(), owner.ID, domain.CreatePlanInput{
			Name:       "Segundo Plano",
			PriceCents: 1990,
			LimitType:  "monthly",
			LimitCount: 4,
		})
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.ErrorAs(t, err, &svcErr)
		assert.Equal(t, 403, svcErr.Code)
	})

	t.Run("no business", func(t *testing.T) {
		_, err := svc.Create(context.Background(), 999999, domain.CreatePlanInput{
			Name:       "X",
			PriceCents: 100,
			LimitType:  "daily",
			LimitCount: 1,
		})
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.ErrorAs(t, err, &svcErr)
		assert.Equal(t, 404, svcErr.Code)
	})
}

func TestPlanService_List(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewPlanService(queries)

	owner := testutil.SeedOwner(t, queries, "plan-list@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "List Cafe", "list-cafe")
	testutil.SeedPlan(t, queries, biz.ID, "Plano A", 1000, "daily", 1)

	resp, err := svc.List(context.Background(), owner.ID)
	require.NoError(t, err)
	assert.Len(t, resp.Plans, 1)
	assert.Equal(t, "Plano A", resp.Plans[0].Name)
}

func TestPlanService_Update(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewPlanService(queries)

	owner := testutil.SeedOwner(t, queries, "plan-upd@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Upd Cafe", "upd-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Old Plan", 1000, "daily", 1)

	resp, err := svc.Update(context.Background(), owner.ID, plan.ID, domain.UpdatePlanInput{
		Name:       "New Plan",
		PriceCents: 2000,
		LimitType:  "monthly",
		LimitCount: 4,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Plan", resp.Name)
	assert.Equal(t, int64(2000), resp.PriceCents)
}

func TestPlanService_Update_NotOwner(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewPlanService(queries)

	owner := testutil.SeedOwner(t, queries, "plan-upd-own@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Own Cafe", "own-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Plan", 1000, "daily", 1)

	other := testutil.SeedOwner(t, queries, "plan-upd-other@test.com", "Other")
	testutil.SeedBusiness(t, queries, other.ID, "Other Cafe", "other-cafe")

	_, err := svc.Update(context.Background(), other.ID, plan.ID, domain.UpdatePlanInput{
		Name:       "Hijack",
		PriceCents: 1,
		LimitType:  "daily",
		LimitCount: 1,
	})
	require.Error(t, err)
	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 403, svcErr.Code)
}

func TestPlanService_Delete(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	svc := service.NewPlanService(queries)

	owner := testutil.SeedOwner(t, queries, "plan-del@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Del Cafe", "del-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "To Delete", 1000, "daily", 1)

	err := svc.Delete(context.Background(), owner.ID, plan.ID)
	require.NoError(t, err)
}
```

- [ ] **Step 2: Rodar testes para ver falhar**

Run: `cd backend && go test ./internal/service/ -run TestPlanService -v`
Expected: FAIL

- [ ] **Step 3: Criar `internal/service/plan.go`**

```go
package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type PlanService struct {
	Queries *repository.Queries
}

func NewPlanService(q *repository.Queries) *PlanService {
	return &PlanService{Queries: q}
}

func (s *PlanService) Create(ctx context.Context, ownerID int64, input domain.CreatePlanInput) (*domain.PlanDetailResponse, error) {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	count, err := s.Queries.CountPlansByBusinessID(ctx, biz.ID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao verificar planos", err)
	}
	if count >= int64(domain.FreeTierPlanLimit) {
		return nil, domain.ErrForbidden("limite de plano atingido no plano gratuito")
	}

	plan, err := s.Queries.CreatePlan(ctx, repository.CreatePlanParams{
		BusinessID:  biz.ID,
		Name:        input.Name,
		Description: pgText(input.Description),
		PriceCents:  input.PriceCents,
		LimitType:   input.LimitType,
		LimitCount:  input.LimitCount,
	})
	if err != nil {
		return nil, domain.ErrInternal("erro ao criar plano", err)
	}
	return toPlanResponse(&plan), nil
}

func (s *PlanService) List(ctx context.Context, ownerID int64) (*domain.PlanListResponse, error) {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	plans, err := s.Queries.ListPlansByBusinessID(ctx, biz.ID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao listar planos", err)
	}

	result := make([]domain.PlanDetailResponse, len(plans))
	for i := range plans {
		result[i] = *toPlanResponse(&plans[i])
	}
	return &domain.PlanListResponse{Plans: result}, nil
}

func (s *PlanService) Update(ctx context.Context, ownerID, planID int64, input domain.UpdatePlanInput) (*domain.PlanDetailResponse, error) {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	existing, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("plano não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar plano", err)
	}
	if existing.BusinessID != biz.ID {
		return nil, domain.ErrForbidden("plano não pertence ao seu negócio")
	}

	updated, err := s.Queries.UpdatePlan(ctx, repository.UpdatePlanParams{
		ID:          planID,
		Name:        input.Name,
		Description: pgText(input.Description),
		PriceCents:  input.PriceCents,
		LimitType:   input.LimitType,
		LimitCount:  input.LimitCount,
	})
	if err != nil {
		return nil, domain.ErrInternal("erro ao atualizar plano", err)
	}
	return toPlanResponse(&updated), nil
}

func (s *PlanService) Delete(ctx context.Context, ownerID, planID int64) error {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return err
	}

	existing, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound("plano não encontrado")
		}
		return domain.ErrInternal("erro ao buscar plano", err)
	}
	if existing.BusinessID != biz.ID {
		return domain.ErrForbidden("plano não pertence ao seu negócio")
	}

	if err := s.Queries.DeactivatePlan(ctx, planID); err != nil {
		return domain.ErrInternal("erro ao desativar plano", err)
	}
	return nil
}

func (s *PlanService) getOwnerBusiness(ctx context.Context, ownerID int64) (*repository.Business, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}
	return &biz, nil
}

func toPlanResponse(p *repository.Plan) *domain.PlanDetailResponse {
	return &domain.PlanDetailResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description.String,
		PriceCents:  p.PriceCents,
		LimitType:   p.LimitType,
		LimitCount:  p.LimitCount,
		Active:      p.Active,
	}
}
```

- [ ] **Step 4: Rodar testes para ver passar**

Run: `cd backend && go test ./internal/service/ -run TestPlanService -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
cd backend && git add internal/service/plan.go internal/service/plan_test.go
git commit -m "feat: PlanService com Create, List, Update, Delete"
```

---

### Task 8: SubscriptionService

**Files:**
- Create: `internal/service/subscription.go`
- Create: `internal/service/subscription_test.go`

- [ ] **Step 1: Escrever teste do SubscriptionService**

Criar `internal/service/subscription_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
	"github.com/clubepay/backend/internal/testutil"
)

func setupSubscriptionService(t *testing.T) (*service.SubscriptionService, *repository.Queries) {
	pool := testutil.SetupTestDB(t)
	queries := repository.New(pool)
	mockPSP := &psp.MockPSP{}
	mockEmail := &email.MockSender{}
	svc := service.NewSubscriptionService(queries, mockPSP, mockEmail)
	return svc, queries
}

func TestSubscriptionService_Subscribe(t *testing.T) {
	svc, queries := setupSubscriptionService(t)

	owner := testutil.SeedOwner(t, queries, "sub-svc@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Sub Cafe", "sub-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Diario", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, queries, "sub1@test.com", "Sub1", "11999999999")

	t.Run("success", func(t *testing.T) {
		sub, err := svc.Subscribe(context.Background(), subscriber.ID, domain.SubscribeInput{PlanID: plan.ID})
		require.NoError(t, err)
		assert.Equal(t, "active", sub.Status)
	})

	t.Run("already subscribed", func(t *testing.T) {
		_, err := svc.Subscribe(context.Background(), subscriber.ID, domain.SubscribeInput{PlanID: plan.ID})
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.ErrorAs(t, err, &svcErr)
		assert.Equal(t, 409, svcErr.Code)
	})
}

func TestSubscriptionService_CancelByOwner(t *testing.T) {
	svc, queries := setupSubscriptionService(t)

	owner := testutil.SeedOwner(t, queries, "cancel-own@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "Cancel Cafe", "cancel-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Plano", 1000, "daily", 1)
	sub1 := testutil.SeedSubscriber(t, queries, "cancel-sub@test.com", "Sub", "")
	subscription := testutil.SeedSubscription(t, queries, plan.ID, sub1.ID, biz.ID)

	err := svc.CancelByOwner(context.Background(), owner.ID, subscription.ID)
	require.NoError(t, err)
}

func TestSubscriptionService_CancelBySubscriber(t *testing.T) {
	svc, queries := setupSubscriptionService(t)

	owner := testutil.SeedOwner(t, queries, "cancel-sub-own@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "CancelSub Cafe", "cancelsub-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Plano", 1000, "daily", 1)
	sub1 := testutil.SeedSubscriber(t, queries, "cancel-sub2@test.com", "Sub", "")
	testutil.SeedSubscription(t, queries, plan.ID, sub1.ID, biz.ID)

	resp, err := svc.CancelBySubscriber(context.Background(), sub1.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", resp.Status)
}

func TestSubscriptionService_GetMyPlan(t *testing.T) {
	svc, queries := setupSubscriptionService(t)

	owner := testutil.SeedOwner(t, queries, "myplan-own@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "MyPlan Cafe", "myplan-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Plano A", 2990, "daily", 1)
	sub1 := testutil.SeedSubscriber(t, queries, "myplan-sub@test.com", "Sub", "")
	testutil.SeedSubscription(t, queries, plan.ID, sub1.ID, biz.ID)

	resp, err := svc.GetMyPlan(context.Background(), sub1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Plano A", resp.Plan.Name)
	assert.Equal(t, "MyPlan Cafe", resp.Business.Name)
}

func TestSubscriptionService_ListByBusiness(t *testing.T) {
	svc, queries := setupSubscriptionService(t)

	owner := testutil.SeedOwner(t, queries, "list-sub@test.com", "Owner")
	biz := testutil.SeedBusiness(t, queries, owner.ID, "List Cafe", "list-sub-cafe")
	plan := testutil.SeedPlan(t, queries, biz.ID, "Plano", 1000, "daily", 1)
	sub1 := testutil.SeedSubscriber(t, queries, "list-sub1@test.com", "Sub1", "")
	testutil.SeedSubscription(t, queries, plan.ID, sub1.ID, biz.ID)

	resp, err := svc.ListByBusiness(context.Background(), owner.ID)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}
```

- [ ] **Step 2: Rodar teste para ver falhar**

Run: `cd backend && go test ./internal/service/ -run TestSubscriptionService -v`
Expected: FAIL

- [ ] **Step 3: Criar `internal/service/subscription.go`**

```go
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
)

type SubscriptionService struct {
	Queries *repository.Queries
	PSP     psp.PSP
	Email   email.Sender
}

func NewSubscriptionService(q *repository.Queries, p psp.PSP, e email.Sender) *SubscriptionService {
	return &SubscriptionService{Queries: q, PSP: p, Email: e}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, subscriberID int64, input domain.SubscribeInput) (*repository.Subscription, error) {
	plan, err := s.Queries.GetPlanByID(ctx, input.PlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("plano não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar plano", err)
	}
	if !plan.Active {
		return nil, domain.ErrBadRequest("plano não está ativo")
	}

	_, err = s.Queries.GetActiveSubscription(ctx, repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   plan.BusinessID,
	})
	if err == nil {
		return nil, domain.ErrConflict("já possui assinatura ativa neste negócio")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInternal("erro ao verificar assinatura existente", err)
	}

	count, err := s.Queries.CountActiveSubscriptionsByBusinessID(ctx, plan.BusinessID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao contar assinantes", err)
	}
	if count >= int64(domain.FreeTierSubscriberLimit) {
		return nil, domain.ErrForbidden("limite de assinantes atingido no plano gratuito")
	}

	subscriber, err := s.Queries.GetUserByID(ctx, subscriberID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao buscar dados do assinante", err)
	}

	customer, err := s.PSP.CreateCustomer(ctx, psp.CreateCustomerRequest{
		Name:  subscriber.Name,
		Email: subscriber.Email,
		Phone: subscriber.Phone.String,
	})
	if err != nil {
		slog.Error("failed to create PSP customer", "error", err)
		return nil, domain.ErrInternal("erro ao criar cliente no PSP", err)
	}

	discountPercent := int32(0)
	referral, refErr := s.Queries.GetReferralByReferredAndBusiness(ctx, repository.GetReferralByReferredAndBusinessParams{
		ReferredID: subscriberID,
		BusinessID: plan.BusinessID,
	})
	if refErr == nil {
		discountPercent = int32(domain.ReferralDiscountPercent)
	}

	priceCents := plan.PriceCents
	if discountPercent > 0 {
		priceCents = priceCents * int64(100-discountPercent) / 100
	}

	pspSub, err := s.PSP.CreateSubscription(ctx, psp.CreateSubscriptionRequest{
		CustomerID:  customer.ID,
		PriceCents:  priceCents,
		Description: plan.Name,
		Cycle:       "MONTHLY",
	})
	if err != nil {
		slog.Error("failed to create PSP subscription", "error", err)
		return nil, domain.ErrInternal("erro ao criar assinatura no PSP", err)
	}

	periodEnd := time.Now().AddDate(0, 1, 0)
	referredBy := pgtype.Int8{}
	if refErr == nil {
		referredBy = pgtype.Int8{Int64: referral.ReferrerID, Valid: true}
	}

	sub, err := s.Queries.CreateSubscriptionWithDiscount(ctx, repository.CreateSubscriptionWithDiscountParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriberID,
		BusinessID:        plan.BusinessID,
		PspSubscriptionID: pgText(pspSub.ID),
		Status:            domain.SubscriptionStatusActive,
		PeriodEnd:         pgtype.Timestamptz{Time: periodEnd, Valid: true},
		ReferredBy:        referredBy,
		DiscountPercent:   discountPercent,
	})
	if err != nil {
		return nil, domain.ErrInternal("erro ao criar assinatura", err)
	}

	go s.sendWelcomeEmail(subscriberID, plan.ID, plan.BusinessID)

	return &sub, nil
}

func (s *SubscriptionService) ListByBusiness(ctx context.Context, ownerID int64) (*domain.SubscriptionListResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}

	subs, err := s.Queries.ListSubscriptionsByBusinessID(ctx, biz.ID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao listar assinaturas", err)
	}

	return &domain.SubscriptionListResponse{Subscriptions: subs}, nil
}

func (s *SubscriptionService) CancelByOwner(ctx context.Context, ownerID, subID int64) error {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound("negócio não encontrado")
		}
		return domain.ErrInternal("erro ao buscar negócio", err)
	}

	sub, err := s.Queries.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound("assinatura não encontrada")
		}
		return domain.ErrInternal("erro ao buscar assinatura", err)
	}

	if sub.BusinessID != biz.ID {
		return domain.ErrForbidden("assinatura não pertence ao seu negócio")
	}

	s.cancelPSPSubscription(ctx, sub.PspSubscriptionID)

	if err := s.Queries.CancelSubscription(ctx, subID); err != nil {
		return domain.ErrInternal("erro ao cancelar assinatura", err)
	}
	return nil
}

func (s *SubscriptionService) CancelBySubscriber(ctx context.Context, subscriberID int64) (*domain.CancelResponse, error) {
	sub, err := s.Queries.GetActiveSubscriptionBySubscriber(ctx, subscriberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("assinatura não encontrada")
		}
		return nil, domain.ErrInternal("erro ao buscar assinatura", err)
	}

	s.cancelPSPSubscription(ctx, sub.PspSubscriptionID)

	if err := s.Queries.CancelSubscription(ctx, sub.ID); err != nil {
		return nil, domain.ErrInternal("erro ao cancelar assinatura", err)
	}

	go s.sendCancellationEmail(subscriberID, sub.PlanID, sub.PeriodEnd)

	var periodEnd *time.Time
	if sub.PeriodEnd.Valid {
		periodEnd = &sub.PeriodEnd.Time
	}

	return &domain.CancelResponse{
		Status:    "cancelled",
		PeriodEnd: periodEnd,
		Message:   "Assinatura cancelada. Acesso continua até o fim do período pago.",
	}, nil
}

func (s *SubscriptionService) GetMyPlan(ctx context.Context, subscriberID int64) (*domain.MyPlanResponse, error) {
	sub, err := s.Queries.GetActiveSubscriptionBySubscriber(ctx, subscriberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("assinatura não encontrada")
		}
		return nil, domain.ErrInternal("erro ao buscar assinatura", err)
	}

	plan, err := s.Queries.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao buscar plano", err)
	}

	biz, err := s.Queries.GetBusinessByID(ctx, sub.BusinessID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}

	var periodEnd *time.Time
	if sub.PeriodEnd.Valid {
		periodEnd = &sub.PeriodEnd.Time
	}

	return &domain.MyPlanResponse{
		Plan:     *toPlanResponse(&plan),
		Business: *toBizResponse(&biz),
		Subscription: domain.SubscriptionInfo{
			ID:        sub.ID,
			Status:    sub.Status,
			PeriodEnd: periodEnd,
		},
	}, nil
}

// cancelPSPSubscription cancels a subscription in PSP (best-effort, logs errors).
func (s *SubscriptionService) cancelPSPSubscription(ctx context.Context, pspSubID pgtype.Text) {
	if pspSubID.Valid && pspSubID.String != "" {
		if err := s.PSP.CancelSubscription(ctx, pspSubID.String); err != nil {
			slog.Error("failed to cancel PSP subscription", "error", err, "psp_id", pspSubID.String)
		}
	}
}

func (s *SubscriptionService) sendWelcomeEmail(subscriberID, planID, businessID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subscriber, err := s.Queries.GetUserByID(ctx, subscriberID)
	if err != nil {
		slog.Error("failed to fetch subscriber for welcome email", "error", err, "subscriber_id", subscriberID)
		return
	}
	plan, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		slog.Error("failed to fetch plan for welcome email", "error", err, "plan_id", planID)
		return
	}
	biz, err := s.Queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		slog.Error("failed to fetch business for welcome email", "error", err, "business_id", businessID)
		return
	}
	subject, body := email.WelcomeEmail(subscriber.Name, plan.Name, biz.Name)
	if err := s.Email.Send(subscriber.Email, subject, body); err != nil {
		slog.Error("failed to send welcome email", "error", err, "to", subscriber.Email)
	}
}

func (s *SubscriptionService) sendCancellationEmail(subscriberID, planID int64, periodEnd pgtype.Timestamptz) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := s.Queries.GetUserByID(ctx, subscriberID)
	if err != nil {
		slog.Error("failed to fetch user for cancellation email", "error", err, "subscriber_id", subscriberID)
		return
	}
	plan, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		slog.Error("failed to fetch plan for cancellation email", "error", err, "plan_id", planID)
		return
	}
	validUntil := ""
	if periodEnd.Valid {
		validUntil = periodEnd.Time.Format("02/01/2006")
	}
	subject, body := email.SubscriptionCancelledEmail(user.Name, plan.Name, validUntil)
	if err := s.Email.Send(user.Email, subject, body); err != nil {
		slog.Error("failed to send cancellation email", "error", err, "to", user.Email)
	}
}
```

Nota: Este service precisa que `GetSubscriptionByID` exista no repository. Verificar se ja existe antes de implementar; se nao existir, adicionar a query SQL.

- [ ] **Step 4: Rodar testes para ver passar**

Run: `cd backend && go test ./internal/service/ -run TestSubscriptionService -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
cd backend && git add internal/service/subscription.go internal/service/subscription_test.go
git commit -m "feat: SubscriptionService com Subscribe, Cancel, GetMyPlan, List"
```

---

### Task 9: UsageService e ReferralService

**Files:**
- Create: `internal/service/usage.go`
- Create: `internal/service/referral.go`
- Create: `internal/service/usage_test.go`
- Create: `internal/service/referral_test.go`

- [ ] **Step 1: Criar `internal/service/usage.go`**

```go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type UsageService struct {
	Queries *repository.Queries
}

func NewUsageService(q *repository.Queries) *UsageService {
	return &UsageService{Queries: q}
}

func (s *UsageService) Validate(ctx context.Context, subscriberID int64, input domain.ValidateUsageInput) (*domain.ValidateUsageResponse, error) {
	biz, err := s.Queries.GetBusinessBySlug(ctx, input.BusinessSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}
	return s.validateUsage(ctx, subscriberID, biz.ID)
}

func (s *UsageService) ValidateByOwner(ctx context.Context, ownerID int64, input domain.ValidateUsageOwnerInput) (*domain.ValidateUsageResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}
	return s.validateUsage(ctx, input.SubscriberID, biz.ID)
}

func (s *UsageService) GetMyUsage(ctx context.Context, subscriberID int64, businessSlug string) (*domain.UsageListResponse, error) {
	biz, err := s.Queries.GetBusinessBySlug(ctx, businessSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("negócio não encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}

	sub, err := s.Queries.GetActiveSubscription(ctx, repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   biz.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("assinatura não encontrada")
		}
		return nil, domain.ErrInternal("erro ao buscar assinatura", err)
	}

	plan, err := s.Queries.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao buscar plano", err)
	}

	now := time.Now()
	var start, end time.Time
	if plan.LimitType == domain.LimitTypeDaily {
		start, end = domain.DailyRange(now)
	} else {
		start, end = domain.MonthlyRange(now)
	}

	usages, err := s.Queries.ListUsagesBySubscription(ctx, repository.ListUsagesBySubscriptionParams{
		SubscriptionID: sub.ID,
		Column2:        pgtype.Timestamptz{Time: start, Valid: true},
		Column3:        pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, domain.ErrInternal("erro ao listar usos", err)
	}

	return &domain.UsageListResponse{
		Used:     len(usages),
		Limit:    plan.LimitCount,
		PlanName: plan.Name,
		Usages:   usages,
	}, nil
}

func (s *UsageService) validateUsage(ctx context.Context, subscriberID, businessID int64) (*domain.ValidateUsageResponse, error) {
	sub, err := s.Queries.GetActiveSubscription(ctx, repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   businessID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("assinatura não encontrada")
		}
		return nil, domain.ErrInternal("erro ao buscar assinatura", err)
	}

	plan, err := s.Queries.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao buscar plano", err)
	}

	now := time.Now()
	var currentCount int64

	if plan.LimitType == domain.LimitTypeDaily {
		start, end := domain.DailyRange(now)
		currentCount, err = s.Queries.CountDailyUsage(ctx, repository.CountDailyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgtype.Timestamptz{Time: start, Valid: true},
			Column3:        pgtype.Timestamptz{Time: end, Valid: true},
		})
	} else {
		start, end := domain.MonthlyRange(now)
		currentCount, err = s.Queries.CountMonthlyUsage(ctx, repository.CountMonthlyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgtype.Timestamptz{Time: start, Valid: true},
			Column3:        pgtype.Timestamptz{Time: end, Valid: true},
		})
	}
	if err != nil {
		return nil, domain.ErrInternal("erro ao contar usos", err)
	}

	if err := domain.CheckUsageLimit(plan.LimitType, int(plan.LimitCount), currentCount); err != nil {
		return nil, domain.ErrForbidden("limite de uso atingido para este período")
	}

	if _, err := s.Queries.CreateUsage(ctx, sub.ID); err != nil {
		return nil, domain.ErrInternal("erro ao registrar uso", err)
	}

	return &domain.ValidateUsageResponse{
		Status:   "validated",
		Used:     currentCount + 1,
		Limit:    plan.LimitCount,
		PlanName: plan.Name,
	}, nil
}
```

- [ ] **Step 2: Criar `internal/service/referral.go`**

```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type ReferralService struct {
	Queries *repository.Queries
}

func NewReferralService(q *repository.Queries) *ReferralService {
	return &ReferralService{Queries: q}
}

func (s *ReferralService) GetOrCreateCode(ctx context.Context, subscriberID int64) (*domain.ReferralCodeResponse, error) {
	sub, err := s.Queries.GetActiveSubscriptionBySubscriber(ctx, subscriberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound("assinatura não encontrada")
		}
		return nil, domain.ErrInternal("erro ao buscar assinatura", err)
	}

	biz, err := s.Queries.GetBusinessByID(ctx, sub.BusinessID)
	if err != nil {
		return nil, domain.ErrInternal("erro ao buscar negócio", err)
	}

	code, err := s.Queries.GetReferralCodeBySubscriberAndBusiness(ctx, repository.GetReferralCodeBySubscriberAndBusinessParams{
		ReferrerID: subscriberID,
		BusinessID: biz.ID,
	})
	if err == nil {
		return &domain.ReferralCodeResponse{Code: code}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInternal("erro ao buscar código de indicação", err)
	}

	codeBytes := make([]byte, 4)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, domain.ErrInternal("erro ao gerar código", err)
	}
	newCode := hex.EncodeToString(codeBytes)

	if _, err := s.Queries.CreateReferral(ctx, repository.CreateReferralParams{
		ReferrerID: subscriberID,
		ReferredID: subscriberID,
		BusinessID: biz.ID,
		Code:       newCode,
	}); err != nil {
		return nil, domain.ErrInternal("erro ao criar código de indicação", err)
	}

	return &domain.ReferralCodeResponse{Code: newCode}, nil
}

func (s *ReferralService) Apply(ctx context.Context, subscriberID int64, input domain.ApplyReferralInput) error {
	referral, err := s.Queries.GetReferralByCode(ctx, input.Code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound("código de indicação não encontrado")
		}
		return domain.ErrInternal("erro ao buscar indicação", err)
	}

	if referral.ReferrerID == subscriberID {
		return domain.ErrBadRequest("não é possível usar seu próprio código de indicação")
	}

	count, err := s.Queries.CountActiveReferralsByReferrer(ctx, repository.CountActiveReferralsByReferrerParams{
		ReferrerID: referral.ReferrerID,
		BusinessID: referral.BusinessID,
	})
	if err != nil {
		return domain.ErrInternal("erro ao contar indicações", err)
	}
	if count > int64(domain.ReferralLimit) {
		return domain.ErrForbidden("limite de indicações atingido (máximo 3)")
	}

	if _, err := s.Queries.CreateReferral(ctx, repository.CreateReferralParams{
		ReferrerID: referral.ReferrerID,
		ReferredID: subscriberID,
		BusinessID: referral.BusinessID,
		Code:       input.Code,
	}); err != nil {
		return domain.ErrInternal("erro ao aplicar indicação", err)
	}

	return nil
}
```

- [ ] **Step 3: Escrever testes basicos e rodar**

Criar `internal/service/usage_test.go` e `internal/service/referral_test.go` com testes para os happy paths e error paths principais. Seguir o mesmo padrao dos testes de BusinessService e PlanService.

- [ ] **Step 4: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
cd backend && git add internal/service/usage.go internal/service/referral.go internal/service/usage_test.go internal/service/referral_test.go
git commit -m "feat: UsageService e ReferralService"
```

---

### Task 10: AuthService

**Files:**
- Create: `internal/service/auth.go`
- Create: `internal/service/auth_test.go`

- [ ] **Step 1: Criar `internal/service/auth.go`**

Migrar logica de RegisterOwner, RegisterSubscriber, Login, RequestPasswordReset, ConfirmPasswordReset dos handlers. Incluir:
- `generateSlug()` e `normalizeRune()` (mover de handler/auth.go)
- `isDuplicateKeyError()` (mover de handler/auth.go)
- `sendPasswordResetEmail()` com `context.WithTimeout` e log de erros

```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/repository"
)

type AuthService struct {
	Queries *repository.Queries
	Config  *config.Config
	Email   email.Sender
}

func NewAuthService(q *repository.Queries, cfg *config.Config, e email.Sender) *AuthService {
	return &AuthService{Queries: q, Config: cfg, Email: e}
}

func (s *AuthService) RegisterOwner(ctx context.Context, input domain.RegisterOwnerInput) (*domain.AuthResponse, error) {
	if input.Segment == "" {
		input.Segment = "cafeteria"
	}

	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		return nil, domain.ErrInternal("failed to hash password", err)
	}

	user, err := s.Queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        input.Email,
		PasswordHash: hash,
		Name:         input.Name,
		Phone:        pgText(input.Phone),
		Role:         domain.RoleOwner,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.ErrConflict("email already in use")
		}
		return nil, domain.ErrInternal("failed to create user", err)
	}

	slug := generateSlug(input.BusinessName)
	biz, err := s.Queries.CreateBusiness(ctx, repository.CreateBusinessParams{
		OwnerID: user.ID,
		Name:    input.BusinessName,
		Slug:    slug,
		Segment: input.Segment,
	})
	if err != nil {
		return nil, domain.ErrInternal("failed to create business", err)
	}

	token, err := domain.GenerateJWT(user.ID, domain.RoleOwner, s.Config.JWTSecret, domain.OwnerJWTExpiry)
	if err != nil {
		return nil, domain.ErrInternal("failed to generate token", err)
	}

	bizResp := &domain.BusinessResponse{
		ID: biz.ID, Name: biz.Name, Slug: biz.Slug, Segment: biz.Segment,
	}

	return &domain.AuthResponse{
		Token:    token,
		User:     toUserResponse(&user),
		Business: bizResp,
	}, nil
}

func (s *AuthService) RegisterSubscriber(ctx context.Context, input domain.RegisterSubscriberInput) (*domain.AuthResponse, error) {
	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		return nil, domain.ErrInternal("failed to hash password", err)
	}

	user, err := s.Queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        input.Email,
		PasswordHash: hash,
		Name:         input.Name,
		Phone:        pgText(input.Phone),
		Role:         domain.RoleSubscriber,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.ErrConflict("email already in use")
		}
		return nil, domain.ErrInternal("failed to create user", err)
	}

	token, err := domain.GenerateJWT(user.ID, domain.RoleSubscriber, s.Config.JWTSecret, domain.SubscriberJWTExpiry)
	if err != nil {
		return nil, domain.ErrInternal("failed to generate token", err)
	}

	return &domain.AuthResponse{
		Token: token,
		User:  toUserResponse(&user),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	user, err := s.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBadRequest("invalid credentials")
		}
		return nil, domain.ErrInternal("failed to query user", err)
	}

	if !domain.CheckPassword(input.Password, user.PasswordHash) {
		return nil, domain.ErrBadRequest("invalid credentials")
	}

	expiry := domain.OwnerJWTExpiry
	if user.Role == domain.RoleSubscriber {
		expiry = domain.SubscriberJWTExpiry
	}

	token, err := domain.GenerateJWT(user.ID, user.Role, s.Config.JWTSecret, expiry)
	if err != nil {
		return nil, domain.ErrInternal("failed to generate token", err)
	}

	return &domain.AuthResponse{
		Token: token,
		User:  toUserResponse(&user),
	}, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, emailAddr string) {
	user, err := s.Queries.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	if _, err := s.Queries.CreatePasswordReset(ctx, repository.CreatePasswordResetParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}); err != nil {
		return
	}

	go s.sendPasswordResetEmail(user.Name, user.Email, token)
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, input domain.ConfirmPasswordResetInput) error {
	reset, err := s.Queries.GetPasswordResetByToken(ctx, input.Token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrBadRequest("invalid or expired token")
		}
		return domain.ErrInternal("failed to validate token", err)
	}

	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		return domain.ErrInternal("failed to hash password", err)
	}

	if err := s.Queries.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:           reset.UserID,
		PasswordHash: hash,
	}); err != nil {
		return domain.ErrInternal("failed to update password", err)
	}

	if err := s.Queries.MarkPasswordResetUsed(ctx, reset.ID); err != nil {
		return domain.ErrInternal("failed to mark token used", err)
	}

	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*domain.ProfileResponse, error) {
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound("usuario nao encontrado")
		}
		return nil, domain.ErrInternal("erro ao buscar perfil", err)
	}
	return &domain.ProfileResponse{User: toUserResponse(&user)}, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, input domain.UpdateProfileInput) (*domain.ProfileResponse, error) {
	user, err := s.Queries.UpdateUserProfile(ctx, repository.UpdateUserProfileParams{
		ID:    userID,
		Name:  input.Name,
		Phone: pgText(input.Phone),
	})
	if err != nil {
		return nil, domain.ErrInternal("erro ao atualizar perfil", err)
	}
	return &domain.ProfileResponse{User: toUserResponse(&user)}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, input domain.ChangePasswordInput) error {
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return domain.ErrInternal("erro ao buscar usuario", err)
	}

	if !domain.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		return domain.ErrBadRequest("senha atual incorreta")
	}

	hash, err := domain.HashPassword(input.NewPassword)
	if err != nil {
		return domain.ErrInternal("erro ao processar senha", err)
	}

	if err := s.Queries.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hash,
	}); err != nil {
		return domain.ErrInternal("erro ao atualizar senha", err)
	}

	return nil
}

func (s *AuthService) sendPasswordResetEmail(userName, userEmail, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = ctx // used to satisfy linter; email.Send doesn't take context

	resetURL := s.Config.FrontendURL + "/resetar-senha?token=" + token
	subject, body := email.PasswordResetEmail(userName, resetURL)
	if s.Email != nil {
		if err := s.Email.Send(userEmail, subject, body); err != nil {
			slog.Error("failed to send password reset email", "error", err, "to", userEmail)
		}
	}
}

func toUserResponse(u *repository.User) domain.UserResponse {
	return domain.UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Phone: u.Phone.String,
		Role:  u.Role,
	}
}

func generateSlug(name string) string {
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return unicode.ToLower(r)
		case r >= '0' && r <= '9':
			return r
		case r == ' ' || r == '-' || r == '_':
			return '-'
		default:
			return normalizeRune(r)
		}
	}, name)

	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func normalizeRune(r rune) rune {
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
		'Á': 'a', 'À': 'a', 'Ã': 'a', 'Â': 'a', 'Ä': 'a',
		'É': 'e', 'È': 'e', 'Ê': 'e', 'Ë': 'e',
		'Í': 'i', 'Ì': 'i', 'Î': 'i', 'Ï': 'i',
		'Ó': 'o', 'Ò': 'o', 'Õ': 'o', 'Ô': 'o', 'Ö': 'o',
		'Ú': 'u', 'Ù': 'u', 'Û': 'u', 'Ü': 'u',
		'Ç': 'c', 'Ñ': 'n',
	}
	if v, ok := replacements[r]; ok {
		return v
	}
	return -1
}

func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
```

- [ ] **Step 2: Escrever testes para AuthService**

Criar `internal/service/auth_test.go` com testes para RegisterOwner, Login, RegisterSubscriber, ConfirmPasswordReset, GetProfile, UpdateProfile, ChangePassword.

- [ ] **Step 3: Rodar testes**

Run: `cd backend && go test ./internal/service/ -v`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
cd backend && git add internal/service/auth.go internal/service/auth_test.go
git commit -m "feat: AuthService com Register, Login, Profile, PasswordReset"
```

---

## Fase 3: Refatorar Handlers para usar Services

---

### Task 11: Atualizar Handler struct e main.go

**Files:**
- Modify: `internal/handler/handler.go`
- Modify: `cmd/api/main.go`

- [ ] **Step 1: Atualizar `internal/handler/handler.go`**

Substituir a Handler struct e o construtor New:

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
)

type Handler struct {
	Auth          *service.AuthService
	Business      *service.BusinessService
	Plans         *service.PlanService
	Subscriptions *service.SubscriptionService
	Usage         *service.UsageService
	Referrals     *service.ReferralService
	Config        *config.Config
	// Keep these for webhook and cron handlers that need direct access
	Queries *repository.Queries
	PSP     psp.PSP
}

func New(q *repository.Queries, cfg *config.Config, p psp.PSP, e email.Sender) *Handler {
	return &Handler{
		Auth:          service.NewAuthService(q, cfg, e),
		Business:      service.NewBusinessService(q),
		Plans:         service.NewPlanService(q),
		Subscriptions: service.NewSubscriptionService(q, p, e),
		Usage:         service.NewUsageService(q),
		Referrals:     service.NewReferralService(q),
		Config:        cfg,
		Queries:       q,
		PSP:           p,
	}
}
```

Manter writeJSON, writeError, readJSON, pgText, pgTimestamptz e handleServiceError.

Adicionar import de `email` e `service`.

- [ ] **Step 2: Verificar que `main.go` nao precisa mudar**

O construtor `handler.New` mantem a mesma assinatura, entao `cmd/api/main.go` nao muda.

- [ ] **Step 3: Verificar compilacao**

Run: `cd backend && go build ./...`
Expected: BUILD OK

- [ ] **Step 4: Commit**

```bash
cd backend && git add internal/handler/handler.go
git commit -m "refactor: Handler struct com services injetados"
```

---

### Task 12: Refatorar handlers para usar services

**Files:**
- Modify: `internal/handler/auth.go`
- Modify: `internal/handler/business.go`
- Modify: `internal/handler/plan.go`
- Modify: `internal/handler/subscription.go`
- Modify: `internal/handler/subscriber.go`
- Modify: `internal/handler/usage.go`
- Modify: `internal/handler/referral.go`
- Modify: `internal/handler/profile.go`
- Modify: `internal/handler/public.go`
- Modify: `internal/handler/search.go`
- Modify: `internal/handler/owner_validate.go`
- Modify: `internal/handler/cron.go`
- Modify: `internal/handler/webhook.go`

Cada handler fica fino: parse request -> validate -> call service -> handleServiceError ou writeJSON.

- [ ] **Step 1: Refatorar `handler/auth.go`**

Manter generateSlug/normalizeRune/isDuplicateKeyError no service (ja movidos na Task 10). Handlers ficam:

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
	resp, err := h.Auth.RegisterOwner(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}
```

Aplicar mesmo padrao para Login, RegisterSubscriber, RequestPasswordReset, ConfirmPasswordReset.

- [ ] **Step 2: Refatorar `handler/business.go`**

```go
func (h *Handler) GetBusiness(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())
	resp, err := h.Business.GetByOwner(r.Context(), ownerID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"business": resp})
}
```

- [ ] **Step 3: Refatorar `handler/plan.go`**

CreatePlan, ListPlans, UpdatePlan, DeletePlan — todos delegam ao PlanService. Remove validacao manual.

- [ ] **Step 4: Refatorar `handler/subscription.go`**

Subscribe, ListSubscriptions, CancelSubscriptionByOwner — delegam ao SubscriptionService.

- [ ] **Step 5: Refatorar `handler/subscriber.go`**

MyPlan e CancelBySubscriber — delegam ao SubscriptionService.

- [ ] **Step 6: Refatorar `handler/usage.go`**

ValidateUsage e MyUsage — delegam ao UsageService.

- [ ] **Step 7: Refatorar `handler/referral.go`**

MyReferralCode e ApplyReferral — delegam ao ReferralService.

- [ ] **Step 8: Refatorar `handler/profile.go`**

GetProfile, UpdateProfile, ChangePassword — delegam ao AuthService. Substituir structs anonimas por domain types.

- [ ] **Step 9: Refatorar `handler/public.go`**

GetPublicBusiness e GetPublicPlans — podem usar queries diretamente (simples) ou criar metodos no BusinessService/PlanService.

- [ ] **Step 10: Refatorar `handler/search.go`**

Usar BusinessService.GetByOwner para o lookup e queries diretamente para o search.

- [ ] **Step 11: Refatorar `handler/owner_validate.go`**

Usar UsageService.ValidateByOwner.

- [ ] **Step 12: Refatorar `handler/cron.go` e `handler/webhook.go`**

Cron e webhook continuam com acesso direto a Queries e PSP (conforme spec). Atualizar goroutines de email para usar context.WithTimeout e logar erros.

- [ ] **Step 13: Rodar todos os testes**

Run: `cd backend && go test ./...`
Expected: ALL PASS. Se algum handler test falhar, ajustar o setup (setupHandler ja cria services via handler.New).

- [ ] **Step 14: Commit**

```bash
cd backend && git add internal/handler/
git commit -m "refactor: handlers delegam logica para services"
```

---

## Fase 4: Testes

---

### Task 13: Testes para config.go

**Files:**
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Escrever testes**

```go
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/config"
)

func TestLoad_RequiredFields(t *testing.T) {
	t.Run("missing DATABASE_URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("JWT_SECRET", "test-secret")
		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL")
	})

	t.Run("missing JWT_SECRET", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "")
		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET")
	})
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret")
	// Clear optional vars
	t.Setenv("PORT", "")
	t.Setenv("ASAAS_URL", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("FRONTEND_URL", "")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "https://sandbox.asaas.com/api/v3", cfg.AsaasURL)
	assert.Equal(t, "587", cfg.SMTPPort)
	assert.Equal(t, "*", cfg.CORSOrigins)
	assert.Equal(t, "http://localhost:3000", cfg.FrontendURL)
}

func TestLoad_AllVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://prod/db")
	t.Setenv("JWT_SECRET", "prod-secret")
	t.Setenv("PORT", "9090")
	t.Setenv("ASAAS_API_KEY", "key123")
	t.Setenv("ASAAS_URL", "https://api.asaas.com/api/v3")
	t.Setenv("CRON_SECRET", "cron123")
	t.Setenv("ASAAS_WEBHOOK_SECRET", "whsec123")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "user@example.com")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("CORS_ORIGINS", "https://example.com")
	t.Setenv("FRONTEND_URL", "https://app.example.com")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "key123", cfg.AsaasAPIKey)
	assert.Equal(t, "https://api.asaas.com/api/v3", cfg.AsaasURL)
	assert.Equal(t, "smtp.example.com", cfg.SMTPHost)
}
```

- [ ] **Step 2: Rodar testes**

Run: `cd backend && go test ./internal/config/ -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
cd backend && git add internal/config/config_test.go
git commit -m "test: testes para config.Load com required fields e defaults"
```

---

### Task 14: Testes para profile handlers

**Files:**
- Create: `internal/handler/profile_test.go`

- [ ] **Step 1: Escrever testes**

```go
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

func TestGetProfile_Success(t *testing.T) {
	h := setupHandler(t)

	regBody := map[string]string{
		"email": "profile@test.com", "password": "password123",
		"name": "Profile User", "business_name": "Profile Cafe",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userID := int64(regResp["user"].(map[string]interface{})["id"].(float64))

	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()
	h.GetProfile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	user := resp["user"].(map[string]interface{})
	assert.Equal(t, "Profile User", user["name"])
}

func TestUpdateProfile_Success(t *testing.T) {
	h := setupHandler(t)

	regBody := map[string]string{
		"email": "profile-upd@test.com", "password": "password123",
		"name": "Old Name", "business_name": "Cafe",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userID := int64(regResp["user"].(map[string]interface{})["id"].(float64))

	updBody := map[string]string{"name": "New Name", "phone": "11999999999"}
	ub, _ := json.Marshal(updBody)
	req := httptest.NewRequest(http.MethodPut, "/api/profile", bytes.NewReader(ub))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	user := resp["user"].(map[string]interface{})
	assert.Equal(t, "New Name", user["name"])
}

func TestUpdateProfile_InvalidJSON(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/api/profile", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, 1, "owner")
	rr := httptest.NewRecorder()
	h.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestChangePassword_Success(t *testing.T) {
	h := setupHandler(t)

	regBody := map[string]string{
		"email": "changepw@test.com", "password": "password123",
		"name": "PW User", "business_name": "PW Cafe",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userID := int64(regResp["user"].(map[string]interface{})["id"].(float64))

	pwBody := map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword123",
	}
	pb, _ := json.Marshal(pwBody)
	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader(pb))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	h := setupHandler(t)

	regBody := map[string]string{
		"email": "changepw-wrong@test.com", "password": "password123",
		"name": "PW User", "business_name": "PW Cafe2",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userID := int64(regResp["user"].(map[string]interface{})["id"].(float64))

	pwBody := map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword123",
	}
	pb, _ := json.Marshal(pwBody)
	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader(pb))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestChangePassword_ShortPassword(t *testing.T) {
	h := setupHandler(t)

	regBody := map[string]string{
		"email": "changepw-short@test.com", "password": "password123",
		"name": "PW User", "business_name": "PW Cafe3",
	}
	b, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	regReq.Header.Set("Content-Type", "application/json")
	regRr := httptest.NewRecorder()
	h.RegisterOwner(regRr, regReq)
	require.Equal(t, http.StatusCreated, regRr.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(regRr.Body).Decode(&regResp))
	userID := int64(regResp["user"].(map[string]interface{})["id"].(float64))

	pwBody := map[string]string{
		"current_password": "password123",
		"new_password":     "short",
	}
	pb, _ := json.Marshal(pwBody)
	req := httptest.NewRequest(http.MethodPost, "/api/profile/change-password", bytes.NewReader(pb))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req, userID, "owner")
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
```

- [ ] **Step 2: Rodar testes**

Run: `cd backend && go test ./internal/handler/ -run "TestGetProfile|TestUpdateProfile|TestChangePassword" -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
cd backend && git add internal/handler/profile_test.go
git commit -m "test: testes para GetProfile, UpdateProfile, ChangePassword"
```

---

### Task 15: Verificacao final

- [ ] **Step 1: Rodar todos os testes**

Run: `cd backend && go test ./... -count=1`
Expected: ALL PASS

- [ ] **Step 2: Verificar go vet**

Run: `cd backend && go vet ./...`
Expected: No warnings

- [ ] **Step 3: Verificar que nao ha map[string]interface{} nos handlers (alem de wrappers simples de lista)**

Run: `grep -n 'map\[string\]interface{}' internal/handler/*.go`
Expected: Minimal occurrences (only in `writeError` helper and list wrappers during transition)

- [ ] **Step 4: Verificar que nao ha context.Background() sem timeout nos handlers/services**

Run: `grep -rn 'context.Background()' internal/handler/ internal/service/`
Expected: Only inside functions that immediately call `context.WithTimeout`

- [ ] **Step 5: Verificar que nao ha constantes de negocio nos handlers**

Run: `grep -n 'const.*Limit\|const.*Tier\|const.*Discount\|const.*Grace' internal/handler/*.go`
Expected: No results

- [ ] **Step 6: Commit final**

```bash
git add -A && git commit -m "refactor: verificacao final do refactoring do backend"
```
