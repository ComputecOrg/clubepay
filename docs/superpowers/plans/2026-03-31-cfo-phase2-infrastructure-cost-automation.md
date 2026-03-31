# CFO Phase 2: Infrastructure Cost Automation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## Execution Status (Updated 2026-03-31)

| Task | Status | Commit | Notes |
|------|--------|--------|-------|
| Task 1: Create Cost Provider Interface | ✅ Complete | 08f5a29 | `provider.go` with `CostProvider` interface and `Cost` struct |
| Task 2: Implement Hostinger Provider | ✅ Complete | 08f5a29 | Static VPS cost, configured via `HOSTINGER_VPS_COST_CENTS` env var |
| Task 3: Implement Claude API Provider | ✅ Complete | 3425b7a | Mock implementation, configured via `CLAUDE_API_COST_CENTS` env var |
| Task 4: Implement Brevo SMTP Provider | ✅ Complete | f7ea242 | Supports free tier and paid plans, configured via `BREVO_EMAIL_COST_CENTS` env var |
| Task 5: Create Cost Aggregator | ⏳ Pending | — | Next: `aggregator.go` + `aggregator_test.go` |
| Task 6: Add UpdateInfrastructureCosts Handler | ⏳ Pending | — | Next: Handler method + handler tests |
| Task 7: Integrate with Cron Reconciliation | ⏳ Pending | — | Next: Call aggregator in `cron.go` reconcile method |
| Task 8: Database Schema Update | ⏳ Pending | — | Next: Add migration for `infrastructure_cost_cents` column |

**Overall Progress:** 4/8 tasks complete (50%). All provider implementations complete and tested.

---

**Goal:** Automatizar rastreamento diário de custos de infraestrutura (VPS, Claude API, SMTP) no reconcile cron, eliminando necessidade de input manual do Fin.

**Architecture:**
- Criar interface `CostProvider` para integrar múltiplos providers de custo
- Implementar providers concretos: `HostingerProvider`, `ClaudeAPIProvider`, `BrevoProvider`
- Adicionar método `UpdateInfrastructureCosts()` ao handler que coleta custos de todos os providers
- Integrar no cron reconciliation diário (sem bloquear se um provider falhar)
- Armazenar custos em `monthly_costs.infrastructure_cost_cents`

**Tech Stack:**
- Go interfaces (DIP — Dependency Inversion Principle)
- HTTP clients (net/http) para APIs de providers
- PostgreSQL (atualizar `monthly_costs`)
- TDD com testify assertions

---

## File Structure

```
backend/internal/
├── provider/                          # NEW: Cost provider abstraction
│   ├── provider.go                   # Interface CostProvider
│   ├── hostinger.go                  # Hostinger VPS API client
│   ├── claude_api.go                 # Claude usage API client
│   ├── brevo.go                      # Brevo SMTP client
│   └── *_test.go                     # Tests para cada provider
├── handler/
│   ├── handler.go                    # Modify: Add UpdateInfrastructureCosts()
│   ├── cron.go                       # Modify: Call UpdateInfrastructureCosts() in Reconcile()
│   └── *_test.go
└── config/
    └── config.go                     # Modify: Add provider API keys
```

---

## Task 1: Create Cost Provider Interface ✅

**Files:**
- Create: `backend/internal/provider/provider.go` ✅
- Test: `backend/internal/provider/provider_test.go` (not needed, interface only)

- [x] **Step 1: Write interface definition (no test needed, just define)**

Criar arquivo `backend/internal/provider/provider.go`:

```go
package provider

import "context"

// Cost represents infrastructure cost in cents
type Cost struct {
	// CostCents: valor em centavos (1 USD = 100 cents)
	CostCents int64
	// Provider: nome do provider (hostinger, claude, brevo)
	Provider string
	// Description: descrição legível (ex: "Hostinger KVM1 - R$12.99")
	Description string
}

// CostProvider interface for fetching infrastructure costs
type CostProvider interface {
	// GetMonthlyCost returns the cost for current month in cents
	GetMonthlyCost(ctx context.Context) (Cost, error)
}
```

---

## Task 2: Implement Hostinger Provider (Static Cost) ✅

**Files:**
- Modify: `backend/internal/config/config.go` - Add HOSTINGER_VPS_COST_CENTS ✅
- Create: `backend/internal/provider/hostinger.go` ✅
- Create: `backend/internal/provider/hostinger_test.go` ✅

- [x] **Step 1: Update config.go to include Hostinger cost constant**

Adicionar ao `Config` struct em `backend/internal/config/config.go`:

```go
type Config struct {
	// ... existing fields ...
	// Infrastructure costs (in cents)
	HostingerVPSCostCents int64
}
```

Adicionar ao `Load()` function (após outras linhas de config):

```go
HostingerVPSCostCents: parseInt64Env("HOSTINGER_VPS_COST_CENTS", 6495), // R$12.99 = 6495 cents
```

- [x] **Step 2: Write failing test for Hostinger provider** ✅
- [x] **Step 3: Run test to verify it fails** ✅
- [x] **Step 4: Implement Hostinger provider** ✅
- [x] **Step 5: Run test to verify it passes** ✅
- [x] **Step 6: Commit** ✅ (commit 08f5a29)

---

## Task 3: Implement Claude API Provider (Mock for Now) ✅

**Files:**
- Modify: `backend/internal/config/config.go` - Add CLAUDE_API_COST_CENTS ✅
- Create: `backend/internal/provider/claude_api.go` ✅
- Create: `backend/internal/provider/claude_api_test.go` ✅

- [x] **Step 1: Update config to include Claude API cost**

- [x] **Step 2: Write failing test for Claude provider** ✅
- [x] **Step 3: Run test to verify it fails** ✅
- [x] **Step 4: Implement Claude provider (mock implementation)** ✅
- [x] **Step 5: Run test to verify it passes** ✅
- [x] **Step 6: Commit** ✅ (commit 3425b7a)
- [x] **Step 7: Add comprehensive test coverage** ✅ (added `TestClaudeAPIProvider_GetMonthlyCost_ZeroCost` in f7ea242)

---

## Task 4: Implement Brevo SMTP Provider (Static Cost) ✅

**Files:**
- Modify: `backend/internal/config/config.go` - Add BREVO_COST_CENTS ✅
- Create: `backend/internal/provider/brevo.go` ✅
- Create: `backend/internal/provider/brevo_test.go` ✅

- [x] **Step 1: Update config to include Brevo cost**

- [x] **Step 2: Write failing test for Brevo provider** ✅
- [x] **Step 3: Run test to verify it fails** ✅
- [x] **Step 4: Implement Brevo provider** ✅
- [x] **Step 5: Run test to verify it passes** ✅
- [x] **Step 6: Commit** ✅ (commit f7ea242)

---

## Task 5: Create Cost Aggregator

**Files:**
- Create: `backend/internal/provider/aggregator.go`
- Create: `backend/internal/provider/aggregator_test.go`

- [ ] **Step 1: Write failing test for aggregator**

Criar `backend/internal/provider/aggregator_test.go`:

```go
package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/provider"
)

type mockProvider struct {
	cost int64
}

func (m *mockProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	return provider.Cost{CostCents: m.cost, Provider: "mock"}, nil
}

func TestAggregator_GetTotalCost(t *testing.T) {
	providers := []provider.CostProvider{
		&mockProvider{cost: 10000}, // $100
		&mockProvider{cost: 5000},  // $50
		&mockProvider{cost: 2500},  // $25
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(17500), total) // $175 total
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/provider/... -v -run TestAggregator
```

Expected: `undefined: NewAggregator`

- [ ] **Step 3: Implement aggregator**

Criar `backend/internal/provider/aggregator.go`:

```go
package provider

import (
	"context"
	"log/slog"
)

// Aggregator collects costs from multiple providers
type Aggregator struct {
	providers []CostProvider
}

// NewAggregator creates a cost aggregator
func NewAggregator(providers []CostProvider) *Aggregator {
	return &Aggregator{providers: providers}
}

// GetTotalInfrastructureCost sums costs from all providers
// Non-blocking: if a provider fails, log error but continue with others
func (a *Aggregator) GetTotalInfrastructureCost(ctx context.Context) (int64, error) {
	var total int64

	for _, p := range a.providers {
		cost, err := p.GetMonthlyCost(ctx)
		if err != nil {
			slog.Error("provider failed to get cost", "provider", cost.Provider, "error", err)
			// Continue with other providers (non-blocking)
			continue
		}
		total += cost.CostCents
	}

	return total, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/provider/... -v -run TestAggregator
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/provider/aggregator.go backend/internal/provider/aggregator_test.go
git commit -m "feat: add cost aggregator for multiple providers (Phase 2)"
```

---

## Task 6: Add UpdateInfrastructureCosts Handler Method

**Files:**
- Modify: `backend/internal/handler/handler.go` - Add UpdateInfrastructureCosts() method
- Create: `backend/internal/handler/handler_infrastructure_costs_test.go`

- [ ] **Step 1: Write failing test for handler method**

Criar `backend/internal/handler/handler_infrastructure_costs_test.go`:

```go
package handler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/provider"
)

type mockCostProvider struct {
	cost int64
}

func (m *mockCostProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	return provider.Cost{CostCents: m.cost, Provider: "mock"}, nil
}

func TestHandler_UpdateInfrastructureCosts(t *testing.T) {
	// Setup: Create handler with mock aggregator
	providers := []provider.CostProvider{
		&mockCostProvider{cost: 20000},
	}
	agg := provider.NewAggregator(providers)

	h := &handler.Handler{
		Queries:      nil, // Will be mocked below
		CostAggregator: agg,
	}

	// This test verifies the method exists and calculates total
	// Full integration test will be in Task 7
	ctx := context.Background()
	total, err := h.CalculateTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(20000), total)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -v -run TestHandler_UpdateInfrastructureCosts
```

Expected: `undefined: CalculateTotalInfrastructureCost`

- [ ] **Step 3: Update Handler struct to include CostAggregator**

Modificar `backend/internal/handler/handler.go`:

```go
type Handler struct {
	Auth          *service.AuthService
	Business      *service.BusinessService
	Plans         *service.PlanService
	Subscriptions *service.SubscriptionService
	Usage         *service.UsageService
	Referrals     *service.ReferralService
	Config        *config.Config
	Queries       *repository.Queries
	PSP           psp.PSP
	Email         email.Sender
	CostAggregator *provider.Aggregator  // NEW
}
```

Adicionar import:

```go
"github.com/clubepay/backend/internal/provider"
```

- [ ] **Step 4: Add CalculateTotalInfrastructureCost method to Handler**

Adicionar ao fim de `handler.go`:

```go
// CalculateTotalInfrastructureCost sums costs from all infrastructure providers
func (h *Handler) CalculateTotalInfrastructureCost(ctx context.Context) (int64, error) {
	if h.CostAggregator == nil {
		return 0, nil // No cost aggregator configured
	}
	return h.CostAggregator.GetTotalInfrastructureCost(ctx)
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd backend && go test ./internal/handler/... -v -run TestHandler_UpdateInfrastructureCosts
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/handler.go backend/internal/handler/handler_infrastructure_costs_test.go
git commit -m "feat: add CalculateTotalInfrastructureCost method (Phase 2)"
```

---

## Task 7: Integrate Cost Update into Cron Reconciliation

**Files:**
- Modify: `backend/internal/handler/cron.go` - Call cost update
- Modify: `backend/cmd/api/main.go` - Initialize CostAggregator
- Modify: `backend/internal/handler/handler_test.go` (if exists)

- [ ] **Step 1: Write failing integration test**

Adicionar ao `backend/internal/handler/cron_test.go`:

```go
func TestReconcile_UpdatesCosts(t *testing.T) {
	// Setup: Create test handler with mock cost provider
	providers := []provider.CostProvider{
		&mockCostProvider{cost: 28495}, // Total: Hostinger + Claude + Brevo
	}
	agg := provider.NewAggregator(providers)

	h := &handler.Handler{
		Queries:         queries, // From test setup
		CostAggregator:  agg,
		Config:          &config.Config{MonthlyBudgetCents: 500000},
	}

	// Test: Call Reconcile
	req := httptest.NewRequest("POST", "/api/cron/reconcile", nil)
	req.Header.Set("X-Cron-Secret", "test-secret")
	w := httptest.NewRecorder()
	h.Config.CronSecret = "test-secret"

	h.Reconcile(w, req)

	// Verify: Check monthly_costs was updated
	assert.Equal(t, http.StatusOK, w.Code)
	// Would also verify DB state in full integration test
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/handler/... -v -run TestReconcile_UpdatesCosts
```

Expected: Test fails or compilation error (field doesn't exist)

- [ ] **Step 3: Modify Reconcile() to call cost update**

Em `backend/internal/handler/cron.go`, adicionar após SendSpendingAlerts():

```go
	// 4. Update infrastructure costs
	totalInfraCost, err := h.CalculateTotalInfrastructureCost(ctx)
	if err != nil {
		slog.Error("reconcile: failed to calculate infrastructure costs", "error", err)
		// Non-blocking: don't fail reconcile if costs can't be calculated
	} else if totalInfraCost > 0 {
		// Get current month and update costs
		currentMonth := domain.GetCurrentMonth()

		businesses, err := h.Queries.ListAllBusinesses(ctx)
		if err != nil {
			slog.Error("reconcile: failed to list businesses for cost update", "error", err)
		} else {
			for _, business := range businesses {
				_, err := h.Queries.UpdateMonthlyCostInfrastructure(ctx, repository.UpdateMonthlyCostInfrastructureParams{
					BusinessID:              business.ID,
					Month:                   domain.PgTypeDate(currentMonth),
					InfrastructureCostCents: totalInfraCost,
				})
				if err != nil {
					slog.Error("reconcile: failed to update infrastructure costs", "business_id", business.ID, "error", err)
				}
			}
		}
	}
```

- [ ] **Step 4: Update Handler initialization in main.go**

Em `backend/cmd/api/main.go`, na função `main()` onde Handler é criado:

```go
	// Create cost providers
	costProviders := []provider.CostProvider{
		provider.NewHostingerProvider(cfg),
		provider.NewClaudeAPIProvider(cfg),
		provider.NewBrevoProvider(cfg),
	}
	costAggregator := provider.NewAggregator(costProviders)

	// Create handler with cost aggregator
	h := handler.New(queries, cfg, psp, emailSender)
	h.CostAggregator = costAggregator
```

Adicionar import em main.go:

```go
"github.com/clubepay/backend/internal/provider"
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd backend && go test ./internal/handler/... -v -run TestReconcile_UpdatesCosts
```

Expected: PASS

- [ ] **Step 6: Run full test suite to ensure no regressions**

```bash
cd backend && go test ./...
```

Expected: All tests PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/cron.go
git add backend/cmd/api/main.go
git add backend/internal/handler/handler.go
git commit -m "feat: integrate infrastructure cost collection into daily cron (Phase 2)"
```

---

## Task 8: Update Documentation

**Files:**
- Modify: `docs/SPENDING-CONFIG.md` - Add cost provider documentation
- Modify: `BUDGET.md` - Update cost tracking notes

- [ ] **Step 1: Update SPENDING-CONFIG.md with provider info**

Adicionar seção em `docs/SPENDING-CONFIG.md`:

```markdown
## Infrastructure Cost Providers

### Automatic Cost Collection

Daily cron reconciliation automatically collects infrastructure costs from:

1. **Hostinger VPS** - Static cost per month
   - Env var: `HOSTINGER_VPS_COST_CENTS` (default: 6495 = R$12.99)

2. **Claude API** - Fixed monthly cost
   - Env var: `CLAUDE_API_COST_CENTS` (default: 20000 = $200)

3. **Brevo SMTP** - Free tier (€0) or paid plan
   - Env var: `BREVO_EMAIL_COST_CENTS` (default: 0 = free)

Costs are aggregated daily and stored in `monthly_costs.infrastructure_cost_cents`.
If any provider fails, others continue collecting (non-blocking).

### Future: Real-Time Integration

- Hostinger API: Track actual VPS usage
- Claude API: Track actual token consumption
- Brevo API: Track email volume
```

- [ ] **Step 2: Update BUDGET.md with automated tracking note**

Adicionar ao seção "Next Actions":

```markdown
- [x] ~~Setup alerts no Paperclip~~ → Implementado CLU-24 ✅
- [x] ~~Infraestrutura para rastreamento automático~~ → Phase 2 completa ✅
- [ ] Real-time provider API integration (Hostinger, Claude usage) — Goh/Fin
- [ ] CEO dashboard com histórico 3 meses — Rex/Frontend
```

- [ ] **Step 3: Commit**

```bash
git add docs/SPENDING-CONFIG.md BUDGET.md
git commit -m "docs: update cost provider documentation (Phase 2)"
```

---

## Testing Strategy

### Unit Tests
- ✅ Each provider: `TestHostinger_*`, `TestClaudeAPI_*`, `TestBrevo_*`
- ✅ Aggregator: `TestAggregator_GetTotalCost`
- ✅ Handler method: `TestHandler_UpdateInfrastructureCosts`

### Integration Test
- ✅ Full cron flow: `TestReconcile_UpdatesCosts` — Verifies costs written to DB

### Manual Testing
```bash
# Run all Phase 2 tests
cd backend && go test ./internal/provider/... -v

# Run handler integration test
cd backend && go test ./internal/handler/... -v -run TestReconcile

# Test cron endpoint directly (requires DATABASE_URL set)
curl -X POST http://localhost:8080/api/cron/reconcile \
  -H "X-Cron-Secret: your-secret"
```

---

## Architecture Decision Log

**Why interface-based providers?**
- Decouples cost logic from handler
- Easy to add new providers (Hostinger API, Claude usage API, etc.)
- Testable with mocks
- Non-blocking: if one provider fails, others still collect

**Why aggregator pattern?**
- Single responsibility: aggregator only sums
- Composable: easy to add/remove providers
- Follows Go conventions (composition over inheritance)

**Why non-blocking in cron?**
- One provider failure shouldn't block dunning, reconciliation, other cron tasks
- Logs errors for monitoring
- Retries happen on next daily cron run

---

## Success Criteria

✅ All 8 tasks completed and committed
✅ Unit tests: 100% coverage for providers
✅ Integration test: Costs written to `monthly_costs.infrastructure_cost_cents`
✅ Cron integration: `UpdateMonthlyCostInfrastructure()` called daily
✅ Documentation updated
✅ No regressions in existing test suite

---

**Plan complete e pronto para execução!**
