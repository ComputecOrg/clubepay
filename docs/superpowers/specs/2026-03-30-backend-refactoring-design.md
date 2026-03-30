# Backend Refactoring — Design Spec

**Data:** 2026-03-30
**Escopo:** 10 melhorias de arquitetura e qualidade de codigo no backend Go
**Abordagem:** Incremental por camadas (4 fases, cada uma compilavel e testavel)

---

## Contexto

Auditoria do backend identificou 10 problemas de arquitetura e qualidade:

1. Ausencia de camada de servico — handlers com 135 linhas de logica misturada
2. Pattern "GetBusinessByOwnerID" duplicado 10+ vezes
3. 16 ocorrencias de `map[string]interface{}` em vez de structs tipadas
4. Constantes de negocio espalhadas em 3 handlers diferentes
5. Validacao inconsistente (validator tags vs manual campo-a-campo)
6. Goroutines com `context.Background()` sem timeout (4 ocorrencias)
7. Sem testes para `profile.go` e `config.go`
8. Logica de cancelamento PSP duplicada em 2 handlers
9. Erros de email ignorados silenciosamente (4 ocorrencias)
10. Structs de input anonimas inline em 9 handlers

---

## Fase 1: Fundacoes

Novos arquivos que nao mudam comportamento existente. Tudo compila, testes passam.

### 1.1 `domain/constants.go`

Centraliza constantes de negocio hoje espalhadas em `handler/plan.go:15`, `handler/subscription.go:21`, `handler/referral.go:15` e magic numbers hardcoded.

```go
package domain

const (
    FreeTierPlanLimit       = 1
    FreeTierSubscriberLimit = 15
    ReferralLimit           = 3
    ReferralDiscountPercent = 10
    GracePeriodDays         = 3
    OwnerJWTExpiryHours     = 24
    SubscriberJWTExpiryDays = 30
)
```

### 1.2 `domain/inputs.go`

Substitui as 9 structs anonimas inline dos handlers por types nomeados com validator tags. Unifica a estrategia de validacao — tudo usa `domain.Validate()`.

Novos types:
- `CreatePlanInput` — name (required), description, price_cents (gt=0), limit_type (oneof=daily monthly), limit_count (gt=0)
- `UpdatePlanInput` = `CreatePlanInput` (mesmos campos, mesmas regras)
- `SubscribeInput` — plan_id (gt=0)
- `ApplyReferralInput` — code (required)
- `UpdateProfileInput` — name (required), phone
- `ChangePasswordInput` — current_password (required), new_password (required, min=8)
- `ValidateUsageOwnerInput` — subscriber_id (gt=0), slug (required)
- `RequestPasswordResetInput` — email (required, email)
- `ConfirmPasswordResetInput` — token (required), password (required, min=8)

### 1.3 `domain/responses.go`

Substitui os 16 `map[string]interface{}` por structs tipadas com json tags.

Structs:
- `AuthResponse` — token, user (UserResponse), business (BusinessResponse, omitempty)
- `UserResponse` — id, email, name, phone (omitempty), role
- `BusinessResponse` — id, name, slug, segment, address, logo_url
- `PlanResponse` — id, name, description, price_cents, limit_type, limit_count, active
- `PlanListResponse` — plans []PlanResponse
- `SubscriptionResponse` — id, plan_id, subscriber_id, status, period_end, discount_percent
- `SubscriptionListResponse` — subscriptions []SubscriptionResponse
- `MyPlanResponse` — plan (PlanResponse), business (BusinessResponse), subscription (SubscriptionInfo)
- `UsageResponse` — plan_name, limit_type, limit_count, current_count, period_start, period_end
- `ValidateUsageResponse` — current_count, limit, plan_name
- `ReferralCodeResponse` — code
- `CancelResponse` — status, period_end, message
- `ReconcileResponse` — blocked, synced
- `ProfileResponse` — id, email, name, phone, role

### 1.4 `domain/service_errors.go`

Erros tipados para a service layer mapear para HTTP status:

```go
type ServiceError struct {
    Code    int
    Message string
    Err     error
}
```

Construtores: `ErrNotFound()`, `ErrConflict()`, `ErrForbidden()`, `ErrBadRequest()`, `ErrInternal()`

Handler helper: `handleServiceError()` faz `errors.As()` e mapeia para `writeError()`.

---

## Fase 2: Service Layer

Cria `internal/service/` e migra logica de negocio dos handlers.

### 2.1 Estrutura

```
internal/service/
  auth.go          — AuthService (Queries, Config, Email)
  business.go      — BusinessService (Queries)
  plan.go          — PlanService (Queries)
  subscription.go  — SubscriptionService (Queries, PSP, Email)
  usage.go         — UsageService (Queries)
  referral.go      — ReferralService (Queries)
```

Cada service recebe apenas as dependencias que usa.

### 2.2 Migracao por service

**AuthService:**
- `RegisterOwner(ctx, input) (*AuthResponse, error)` — hash, criar user, gerar slug, criar business, gerar JWT
- `RegisterSubscriber(ctx, input) (*AuthResponse, error)`
- `Login(ctx, input) (*AuthResponse, error)`
- `RequestPasswordReset(ctx, input) error` — gera token, dispara email async com timeout
- `ConfirmPasswordReset(ctx, input) error`

**BusinessService:**
- `GetByOwner(ctx, ownerID) (*BusinessResponse, error)` — o pattern repetido 10x
- `Update(ctx, ownerID, input) (*BusinessResponse, error)`

**PlanService:**
- `Create(ctx, ownerID, input) (*PlanResponse, error)` — busca business, valida free tier, cria
- `List(ctx, ownerID) (*PlanListResponse, error)`
- `Update(ctx, ownerID, planID, input) (*PlanResponse, error)` — busca business, verifica ownership
- `Delete(ctx, ownerID, planID) error`

**SubscriptionService:**
- `Subscribe(ctx, subscriberID, input) (*SubscriptionResponse, error)` — toda a orquestracao (135 linhas do handler atual)
- `ListByBusiness(ctx, ownerID) (*SubscriptionListResponse, error)`
- `CancelByOwner(ctx, ownerID, subID) error` — unifica logica de cancelamento
- `CancelBySubscriber(ctx, subscriberID) (*CancelResponse, error)` — mesma logica, perspectiva do assinante
- `GetMyPlan(ctx, subscriberID) (*MyPlanResponse, error)`

**UsageService:**
- `Validate(ctx, subscriberID, input) (*ValidateUsageResponse, error)`
- `ValidateByOwner(ctx, ownerID, input) (*ValidateUsageResponse, error)`
- `GetMyUsage(ctx, subscriberID) (*UsageResponse, error)`

**ReferralService:**
- `GetOrCreateCode(ctx, subscriberID) (*ReferralCodeResponse, error)`
- `Apply(ctx, subscriberID, input) error`

### 2.3 Handler struct atualizada

```go
type Handler struct {
    Auth          *service.AuthService
    Business      *service.BusinessService
    Plans         *service.PlanService
    Subscriptions *service.SubscriptionService
    Usage         *service.UsageService
    Referrals     *service.ReferralService
    Config        *config.Config
}
```

Queries, PSP, Email removidos do Handler — vivem nos services.

### 2.4 Handlers ficam finos

Padrao para todo handler:
1. Extrair userID do context
2. Parse + validate input
3. Chamar service
4. handleServiceError ou writeJSON

Exemplo — Subscribe passa de 135 linhas para ~15.

---

## Fase 3: Cleanup dos Handlers

### 3.1 Goroutines com context controlado

As 4 goroutines com `context.Background()` migram para methods privados nos services com:
- `context.WithTimeout(context.Background(), 10*time.Second)`
- Log de erros com slog (tipo do email, destinatario, erro)
- Handler nao lida mais com goroutines

Locais afetados:
- `subscription.go:151-157` (welcome email) -> `SubscriptionService.sendWelcomeEmail()`
- `subscriber.go:92-103` (cancellation email) -> `SubscriptionService.sendCancellationEmail()`
- `webhook.go:93-101` (payment confirmed email) -> `SubscriptionService.sendPaymentConfirmedEmail()`
- `auth.go:234-237` (password reset email) -> `AuthService.sendPasswordResetEmail()`

### 3.2 Erros de email logados

Padrao unico em todos os services:
```go
if err := s.Email.Send(to, subject, body); err != nil {
    slog.Error("failed to send email", "error", err, "to", to, "type", emailType)
}
```

### 3.3 Webhook handler

O webhook handler (`webhook.go`) continua recebendo Queries e PSP diretamente (nao via service) porque ele e um ponto de entrada especial que precisa ler o body raw e validar HMAC. A logica de atualizar status e enviar email migra para `SubscriptionService.HandleWebhookEvent()`.

---

## Fase 4: Testes

### 4.1 `config/config_test.go`

- Env vars obrigatorias ausentes (DATABASE_URL, JWT_SECRET) -> erro
- Valores default (Port=8080, AsaasURL sandbox, SMTPPort=587, etc.)
- Todas as vars presentes -> config completa
- Usa `t.Setenv()` para set/unset env vars

### 4.2 `handler/profile_test.go`

Mesma infraestrutura dos testes existentes (testcontainers + httptest + testify):
- GetProfile: sucesso, user nao encontrado
- UpdateProfile: sucesso, body invalido, nome vazio
- ChangePassword: sucesso, senha atual incorreta, nova senha curta (<8)

### 4.3 `service/*_test.go`

Cada service ganha tests cobrindo a logica de negocio:
- Mock PSP + Mock Email (ja existem em `psp/mock.go` e `email/email.go`)
- Testcontainers para queries reais
- Table-driven tests para validacoes e edge cases
- Testes de concorrencia basicos (subscribe paralelo)

### 4.4 Atualizacao dos handler tests existentes

Os 14 handler test files existentes sao adaptados para injetar services em vez de Queries/PSP/Email. A logica dos testes permanece — so muda a construcao do Handler no setup.

---

## Arquivos criados/modificados

### Novos (Fase 1):
- `internal/domain/constants.go`
- `internal/domain/inputs.go`
- `internal/domain/responses.go`
- `internal/domain/service_errors.go`

### Novos (Fase 2):
- `internal/service/auth.go`
- `internal/service/business.go`
- `internal/service/plan.go`
- `internal/service/subscription.go`
- `internal/service/usage.go`
- `internal/service/referral.go`

### Novos (Fase 4):
- `internal/config/config_test.go`
- `internal/handler/profile_test.go`
- `internal/service/auth_test.go`
- `internal/service/business_test.go`
- `internal/service/plan_test.go`
- `internal/service/subscription_test.go`
- `internal/service/usage_test.go`
- `internal/service/referral_test.go`

### Modificados:
- `internal/handler/handler.go` — Handler struct usa services, add handleServiceError()
- `internal/handler/auth.go` — handlers finos chamando AuthService
- `internal/handler/plan.go` — handlers finos chamando PlanService
- `internal/handler/subscription.go` — handlers finos chamando SubscriptionService
- `internal/handler/subscriber.go` — handlers finos chamando SubscriptionService
- `internal/handler/business.go` — handlers finos chamando BusinessService
- `internal/handler/usage.go` — handlers finos chamando UsageService
- `internal/handler/referral.go` — handlers finos chamando ReferralService
- `internal/handler/webhook.go` — usa SubscriptionService.HandleWebhookEvent()
- `internal/handler/profile.go` — usa AuthService/BusinessService
- `internal/handler/public.go` — usa BusinessService/PlanService
- `internal/handler/search.go` — usa BusinessService
- `internal/handler/cron.go` — usa SubscriptionService
- `internal/handler/owner_validate.go` — usa UsageService
- `cmd/api/main.go` — instancia services, passa pro Handler
- `cmd/api/router.go` — sem mudanca (rotas iguais)
- Todos os `*_test.go` existentes — setup atualizado para injetar services

### Removidos:
- Constantes locais em handler/plan.go, handler/subscription.go, handler/referral.go
- Structs anonimas inline em 9 handlers
- `map[string]interface{}` em 16 handlers

---

## Criterios de sucesso

1. `go build ./...` compila sem erros em cada fase
2. `go test ./...` passa em cada fase
3. `go vet ./...` sem warnings
4. Zero `map[string]interface{}` nos handlers
5. Zero `context.Background()` em goroutines de handler/service (exceto com timeout)
6. Zero constantes de negocio fora de `domain/constants.go`
7. Zero validacao manual campo-a-campo (tudo via validator tags)
8. Cobertura de testes para profile.go e config.go
9. Todos os erros de email logados
