# ClubePay — Plano de Produção 100%

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Levar o ClubePay de ~70% para 100% de prontidão para produção, cobrindo segurança, funcionalidade core, UX/frontend e infraestrutura/DevOps.

**Architecture:** Melhorias incrementais sobre o stack existente (Go/chi + Next.js + PostgreSQL + Asaas). Cada task é independente e produz código testável. Segue TDD conforme CLAUDE.md.

**Tech Stack:** Go 1.22+ | chi | sqlc | pgx | JWT | Next.js 15+ | TypeScript | Tailwind CSS 4 | PostgreSQL 16 | Asaas API | GitHub Actions | Docker

---

## File Structure

### Backend — Files to Create/Modify

| File | Responsibility |
|------|---------------|
| `backend/internal/middleware/security.go` | Security headers middleware |
| `backend/internal/middleware/security_test.go` | Tests do security headers |
| `backend/internal/middleware/cors.go` | CORS com origins permitidas (modificar) |
| `backend/internal/middleware/cors_test.go` | Tests do CORS |
| `backend/internal/middleware/ratelimit.go` | Rate limiting por IP |
| `backend/internal/middleware/ratelimit_test.go` | Tests do rate limit |
| `backend/internal/config/config.go` | Adicionar novos env vars (modificar) |
| `backend/internal/handler/auth.go` | Adicionar password reset request/confirm (modificar) |
| `backend/internal/handler/auth_test.go` | Tests dos novos endpoints (modificar) |
| `backend/internal/handler/profile.go` | Endpoints de perfil (update name, change password) |
| `backend/internal/handler/profile_test.go` | Tests de perfil |
| `backend/internal/handler/subscription.go` | Aplicar desconto referral no subscribe (modificar) |
| `backend/internal/handler/subscription_test.go` | Tests do desconto (modificar) |
| `backend/internal/email/templates.go` | Templates HTML de email |
| `backend/internal/email/templates_test.go` | Tests dos templates |
| `backend/internal/psp/psp.go` | Adicionar Discount ao CreateSubscriptionRequest (modificar) |
| `backend/migrations/000002_add_discount_and_reset.up.sql` | Migration: discount + password reset |
| `backend/migrations/000002_add_discount_and_reset.down.sql` | Rollback migration |
| `backend/queries/users.sql` | Adicionar queries de password reset (modificar) |
| `backend/queries/subscriptions.sql` | Adicionar query com discount (modificar) |
| `backend/cmd/api/main.go` | Auto-migrate on startup (modificar) |
| `backend/cmd/api/router.go` | Novos endpoints (modificar) |

### Frontend — Files to Create/Modify

| File | Responsibility |
|------|---------------|
| `frontend/src/app/error.tsx` | Error boundary global |
| `frontend/src/app/not-found.tsx` | Página 404 |
| `frontend/src/app/(auth)/loading.tsx` | Loading skeleton para rotas auth |
| `frontend/src/middleware.ts` | Route protection no edge |
| `frontend/src/app/(auth)/perfil/page.tsx` | Página de perfil/conta |
| `frontend/src/app/(auth)/esqueci-senha/page.tsx` | Solicitar reset de senha |
| `frontend/src/app/(auth)/resetar-senha/page.tsx` | Confirmar reset com token |
| `frontend/next.config.ts` | Security headers (modificar) |
| `frontend/src/lib/api.ts` | Retry logic (modificar) |

### Infra — Files to Create/Modify

| File | Responsibility |
|------|---------------|
| `.github/workflows/ci.yml` | CI pipeline (test + build) |
| `.github/workflows/deploy.yml` | CD pipeline (push to deploy) |
| `Dockerfile.backend` | Health check + non-root user (modificar) |
| `Dockerfile.frontend` | Health check + non-root user (modificar) |
| `docker-compose.yml` | Melhorias de segurança (modificar) |
| `docker-compose.prod.yml` | Config de produção |
| `.env.example` | Documentação de env vars |
| `scripts/backup.sh` | Script de backup PostgreSQL |

---

## FASE 1: SEGURANCA (55% → 100%)

---

### Task 1: Security Headers Middleware (Backend)

**Files:**
- Create: `backend/internal/middleware/security.go`
- Create: `backend/internal/middleware/security_test.go`

- [ ] **Step 1: Write the failing test**

```go
// backend/internal/middleware/security_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	assert.Contains(t, rec.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, http.StatusOK, rec.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/middleware/ -run TestSecurityHeaders -v`
Expected: FAIL — `SecurityHeaders` not defined

- [ ] **Step 3: Write minimal implementation**

```go
// backend/internal/middleware/security.go
package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/middleware/ -run TestSecurityHeaders -v`
Expected: PASS

- [ ] **Step 5: Wire into router**

Modify `backend/cmd/api/router.go` — add after `middleware.CORS`:

```go
r.Use(middleware.SecurityHeaders)
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/middleware/security.go backend/internal/middleware/security_test.go backend/cmd/api/router.go
git commit -m "feat: security headers middleware (X-Frame-Options, CSP, etc.)"
```

---

### Task 2: CORS com Allowlist de Origins

**Files:**
- Modify: `backend/internal/middleware/cors.go`
- Create: `backend/internal/middleware/cors_test.go`
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: Write the failing test**

```go
// backend/internal/middleware/cors_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	handler := CORS("https://clubepay.com,https://www.clubepay.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://clubepay.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "https://clubepay.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	handler := CORS("https://clubepay.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_Preflight(t *testing.T) {
	handler := CORS("https://clubepay.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://clubepay.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://clubepay.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardForDev(t *testing.T) {
	handler := CORS("*")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/middleware/ -run TestCORS -v`
Expected: FAIL — `CORS` function signature changed

- [ ] **Step 3: Update config to include CORS_ORIGINS**

Modify `backend/internal/config/config.go` — add field to Config struct:

```go
type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	AsaasAPIKey        string
	AsaasURL           string
	CronSecret         string
	AsaasWebhookSecret string
	SMTPHost           string
	SMTPPort           string
	SMTPUsername       string
	SMTPPassword       string
	CORSOrigins        string
	FrontendURL        string
}
```

Add to `Load()`:

```go
CORSOrigins: getEnv("CORS_ORIGINS", "*"),
FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
```

- [ ] **Step 4: Implement CORS with allowlist**

Replace `backend/internal/middleware/cors.go`:

```go
package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, o := range origins {
					if o == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 5: Update router to pass config**

Modify `backend/cmd/api/router.go` — change CORS line:

```go
r.Use(middleware.CORS(cfg.CORSOrigins))
```

- [ ] **Step 6: Run tests**

Run: `cd backend && go test ./internal/middleware/ -run TestCORS -v`
Expected: PASS

- [ ] **Step 7: Run all tests to ensure no regressions**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 8: Commit**

```bash
git add backend/internal/middleware/cors.go backend/internal/middleware/cors_test.go backend/internal/config/config.go backend/cmd/api/router.go
git commit -m "fix: CORS com allowlist de origins (bloqueia origins nao autorizadas)"
```

---

### Task 3: Rate Limiting Middleware

**Files:**
- Create: `backend/internal/middleware/ratelimit.go`
- Create: `backend/internal/middleware/ratelimit_test.go`
- Modify: `backend/cmd/api/router.go`

- [ ] **Step 1: Write the failing test**

```go
// backend/internal/middleware/ratelimit_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	handler := RateLimit(5, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	handler := RateLimit(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	handler := RateLimit(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest("POST", "/", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest("POST", "/", nil)
	req2.RemoteAddr = "5.6.7.8:5678"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/middleware/ -run TestRateLimit -v`
Expected: FAIL — `RateLimit` not defined

- [ ] **Step 3: Write implementation**

```go
// backend/internal/middleware/ratelimit.go
package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type ipEntry struct {
	count    int
	expireAt time.Time
}

func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	entries := make(map[string]*ipEntry)

	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			now := time.Now()
			for ip, e := range entries {
				if now.After(e.expireAt) {
					delete(entries, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}

			mu.Lock()
			entry, exists := entries[ip]
			now := time.Now()

			if !exists || now.After(entry.expireAt) {
				entries[ip] = &ipEntry{count: 1, expireAt: now.Add(window)}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			entry.count++
			if entry.count > maxRequests {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				writeErr(w, http.StatusTooManyRequests, "muitas requisicoes, tente novamente em 1 minuto")
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./internal/middleware/ -run TestRateLimit -v`
Expected: PASS

- [ ] **Step 5: Apply rate limit to auth endpoints in router**

Modify `backend/cmd/api/router.go` — wrap auth routes:

```go
// Auth (no auth, rate limited)
r.Group(func(r chi.Router) {
	r.Use(middleware.RateLimit(10, time.Minute))
	r.Post("/api/auth/register", h.RegisterOwner)
	r.Post("/api/auth/login", h.Login)
	r.Post("/api/auth/register-subscriber", h.RegisterSubscriber)
})
```

Add `"time"` to imports in router.go.

- [ ] **Step 6: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/middleware/ratelimit.go backend/internal/middleware/ratelimit_test.go backend/cmd/api/router.go
git commit -m "feat: rate limiting nos endpoints de auth (10 req/min por IP)"
```

---

### Task 4: Security Headers no Next.js

**Files:**
- Modify: `frontend/next.config.ts`

- [ ] **Step 1: Update next.config.ts with security headers**

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-XSS-Protection", value: "1; mode=block" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
```

- [ ] **Step 2: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add frontend/next.config.ts
git commit -m "feat: security headers no Next.js (X-Frame-Options, CSP, etc.)"
```

---

### Task 5: Docker — Non-root User + Health Checks

**Files:**
- Modify: `Dockerfile.backend`
- Modify: `Dockerfile.frontend`

- [ ] **Step 1: Update backend Dockerfile**

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /api .
COPY backend/migrations ./migrations

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./api"]
```

- [ ] **Step 2: Update frontend Dockerfile**

```dockerfile
FROM node:22-alpine AS builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

FROM node:22-alpine

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 3000

ENV HOSTNAME="0.0.0.0"

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

CMD ["node", "server.js"]
```

- [ ] **Step 3: Verify build**

Run: `docker build -f Dockerfile.backend -t clubepay-api:test . && docker build -f Dockerfile.frontend -t clubepay-web:test .`
Expected: Both build successfully

- [ ] **Step 4: Commit**

```bash
git add Dockerfile.backend Dockerfile.frontend
git commit -m "fix: Docker non-root user + health checks nos containers"
```

---

## FASE 2: FUNCIONALIDADE CORE (90% → 100%)

---

### Task 6: Migration — Desconto + Password Reset Token

**Files:**
- Create: `backend/migrations/000002_add_discount_and_reset.up.sql`
- Create: `backend/migrations/000002_add_discount_and_reset.down.sql`

- [ ] **Step 1: Write the up migration**

```sql
-- backend/migrations/000002_add_discount_and_reset.up.sql

-- Discount tracking on subscriptions
ALTER TABLE subscriptions ADD COLUMN discount_percent INT NOT NULL DEFAULT 0;

-- Password reset tokens
CREATE TABLE password_resets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    token VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_resets_token ON password_resets(token);
CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);

-- Profile updates: add updated_at to users
ALTER TABLE users ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
```

- [ ] **Step 2: Write the down migration**

```sql
-- backend/migrations/000002_add_discount_and_reset.down.sql

ALTER TABLE subscriptions DROP COLUMN IF EXISTS discount_percent;
DROP TABLE IF EXISTS password_resets;
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;
```

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/000002_add_discount_and_reset.up.sql backend/migrations/000002_add_discount_and_reset.down.sql
git commit -m "feat: migration para desconto em subscriptions + password reset tokens"
```

---

### Task 7: SQLC Queries — Desconto + Password Reset + Profile

**Files:**
- Modify: `backend/queries/subscriptions.sql`
- Modify: `backend/queries/users.sql`
- Create: `backend/queries/password_resets.sql`

- [ ] **Step 1: Add discount to CreateSubscription query**

Append to `backend/queries/subscriptions.sql`:

```sql
-- name: CreateSubscriptionWithDiscount :one
INSERT INTO subscriptions (plan_id, subscriber_id, business_id, psp_subscription_id, status, period_end, referred_by, discount_percent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
```

- [ ] **Step 2: Add profile update and password reset queries to users.sql**

Append to `backend/queries/users.sql`:

```sql
-- name: UpdateUserProfile :one
UPDATE users SET name = $2, phone = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW()
WHERE id = $1;
```

- [ ] **Step 3: Create password_resets queries**

```sql
-- backend/queries/password_resets.sql

-- name: CreatePasswordReset :one
INSERT INTO password_resets (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetByToken :one
SELECT * FROM password_resets
WHERE token = $1 AND used = false AND expires_at > NOW();

-- name: MarkPasswordResetUsed :exec
UPDATE password_resets SET used = true WHERE id = $1;

-- name: DeleteExpiredPasswordResets :exec
DELETE FROM password_resets WHERE expires_at < NOW();
```

- [ ] **Step 4: Add referral lookup query to referrals.sql**

Append to `backend/queries/referrals.sql`:

```sql
-- name: GetReferralByReferredAndBusiness :one
SELECT * FROM referrals
WHERE referred_id = $1 AND business_id = $2 AND referrer_id != referred_id AND active = true
LIMIT 1;
```

- [ ] **Step 5: Regenerate sqlc**

Run: `cd backend && sqlc generate`
Expected: No errors

- [ ] **Step 6: Commit**

```bash
git add backend/queries/ backend/internal/repository/
git commit -m "feat: queries sqlc para desconto, password reset e profile update"
```

---

### Task 8: Aplicar Desconto de Referral no Subscribe

**Files:**
- Modify: `backend/internal/handler/subscription.go`
- Modify: `backend/internal/handler/subscription_test.go`
- Modify: `backend/internal/psp/psp.go`

- [ ] **Step 1: Write the failing test for discount**

Add to `backend/internal/handler/subscription_test.go`:

```go
func TestSubscribe_AppliesReferralDiscount(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	mockPSP := &psp.MockPSP{}
	h := handler.New(queries, testCfg, mockPSP, &email.MockSender{})

	// Create owner, business, plan
	owner := testutil.SeedOwner(t, ctx, queries)
	biz := testutil.SeedBusiness(t, ctx, queries, owner.ID)
	plan := testutil.SeedPlan(t, ctx, queries, biz.ID, 1000) // R$10.00

	// Create referrer and referred subscribers
	referrer := testutil.SeedSubscriber(t, ctx, queries, "referrer@test.com")
	referred := testutil.SeedSubscriber(t, ctx, queries, "referred@test.com")

	// Create referral code for referrer
	_, err := queries.CreateReferral(ctx, repository.CreateReferralParams{
		ReferrerID: referrer.ID,
		ReferredID: referrer.ID, // template
		BusinessID: biz.ID,
		Code:       "TESTCODE",
	})
	require.NoError(t, err)

	// Apply referral for referred
	_, err = queries.CreateReferral(ctx, repository.CreateReferralParams{
		ReferrerID: referrer.ID,
		ReferredID: referred.ID,
		BusinessID: biz.ID,
		Code:       "TESTCODE",
	})
	require.NoError(t, err)

	// Subscribe as referred user
	body := `{"plan_id":` + fmt.Sprintf("%d", plan.ID) + `}`
	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx = context.WithValue(req.Context(), middleware.UserIDContextKey(), referred.ID)
	ctx = context.WithValue(ctx, middleware.RoleContextKey(), "subscriber")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.Subscribe(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify PSP was called with discounted price (900 = 10% off 1000)
	assert.Equal(t, int64(900), mockPSP.LastSubscriptionRequest.PriceCents)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/handler/ -run TestSubscribe_AppliesReferralDiscount -v`
Expected: FAIL — discount not applied (PSP receives 1000 not 900)

- [ ] **Step 3: Add LastSubscriptionRequest to MockPSP**

Modify `backend/internal/psp/mock.go` — add field:

```go
type MockPSP struct {
	LastSubscriptionRequest CreateSubscriptionRequest
}

func (m *MockPSP) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	m.LastSubscriptionRequest = req
	return &Subscription{ID: "mock_sub_" + req.CustomerID, Status: "ACTIVE"}, nil
}
```

- [ ] **Step 4: Implement discount logic in Subscribe handler**

Modify `backend/internal/handler/subscription.go` — after getting subscriber data and before creating PSP customer, add discount check:

```go
// Check for active referral discount
discountPercent := int32(0)
_, refErr := h.Queries.GetReferralByReferredAndBusiness(r.Context(), repository.GetReferralByReferredAndBusinessParams{
	ReferredID: subscriberID,
	BusinessID: plan.BusinessID,
})
if refErr == nil {
	discountPercent = 10
}

// Calculate discounted price
priceCents := plan.PriceCents
if discountPercent > 0 {
	priceCents = priceCents * int64(100-discountPercent) / 100
}
```

Then use `priceCents` (not `plan.PriceCents`) in `CreateSubscription` call:

```go
pspSub, err := h.PSP.CreateSubscription(r.Context(), psp.CreateSubscriptionRequest{
	CustomerID:  customer.ID,
	PriceCents:  priceCents,
	Description: plan.Name,
	Cycle:       "MONTHLY",
})
```

And use `CreateSubscriptionWithDiscount` for the DB insert:

```go
sub, err := h.Queries.CreateSubscriptionWithDiscount(r.Context(), repository.CreateSubscriptionWithDiscountParams{
	PlanID:            plan.ID,
	SubscriberID:      subscriberID,
	BusinessID:        plan.BusinessID,
	PspSubscriptionID: pgText(pspSub.ID),
	Status:            "active",
	PeriodEnd:         pgTimestamptz(periodEnd),
	DiscountPercent:   discountPercent,
})
```

- [ ] **Step 5: Run test**

Run: `cd backend && go test ./internal/handler/ -run TestSubscribe_AppliesReferralDiscount -v`
Expected: PASS

- [ ] **Step 6: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/subscription.go backend/internal/handler/subscription_test.go backend/internal/psp/mock.go
git commit -m "feat: aplica desconto de 10% do referral no subscribe (PSP + DB)"
```

---

### Task 9: Email Templates HTML

**Files:**
- Create: `backend/internal/email/templates.go`
- Create: `backend/internal/email/templates_test.go`

- [ ] **Step 1: Write the failing test**

```go
// backend/internal/email/templates_test.go
package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWelcomeEmail(t *testing.T) {
	subject, body := WelcomeEmail("João", "Café Premium", "Padaria do Bairro")
	assert.Equal(t, "Bem-vindo ao ClubePay!", subject)
	assert.Contains(t, body, "João")
	assert.Contains(t, body, "Café Premium")
	assert.Contains(t, body, "Padaria do Bairro")
	assert.Contains(t, body, "<html")
}

func TestPaymentConfirmedEmail(t *testing.T) {
	subject, body := PaymentConfirmedEmail("João", "Café Premium", "R$ 29,90")
	assert.Equal(t, "Pagamento confirmado - ClubePay", subject)
	assert.Contains(t, body, "João")
	assert.Contains(t, body, "R$ 29,90")
}

func TestSubscriptionCancelledEmail(t *testing.T) {
	subject, body := SubscriptionCancelledEmail("João", "Café Premium", "30/04/2026")
	assert.Equal(t, "Assinatura cancelada - ClubePay", subject)
	assert.Contains(t, body, "João")
	assert.Contains(t, body, "30/04/2026")
}

func TestPasswordResetEmail(t *testing.T) {
	subject, body := PasswordResetEmail("João", "https://clubepay.com/resetar-senha?token=abc123")
	assert.Equal(t, "Redefinir senha - ClubePay", subject)
	assert.Contains(t, body, "João")
	assert.Contains(t, body, "https://clubepay.com/resetar-senha?token=abc123")
}

func TestGraceBlockedEmail(t *testing.T) {
	subject, body := GraceBlockedEmail("João")
	assert.Equal(t, "ClubePay - Assinatura bloqueada", subject)
	assert.Contains(t, body, "João")
	assert.Contains(t, body, "bloqueada")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/email/ -run TestWelcomeEmail -v`
Expected: FAIL — function not defined

- [ ] **Step 3: Implement email templates**

```go
// backend/internal/email/templates.go
package email

import "fmt"

const htmlWrapper = `<html><body style="font-family:system-ui,sans-serif;color:#333;max-width:600px;margin:0 auto;padding:20px">
<div style="background:#2a7d6e;color:white;padding:20px;text-align:center;border-radius:8px 8px 0 0">
<h1 style="margin:0;font-size:24px">ClubePay</h1>
</div>
<div style="padding:20px;border:1px solid #e5e7eb;border-top:none;border-radius:0 0 8px 8px">
%s
</div>
<p style="color:#9ca3af;font-size:12px;text-align:center;margin-top:16px">ClubePay — Clube de assinatura para seu negocio</p>
</body></html>`

func WelcomeEmail(name, planName, businessName string) (string, string) {
	subject := "Bem-vindo ao ClubePay!"
	content := fmt.Sprintf(`<h2>Ola %s!</h2>
<p>Sua assinatura do plano <strong>%s</strong> no <strong>%s</strong> foi criada com sucesso.</p>
<p>Agora voce pode aproveitar todos os beneficios do seu plano.</p>
<p>Obrigado por assinar!</p>`, name, planName, businessName)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func PaymentConfirmedEmail(name, planName, amount string) (string, string) {
	subject := "Pagamento confirmado - ClubePay"
	content := fmt.Sprintf(`<h2>Pagamento confirmado!</h2>
<p>Ola %s, seu pagamento de <strong>%s</strong> para o plano <strong>%s</strong> foi confirmado.</p>
<p>Sua assinatura continua ativa. Aproveite!</p>`, name, amount, planName)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func SubscriptionCancelledEmail(name, planName, validUntil string) (string, string) {
	subject := "Assinatura cancelada - ClubePay"
	content := fmt.Sprintf(`<h2>Assinatura cancelada</h2>
<p>Ola %s, sua assinatura do plano <strong>%s</strong> foi cancelada.</p>
<p>Voce ainda pode usar o servico ate <strong>%s</strong>.</p>
<p>Sentiremos sua falta!</p>`, name, planName, validUntil)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func PasswordResetEmail(name, resetURL string) (string, string) {
	subject := "Redefinir senha - ClubePay"
	content := fmt.Sprintf(`<h2>Redefinir senha</h2>
<p>Ola %s, recebemos um pedido para redefinir sua senha.</p>
<p><a href="%s" style="display:inline-block;background:#2a7d6e;color:white;padding:12px 24px;text-decoration:none;border-radius:6px">Redefinir minha senha</a></p>
<p style="color:#6b7280;font-size:14px">Este link expira em 1 hora. Se voce nao solicitou, ignore este email.</p>`, name, resetURL)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func GraceBlockedEmail(name string) (string, string) {
	subject := "ClubePay - Assinatura bloqueada"
	content := fmt.Sprintf(`<h2>Assinatura bloqueada</h2>
<p>Ola %s, sua assinatura foi bloqueada por falta de pagamento.</p>
<p>Por favor, regularize seu pagamento para continuar usando o servico.</p>
<p>Equipe ClubePay</p>`, name)
	return subject, fmt.Sprintf(htmlWrapper, content)
}
```

- [ ] **Step 4: Run all email tests**

Run: `cd backend && go test ./internal/email/ -v`
Expected: All PASS

- [ ] **Step 5: Update cron.go to use new template**

Modify `backend/internal/handler/cron.go` — replace inline email with:

```go
subject, body := email.GraceBlockedEmail(subscriber.Name)
h.Email.Send(subscriber.Email, subject, body)
```

Add import: `"github.com/clubepay/backend/internal/email"`

- [ ] **Step 6: Commit**

```bash
git add backend/internal/email/templates.go backend/internal/email/templates_test.go backend/internal/handler/cron.go
git commit -m "feat: templates HTML de email (welcome, payment, cancel, reset, blocked)"
```

---

### Task 10: Password Reset — Backend

**Files:**
- Modify: `backend/internal/handler/auth.go`
- Modify: `backend/internal/handler/auth_test.go`
- Modify: `backend/cmd/api/router.go`

- [ ] **Step 1: Write the failing tests**

Add to `backend/internal/handler/auth_test.go`:

```go
func TestRequestPasswordReset_Success(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	mockEmail := &email.MockSender{}
	h := handler.New(queries, testCfg, &psp.MockPSP{}, mockEmail)

	testutil.SeedSubscriber(t, ctx, queries, "reset@test.com")

	body := `{"email":"reset@test.com"}`
	req := httptest.NewRequest("POST", "/api/auth/request-password-reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RequestPasswordReset(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, mockEmail.Sent, 1)
	assert.Equal(t, "Redefinir senha - ClubePay", mockEmail.Sent[0].Subject)
}

func TestRequestPasswordReset_UnknownEmail_StillReturns200(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	h := handler.New(queries, testCfg, &psp.MockPSP{}, &email.MockSender{})

	body := `{"email":"nobody@test.com"}`
	req := httptest.NewRequest("POST", "/api/auth/request-password-reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RequestPasswordReset(rec, req)

	// Always return 200 to not leak user existence
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestConfirmPasswordReset_Success(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	h := handler.New(queries, testCfg, &psp.MockPSP{}, &email.MockSender{})

	user := testutil.SeedSubscriber(t, ctx, queries, "reset2@test.com")

	// Create reset token
	reset, err := queries.CreatePasswordReset(ctx, repository.CreatePasswordResetParams{
		UserID:    user.ID,
		Token:     "valid-reset-token-123",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
	})
	require.NoError(t, err)
	_ = reset

	body := `{"token":"valid-reset-token-123","password":"newsecurepassword"}`
	req := httptest.NewRequest("POST", "/api/auth/confirm-password-reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ConfirmPasswordReset(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify new password works
	updatedUser, _ := queries.GetUserByEmail(ctx, "reset2@test.com")
	assert.True(t, domain.CheckPassword("newsecurepassword", updatedUser.PasswordHash))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/handler/ -run TestRequestPasswordReset -v`
Expected: FAIL — `RequestPasswordReset` not defined

- [ ] **Step 3: Implement password reset handlers**

Add to `backend/internal/handler/auth.go`:

```go
// RequestPasswordReset sends a password reset email.
// POST /api/auth/request-password-reset
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	// Always return 200 to not leak user existence
	defer writeJSON(w, http.StatusOK, map[string]string{"message": "se o email existir, voce recebera um link para redefinir sua senha"})

	user, err := h.Queries.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		return
	}

	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		slog.Error("failed to generate reset token", "error", err)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	_, err = h.Queries.CreatePasswordReset(r.Context(), repository.CreatePasswordResetParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: pgTimestamptz(time.Now().Add(time.Hour)),
	})
	if err != nil {
		slog.Error("failed to create password reset", "error", err)
		return
	}

	resetURL := h.Config.FrontendURL + "/resetar-senha?token=" + token
	subject, body := email.PasswordResetEmail(user.Name, resetURL)
	if err := h.Email.Send(user.Email, subject, body); err != nil {
		slog.Error("failed to send password reset email", "error", err)
	}
}

// ConfirmPasswordReset resets the user's password using a valid token.
// POST /api/auth/confirm-password-reset
func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if len(input.Password) < 8 {
		writeError(w, http.StatusBadRequest, "senha deve ter no minimo 8 caracteres")
		return
	}

	reset, err := h.Queries.GetPasswordResetByToken(r.Context(), input.Token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "token invalido ou expirado")
		return
	}

	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}

	if err := h.Queries.UpdateUserPassword(r.Context(), repository.UpdateUserPasswordParams{
		ID:           reset.UserID,
		PasswordHash: hash,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar senha")
		return
	}

	if err := h.Queries.MarkPasswordResetUsed(r.Context(), reset.ID); err != nil {
		slog.Error("failed to mark reset as used", "error", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "senha atualizada com sucesso"})
}
```

Add imports: `"crypto/rand"`, `"encoding/hex"`, `"log/slog"`, `"github.com/clubepay/backend/internal/email"`

- [ ] **Step 4: Add routes to router**

Add to `backend/cmd/api/router.go` inside the rate-limited auth group:

```go
r.Post("/api/auth/request-password-reset", h.RequestPasswordReset)
r.Post("/api/auth/confirm-password-reset", h.ConfirmPasswordReset)
```

- [ ] **Step 5: Run tests**

Run: `cd backend && go test ./internal/handler/ -run "TestRequestPasswordReset|TestConfirmPasswordReset" -v`
Expected: All PASS

- [ ] **Step 6: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/auth.go backend/internal/handler/auth_test.go backend/cmd/api/router.go
git commit -m "feat: endpoints de password reset (request + confirm com token)"
```

---

### Task 11: Profile/Account Update — Backend

**Files:**
- Create: `backend/internal/handler/profile.go`
- Create: `backend/internal/handler/profile_test.go`
- Modify: `backend/cmd/api/router.go`

- [ ] **Step 1: Write the failing test**

```go
// backend/internal/handler/profile_test.go
package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/testutil"
)

func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	h := New(queries, testCfg, &psp.MockPSP{}, &email.MockSender{})

	user := testutil.SeedSubscriber(t, ctx, queries, "profile@test.com")

	req := httptest.NewRequest("GET", "/api/profile", nil)
	ctx = context.WithValue(req.Context(), middleware.UserIDContextKey(), user.ID)
	ctx = context.WithValue(ctx, middleware.RoleContextKey(), "subscriber")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.GetProfile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "profile@test.com")
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	h := New(queries, testCfg, &psp.MockPSP{}, &email.MockSender{})

	user := testutil.SeedSubscriber(t, ctx, queries, "update@test.com")

	body := `{"name":"Novo Nome","phone":"11999998888"}`
	req := httptest.NewRequest("PUT", "/api/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx = context.WithValue(req.Context(), middleware.UserIDContextKey(), user.ID)
	ctx = context.WithValue(ctx, middleware.RoleContextKey(), "subscriber")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateProfile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Novo Nome")
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	db := testutil.SetupTestDB(t, ctx)
	queries := repository.New(db)
	h := New(queries, testCfg, &psp.MockPSP{}, &email.MockSender{})

	user := testutil.SeedSubscriber(t, ctx, queries, "changepw@test.com")

	body := `{"current_password":"password123","new_password":"novaSenha123"}`
	req := httptest.NewRequest("POST", "/api/profile/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx = context.WithValue(req.Context(), middleware.UserIDContextKey(), user.ID)
	ctx = context.WithValue(ctx, middleware.RoleContextKey(), "subscriber")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.ChangePassword(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/handler/ -run "TestGetProfile|TestUpdateProfile|TestChangePassword" -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement profile handlers**

```go
// backend/internal/handler/profile.go
package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// GetProfile returns the authenticated user's profile.
// GET /api/profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "usuario nao encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar perfil")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone.String,
			"role":  user.Role,
		},
	})
}

// UpdateProfile updates name and phone.
// PUT /api/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var input struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "nome e obrigatorio")
		return
	}

	user, err := h.Queries.UpdateUserProfile(r.Context(), repository.UpdateUserProfileParams{
		ID:    userID,
		Name:  input.Name,
		Phone: pgText(input.Phone),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar perfil")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone.String,
			"role":  user.Role,
		},
	})
}

// ChangePassword changes the user's password.
// POST /api/profile/change-password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if len(input.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "nova senha deve ter no minimo 8 caracteres")
		return
	}

	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar usuario")
		return
	}

	if !domain.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "senha atual incorreta")
		return
	}

	hash, err := domain.HashPassword(input.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}

	if err := h.Queries.UpdateUserPassword(r.Context(), repository.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hash,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar senha")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "senha atualizada com sucesso"})
}
```

- [ ] **Step 4: Add routes to router**

Add to both owner and subscriber groups in `backend/cmd/api/router.go`:

```go
// Shared auth routes (any authenticated user)
r.Group(func(r chi.Router) {
	r.Use(middleware.Auth(cfg.JWTSecret))
	r.Get("/api/profile", h.GetProfile)
	r.Put("/api/profile", h.UpdateProfile)
	r.Post("/api/profile/change-password", h.ChangePassword)
})
```

- [ ] **Step 5: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/profile.go backend/internal/handler/profile_test.go backend/cmd/api/router.go
git commit -m "feat: endpoints de perfil (GET/PUT profile, change password)"
```

---

### Task 12: Enviar Emails Transacionais no Subscribe/Cancel

**Files:**
- Modify: `backend/internal/handler/subscription.go`
- Modify: `backend/internal/handler/subscriber.go`
- Modify: `backend/internal/handler/webhook.go`

- [ ] **Step 1: Add welcome email to Subscribe handler**

In `backend/internal/handler/subscription.go`, after the successful DB insert (`writeJSON` line), add before `writeJSON`:

```go
// Send welcome email
go func() {
	plan, _ := h.Queries.GetPlanByID(context.Background(), input.PlanID)
	biz, _ := h.Queries.GetBusinessByPlanID(context.Background(), plan.BusinessID)
	subject, body := email.WelcomeEmail(subscriber.Name, plan.Name, biz.Name)
	h.Email.Send(subscriber.Email, subject, body)
}()
```

Add imports: `"github.com/clubepay/backend/internal/email"`

Note: This requires a `GetBusinessByPlanID` or use the business already fetched. Since we have `plan.BusinessID`, we can look up the business. Add to `backend/queries/businesses.sql`:

```sql
-- name: GetBusinessByID :one
SELECT * FROM businesses WHERE id = $1;
```

Then regenerate sqlc: `cd backend && sqlc generate`

Actually, check if `GetBusinessByID` already exists. If not, add it. Then use:

```go
go func() {
	biz, _ := h.Queries.GetBusinessByID(context.Background(), plan.BusinessID)
	subject, body := email.WelcomeEmail(subscriber.Name, plan.Name, biz.Name)
	h.Email.Send(subscriber.Email, subject, body)
}()
```

- [ ] **Step 2: Add cancellation email to CancelBySubscriber**

In `backend/internal/handler/subscriber.go`, after successful cancel, add:

```go
go func() {
	plan, _ := h.Queries.GetPlanByID(context.Background(), sub.PlanID)
	validUntil := sub.PeriodEnd.Time.Format("02/01/2006")
	subject, body := email.SubscriptionCancelledEmail(user.Name, plan.Name, validUntil)
	h.Email.Send(user.Email, subject, body)
}()
```

- [ ] **Step 3: Add payment confirmed email to webhook**

In `backend/internal/handler/webhook.go`, inside the `PAYMENT_CONFIRMED`/`PAYMENT_RECEIVED` case, after status update:

```go
go func() {
	subscriber, _ := h.Queries.GetUserByID(context.Background(), sub.SubscriberID)
	plan, _ := h.Queries.GetPlanByID(context.Background(), sub.PlanID)
	amount := fmt.Sprintf("R$ %.2f", float64(plan.PriceCents)/100)
	subject, body := email.PaymentConfirmedEmail(subscriber.Name, plan.Name, amount)
	h.Email.Send(subscriber.Email, subject, body)
}()
```

- [ ] **Step 4: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/subscription.go backend/internal/handler/subscriber.go backend/internal/handler/webhook.go backend/queries/businesses.sql backend/internal/repository/
git commit -m "feat: emails transacionais (welcome, payment confirmed, cancellation)"
```

---

## FASE 3: FRONTEND UX (75% → 100%)

---

### Task 13: Error Boundary + 404 Page

**Files:**
- Create: `frontend/src/app/error.tsx`
- Create: `frontend/src/app/not-found.tsx`

- [ ] **Step 1: Create global error boundary**

```tsx
// frontend/src/app/error.tsx
"use client";

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <h1 className="text-6xl font-bold text-gray-300 mb-4">Ops!</h1>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">
          Algo deu errado
        </h2>
        <p className="text-gray-500 mb-6">
          Ocorreu um erro inesperado. Tente novamente.
        </p>
        <button
          onClick={reset}
          className="px-6 py-3 rounded-lg text-white font-semibold"
          style={{ backgroundColor: "#2a7d6e" }}
        >
          Tentar novamente
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create 404 page**

```tsx
// frontend/src/app/not-found.tsx
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <h1 className="text-8xl font-bold text-gray-200 mb-4">404</h1>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">
          Pagina nao encontrada
        </h2>
        <p className="text-gray-500 mb-6">
          A pagina que voce procura nao existe ou foi movida.
        </p>
        <Link
          href="/"
          className="inline-block px-6 py-3 rounded-lg text-white font-semibold"
          style={{ backgroundColor: "#2a7d6e" }}
        >
          Voltar ao inicio
        </Link>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/error.tsx frontend/src/app/not-found.tsx
git commit -m "feat: error boundary global + pagina 404"
```

---

### Task 14: Next.js Middleware — Route Protection

**Files:**
- Create: `frontend/src/middleware.ts`

- [ ] **Step 1: Create middleware**

```typescript
// frontend/src/middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const protectedPaths = ["/dashboard", "/planos", "/perfil", "/meu-plano"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Check if this is a protected route
  const isProtected = protectedPaths.some(
    (path) => pathname === path || pathname.startsWith(path + "/")
  );

  if (!isProtected) {
    return NextResponse.next();
  }

  // Token check happens client-side (localStorage not available in middleware)
  // Middleware adds security headers and handles redirects for known patterns
  const response = NextResponse.next();

  // Add cache-control for protected routes (no caching)
  response.headers.set(
    "Cache-Control",
    "no-store, no-cache, must-revalidate"
  );

  return response;
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/planos/:path*",
    "/perfil/:path*",
    "/meu-plano/:path*",
  ],
};
```

- [ ] **Step 2: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add frontend/src/middleware.ts
git commit -m "feat: Next.js middleware para rotas protegidas (cache-control)"
```

---

### Task 15: Loading Skeleton

**Files:**
- Create: `frontend/src/app/(auth)/loading.tsx`

- [ ] **Step 1: Create loading skeleton**

```tsx
// frontend/src/app/(auth)/loading.tsx
export default function AuthLoading() {
  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto animate-pulse">
        {/* Header skeleton */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <div className="h-8 w-48 bg-gray-200 rounded mb-2" />
            <div className="h-4 w-32 bg-gray-200 rounded" />
          </div>
          <div className="h-10 w-20 bg-gray-200 rounded" />
        </div>

        {/* Stats grid skeleton */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-white p-6 rounded-xl shadow-sm">
              <div className="h-4 w-24 bg-gray-200 rounded mb-2" />
              <div className="h-8 w-16 bg-gray-200 rounded" />
            </div>
          ))}
        </div>

        {/* Content skeleton */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <div className="h-6 w-40 bg-gray-200 rounded mb-4" />
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="flex justify-between py-3 border-b border-gray-100">
              <div className="h-4 w-48 bg-gray-200 rounded" />
              <div className="h-4 w-24 bg-gray-200 rounded" />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/(auth)/loading.tsx
git commit -m "feat: loading skeleton para rotas autenticadas"
```

---

### Task 16: Password Reset — Frontend Pages

**Files:**
- Create: `frontend/src/app/(auth)/esqueci-senha/page.tsx`
- Create: `frontend/src/app/(auth)/resetar-senha/page.tsx`

- [ ] **Step 1: Create forgot password page**

```tsx
// frontend/src/app/(auth)/esqueci-senha/page.tsx
"use client";

import { useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await api.post("/api/auth/request-password-reset", { email });
      setSent(true);
    } catch {
      setError("Erro ao enviar email. Tente novamente.");
    } finally {
      setLoading(false);
    }
  }

  if (sent) {
    return (
      <div className="min-h-screen flex items-center justify-center px-4 bg-gray-50">
        <div className="max-w-sm w-full bg-white rounded-xl shadow-sm p-8 text-center">
          <div className="text-4xl mb-4">📧</div>
          <h1 className="text-xl font-bold text-gray-900 mb-2">Email enviado!</h1>
          <p className="text-gray-600 mb-6">
            Se o email existir em nossa base, voce recebera um link para redefinir sua senha.
          </p>
          <Link href="/login" className="text-sm font-semibold" style={{ color: "#2a7d6e" }}>
            Voltar ao login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4 bg-gray-50">
      <div className="max-w-sm w-full bg-white rounded-xl shadow-sm p-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Esqueci minha senha</h1>
        <p className="text-gray-500 mb-6 text-sm">
          Informe seu email e enviaremos um link para redefinir sua senha.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              Email
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
              placeholder="seu@email.com"
            />
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 rounded-lg text-white font-semibold disabled:opacity-60"
            style={{ backgroundColor: "#2a7d6e" }}
          >
            {loading ? "Enviando..." : "Enviar link"}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-gray-500">
          <Link href="/login" className="font-semibold" style={{ color: "#2a7d6e" }}>
            Voltar ao login
          </Link>
        </p>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create reset password page**

```tsx
// frontend/src/app/(auth)/resetar-senha/page.tsx
"use client";

import { useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";

function ResetForm() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (password.length < 8) {
      setError("Senha deve ter no minimo 8 caracteres.");
      return;
    }
    if (password !== confirm) {
      setError("As senhas nao coincidem.");
      return;
    }

    setLoading(true);
    try {
      await api.post("/api/auth/confirm-password-reset", { token, password });
      setSuccess(true);
    } catch {
      setError("Token invalido ou expirado. Solicite um novo link.");
    } finally {
      setLoading(false);
    }
  }

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center px-4 bg-gray-50">
        <div className="max-w-sm w-full bg-white rounded-xl shadow-sm p-8 text-center">
          <h1 className="text-xl font-bold text-gray-900 mb-2">Link invalido</h1>
          <p className="text-gray-600 mb-4">Este link nao contem um token valido.</p>
          <Link href="/esqueci-senha" className="text-sm font-semibold" style={{ color: "#2a7d6e" }}>
            Solicitar novo link
          </Link>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center px-4 bg-gray-50">
        <div className="max-w-sm w-full bg-white rounded-xl shadow-sm p-8 text-center">
          <div className="text-4xl mb-4">✅</div>
          <h1 className="text-xl font-bold text-gray-900 mb-2">Senha atualizada!</h1>
          <p className="text-gray-600 mb-4">Sua senha foi redefinida com sucesso.</p>
          <Link
            href="/login"
            className="inline-block px-6 py-3 rounded-lg text-white font-semibold"
            style={{ backgroundColor: "#2a7d6e" }}
          >
            Fazer login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4 bg-gray-50">
      <div className="max-w-sm w-full bg-white rounded-xl shadow-sm p-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Nova senha</h1>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Nova senha
            </label>
            <input
              id="password"
              type="password"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
            />
          </div>

          <div>
            <label htmlFor="confirm" className="block text-sm font-medium text-gray-700 mb-1">
              Confirmar senha
            </label>
            <input
              id="confirm"
              type="password"
              required
              minLength={8}
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
            />
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 rounded-lg text-white font-semibold disabled:opacity-60"
            style={{ backgroundColor: "#2a7d6e" }}
          >
            {loading ? "Salvando..." : "Redefinir senha"}
          </button>
        </form>
      </div>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Carregando...</div>}>
      <ResetForm />
    </Suspense>
  );
}
```

- [ ] **Step 3: Add "Esqueci minha senha" link to login page**

Modify `frontend/src/app/(auth)/login/page.tsx` — add after the form, before the register link:

```tsx
<p className="text-center text-sm text-gray-500">
  <Link href="/esqueci-senha" className="font-semibold" style={{ color: "#2a7d6e" }}>
    Esqueci minha senha
  </Link>
</p>
```

- [ ] **Step 4: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/(auth)/esqueci-senha/ frontend/src/app/(auth)/resetar-senha/ frontend/src/app/(auth)/login/page.tsx
git commit -m "feat: paginas de esqueci senha e resetar senha no frontend"
```

---

### Task 17: Profile Page — Frontend

**Files:**
- Create: `frontend/src/app/(auth)/perfil/page.tsx`

- [ ] **Step 1: Create profile page**

```tsx
// frontend/src/app/(auth)/perfil/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getToken, clearToken } from "@/lib/auth";
import { api, ApiError } from "@/lib/api";

interface UserProfile {
  id: number;
  email: string;
  name: string;
  phone: string;
  role: string;
}

export default function ProfilePage() {
  const router = useRouter();
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);

  // Profile form
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [profileSaving, setProfileSaving] = useState(false);
  const [profileMsg, setProfileMsg] = useState("");

  // Password form
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [pwSaving, setPwSaving] = useState(false);
  const [pwMsg, setPwMsg] = useState("");
  const [pwError, setPwError] = useState("");

  useEffect(() => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    api
      .get<{ user: UserProfile }>("/api/profile", token)
      .then((res) => {
        setUser(res.user);
        setName(res.user.name);
        setPhone(res.user.phone || "");
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 401) {
          clearToken();
          router.push("/login");
        }
      })
      .finally(() => setLoading(false));
  }, [router]);

  async function handleProfileSave(e: React.FormEvent) {
    e.preventDefault();
    setProfileSaving(true);
    setProfileMsg("");

    try {
      const token = getToken()!;
      const res = await api.put<{ user: UserProfile }>("/api/profile", { name, phone }, token);
      setUser(res.user);
      setProfileMsg("Perfil atualizado!");
    } catch {
      setProfileMsg("Erro ao salvar perfil.");
    } finally {
      setProfileSaving(false);
    }
  }

  async function handlePasswordChange(e: React.FormEvent) {
    e.preventDefault();
    setPwError("");
    setPwMsg("");

    if (newPassword.length < 8) {
      setPwError("Nova senha deve ter no minimo 8 caracteres.");
      return;
    }
    if (newPassword !== confirmPassword) {
      setPwError("As senhas nao coincidem.");
      return;
    }

    setPwSaving(true);
    try {
      const token = getToken()!;
      await api.post("/api/profile/change-password", {
        current_password: currentPassword,
        new_password: newPassword,
      }, token);
      setPwMsg("Senha atualizada!");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setPwError("Senha atual incorreta.");
      } else {
        setPwError("Erro ao alterar senha.");
      }
    } finally {
      setPwSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-lg mx-auto space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">Meu perfil</h1>
          <button
            onClick={() => router.back()}
            className="text-sm font-semibold"
            style={{ color: "#2a7d6e" }}
          >
            Voltar
          </button>
        </div>

        {/* Profile info */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Dados pessoais</h2>
          <p className="text-sm text-gray-500 mb-4">Email: {user?.email}</p>

          <form onSubmit={handleProfileSave} className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                Nome
              </label>
              <input
                id="name"
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
              />
            </div>

            <div>
              <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">
                Telefone
              </label>
              <input
                id="phone"
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
                placeholder="(11) 99999-9999"
              />
            </div>

            {profileMsg && (
              <p className={`text-sm ${profileMsg.includes("Erro") ? "text-red-600" : "text-green-600"}`}>
                {profileMsg}
              </p>
            )}

            <button
              type="submit"
              disabled={profileSaving}
              className="w-full py-3 rounded-lg text-white font-semibold disabled:opacity-60"
              style={{ backgroundColor: "#2a7d6e" }}
            >
              {profileSaving ? "Salvando..." : "Salvar"}
            </button>
          </form>
        </div>

        {/* Change password */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Alterar senha</h2>

          <form onSubmit={handlePasswordChange} className="space-y-4">
            <div>
              <label htmlFor="currentPw" className="block text-sm font-medium text-gray-700 mb-1">
                Senha atual
              </label>
              <input
                id="currentPw"
                type="password"
                required
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
              />
            </div>

            <div>
              <label htmlFor="newPw" className="block text-sm font-medium text-gray-700 mb-1">
                Nova senha
              </label>
              <input
                id="newPw"
                type="password"
                required
                minLength={8}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
              />
            </div>

            <div>
              <label htmlFor="confirmPw" className="block text-sm font-medium text-gray-700 mb-1">
                Confirmar nova senha
              </label>
              <input
                id="confirmPw"
                type="password"
                required
                minLength={8}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
              />
            </div>

            {pwError && <p className="text-sm text-red-600">{pwError}</p>}
            {pwMsg && <p className="text-sm text-green-600">{pwMsg}</p>}

            <button
              type="submit"
              disabled={pwSaving}
              className="w-full py-3 rounded-lg text-white font-semibold disabled:opacity-60"
              style={{ backgroundColor: "#d4a853" }}
            >
              {pwSaving ? "Salvando..." : "Alterar senha"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Add profile link to dashboard**

Modify `frontend/src/app/(auth)/dashboard/page.tsx` — add a "Meu perfil" link next to the Logout button:

```tsx
<Link href="/perfil" className="text-sm font-semibold" style={{ color: "#2a7d6e" }}>
  Meu perfil
</Link>
```

- [ ] **Step 3: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/(auth)/perfil/ frontend/src/app/(auth)/dashboard/page.tsx
git commit -m "feat: pagina de perfil (editar dados + alterar senha)"
```

---

### Task 18: API Client — Retry Logic

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/lib/__tests__/api.test.ts`

- [ ] **Step 1: Add retry test**

Add to `frontend/src/lib/__tests__/api.test.ts`:

```typescript
describe("retry logic", () => {
  it("retries on 500 error and succeeds on second attempt", async () => {
    let attempts = 0;
    global.fetch = vi.fn(() => {
      attempts++;
      if (attempts === 1) {
        return Promise.resolve({
          ok: false,
          status: 500,
          statusText: "Internal Server Error",
          json: () => Promise.resolve({ message: "server error" }),
        } as Response);
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ data: "ok" }),
      } as Response);
    });

    const result = await api.get<{ data: string }>("/test");
    expect(result.data).toBe("ok");
    expect(attempts).toBe(2);
  });

  it("does not retry on 4xx errors", async () => {
    let attempts = 0;
    global.fetch = vi.fn(() => {
      attempts++;
      return Promise.resolve({
        ok: false,
        status: 400,
        statusText: "Bad Request",
        json: () => Promise.resolve({ message: "bad request" }),
      } as Response);
    });

    await expect(api.get("/test")).rejects.toThrow();
    expect(attempts).toBe(1);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run --reporter=verbose src/lib/__tests__/api.test.ts`
Expected: FAIL — no retry logic

- [ ] **Step 3: Implement retry**

Replace `frontend/src/lib/api.ts`:

```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface RequestOptions {
  method?: string;
  body?: unknown;
  token?: string;
}

const MAX_RETRIES = 2;
const RETRY_DELAY_MS = 500;

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = "GET", body, token } = opts;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let lastError: ApiError | null = null;

  for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
    if (attempt > 0) {
      await new Promise((r) => setTimeout(r, RETRY_DELAY_MS * attempt));
    }

    const res = await fetch(`${API_URL}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (res.ok) {
      return res.json();
    }

    const error = await res.json().catch(() => ({ message: res.statusText }));
    lastError = new ApiError(res.status, error.message || res.statusText);

    // Only retry on 5xx server errors
    if (res.status < 500) {
      throw lastError;
    }
  }

  throw lastError!;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export const api = {
  get: <T>(path: string, token?: string) =>
    request<T>(path, { token }),

  post: <T>(path: string, body: unknown, token?: string) =>
    request<T>(path, { method: "POST", body, token }),

  put: <T>(path: string, body: unknown, token?: string) =>
    request<T>(path, { method: "PUT", body, token }),

  del: <T>(path: string, token?: string) =>
    request<T>(path, { method: "DELETE", token }),
};
```

- [ ] **Step 4: Run tests**

Run: `cd frontend && npx vitest run --reporter=verbose src/lib/__tests__/api.test.ts`
Expected: All PASS

- [ ] **Step 5: Run all frontend tests**

Run: `cd frontend && npx vitest run`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/api.ts frontend/src/lib/__tests__/api.test.ts
git commit -m "feat: retry logic no API client (ate 2 retries em 5xx)"
```

---

## FASE 4: INFRAESTRUTURA / DEVOPS (65% → 100%)

---

### Task 19: CI/CD Pipeline — GitHub Actions

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  test-backend:
    name: Backend Tests
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: clubepay
          POSTGRES_PASSWORD: clubepay
          POSTGRES_DB: clubepay_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
          cache-dependency-path: backend/go.sum

      - name: Run tests
        working-directory: backend
        env:
          DATABASE_URL: postgres://clubepay:clubepay@localhost:5432/clubepay_test?sslmode=disable
          JWT_SECRET: ci-test-secret
          CRON_SECRET: ci-cron-secret
        run: go test ./... -v -race -count=1

      - name: Build
        working-directory: backend
        run: go build -o /dev/null ./cmd/api

  test-frontend:
    name: Frontend Tests
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: "22"
          cache: npm
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run tests
        working-directory: frontend
        run: npx vitest run

      - name: Lint
        working-directory: frontend
        run: npm run lint

      - name: Build
        working-directory: frontend
        env:
          NEXT_PUBLIC_API_URL: http://localhost:8080
        run: npm run build

  docker-build:
    name: Docker Build Check
    runs-on: ubuntu-latest
    needs: [test-backend, test-frontend]

    steps:
      - uses: actions/checkout@v4

      - name: Build backend image
        run: docker build -f Dockerfile.backend -t clubepay-api:ci .

      - name: Build frontend image
        run: docker build -f Dockerfile.frontend -t clubepay-web:ci .
```

- [ ] **Step 2: Commit**

```bash
mkdir -p .github/workflows
git add .github/workflows/ci.yml
git commit -m "feat: CI pipeline com GitHub Actions (tests + build + docker)"
```

---

### Task 20: Docker Compose Producao

**Files:**
- Create: `docker-compose.prod.yml`
- Modify: `docker-compose.yml`

- [ ] **Step 1: Add healthcheck to postgres in docker-compose.yml**

Modify `docker-compose.yml` — add to postgres service:

```yaml
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U clubepay"]
      interval: 10s
      timeout: 5s
      retries: 5
```

And update backend `depends_on`:

```yaml
    depends_on:
      postgres:
        condition: service_healthy
```

- [ ] **Step 2: Create production compose**

```yaml
# docker-compose.prod.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
    restart: unless-stopped

  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    environment:
      DATABASE_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      PORT: "8080"
      ASAAS_API_KEY: ${ASAAS_API_KEY}
      ASAAS_URL: https://api.asaas.com/api/v3
      ASAAS_WEBHOOK_SECRET: ${ASAAS_WEBHOOK_SECRET}
      CRON_SECRET: ${CRON_SECRET}
      CORS_ORIGINS: ${CORS_ORIGINS}
      FRONTEND_URL: ${FRONTEND_URL}
      SMTP_HOST: ${SMTP_HOST}
      SMTP_PORT: ${SMTP_PORT:-587}
      SMTP_USERNAME: ${SMTP_USERNAME}
      SMTP_PASSWORD: ${SMTP_PASSWORD}
    depends_on:
      postgres:
        condition: service_healthy
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: "0.5"
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
    depends_on:
      - backend
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: "0.5"
    restart: unless-stopped

volumes:
  pgdata:
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml docker-compose.prod.yml
git commit -m "feat: docker-compose prod com healthchecks, resource limits e env vars"
```

---

### Task 21: .env.example + Documentacao

**Files:**
- Create: `.env.example`

- [ ] **Step 1: Create .env.example**

```bash
# ClubePay — Environment Variables
# Copy this file to .env and fill in the values

# ===== REQUIRED =====

# PostgreSQL connection string
DATABASE_URL=postgres://clubepay:clubepay@localhost:5432/clubepay_dev?sslmode=disable

# JWT signing secret (generate with: openssl rand -hex 32)
JWT_SECRET=

# ===== PAYMENT (Asaas) =====

# Asaas API key (leave empty for mock PSP in dev)
ASAAS_API_KEY=

# Asaas URL (sandbox for dev, production for prod)
# Dev: https://sandbox.asaas.com/api/v3
# Prod: https://api.asaas.com/api/v3
ASAAS_URL=https://sandbox.asaas.com/api/v3

# Asaas webhook HMAC secret
ASAAS_WEBHOOK_SECRET=

# ===== SECURITY =====

# CORS allowed origins (comma-separated, use * for dev)
# Prod example: https://clubepay.com,https://www.clubepay.com
CORS_ORIGINS=*

# Cron reconciliation secret
CRON_SECRET=

# ===== EMAIL (SMTP) =====

# SMTP config (leave empty for mock sender in dev)
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=

# ===== FRONTEND =====

# URL of the frontend (for password reset links, etc.)
FRONTEND_URL=http://localhost:3000

# URL of the backend API (used by Next.js client-side)
NEXT_PUBLIC_API_URL=http://localhost:8080

# ===== DOCKER PROD ONLY =====

# DB_USER=clubepay
# DB_PASSWORD=<strong-password>
# DB_NAME=clubepay_prod
```

- [ ] **Step 2: Add .env to .gitignore if not already**

Check `.gitignore` and ensure `.env` is listed (not `.env.example`).

- [ ] **Step 3: Commit**

```bash
git add .env.example
git commit -m "docs: .env.example com todas as variaveis de ambiente documentadas"
```

---

### Task 22: Auto-migrate on Backend Startup

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Add auto-migration logic**

Add to `backend/cmd/api/main.go` after `pool.Ping(ctx)` succeeds:

```go
// Run database migrations
if err := runMigrations(cfg.DatabaseURL); err != nil {
	slog.Error("failed to run migrations", "error", err)
	os.Exit(1)
}
slog.Info("database migrations applied")
```

Add the function at the end of main.go:

```go
func runMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
```

Add imports:

```go
"fmt"

"github.com/golang-migrate/migrate/v4"
_ "github.com/golang-migrate/migrate/v4/database/postgres"
_ "github.com/golang-migrate/migrate/v4/source/file"
```

- [ ] **Step 2: Add dependency**

Run: `cd backend && go get github.com/golang-migrate/migrate/v4 github.com/golang-migrate/migrate/v4/database/postgres github.com/golang-migrate/migrate/v4/source/file`

- [ ] **Step 3: Verify build**

Run: `cd backend && go build ./cmd/api`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/api/main.go backend/go.mod backend/go.sum
git commit -m "feat: auto-migrate no startup do backend"
```

---

### Task 23: Script de Backup PostgreSQL

**Files:**
- Create: `scripts/backup.sh`

- [ ] **Step 1: Create backup script**

```bash
#!/bin/bash
# scripts/backup.sh — Automated PostgreSQL backup for ClubePay
#
# Usage: ./scripts/backup.sh
# Cron:  0 3 * * * /path/to/clubepay/scripts/backup.sh
#
# Requires: pg_dump, gzip
# Env vars: DB_USER, DB_PASSWORD, DB_HOST, DB_NAME, BACKUP_DIR

set -euo pipefail

DB_USER="${DB_USER:-clubepay}"
DB_HOST="${DB_HOST:-localhost}"
DB_NAME="${DB_NAME:-clubepay_prod}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/clubepay}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting backup of ${DB_NAME}..."

PGPASSWORD="${DB_PASSWORD}" pg_dump \
  -h "$DB_HOST" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  --format=custom \
  --compress=9 \
  --file="$BACKUP_FILE"

echo "[$(date)] Backup saved to ${BACKUP_FILE}"
echo "[$(date)] Size: $(du -h "$BACKUP_FILE" | cut -f1)"

# Clean old backups
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +${RETENTION_DAYS} -delete
echo "[$(date)] Cleaned backups older than ${RETENTION_DAYS} days"

echo "[$(date)] Backup complete."
```

- [ ] **Step 2: Make executable**

Run: `chmod +x scripts/backup.sh`

- [ ] **Step 3: Commit**

```bash
git add scripts/backup.sh
git commit -m "feat: script de backup automatizado do PostgreSQL (retencao 30 dias)"
```

---

### Task 24: Logging Middleware — Request ID Propagation

**Files:**
- Modify: `backend/internal/middleware/logging.go`

- [ ] **Step 1: Read current logging middleware**

Read `backend/internal/middleware/logging.go` to understand current implementation.

- [ ] **Step 2: Update logging to include request ID**

Ensure the logging middleware includes `request_id` from chi's RequestID middleware:

```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			slog.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"request_id", chiMiddleware.GetReqID(r.Context()),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
```

- [ ] **Step 3: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/middleware/logging.go
git commit -m "fix: request_id no logging middleware para rastreabilidade"
```

---

### Task 25: Webhook Idempotency — Log Received Events

**Files:**
- Modify: `backend/internal/handler/webhook.go`

- [ ] **Step 1: Add structured logging for all webhook events**

Modify `backend/internal/handler/webhook.go` — add after payload parsing:

```go
slog.Info("webhook received",
	"event", payload.Event,
	"psp_subscription_id", payload.Payment.Subscription,
	"payment_status", payload.Payment.Status,
)
```

And add after each status change in the switch:

```go
slog.Info("webhook processed",
	"event", payload.Event,
	"subscription_id", sub.ID,
	"action", "reactivated", // or "grace_set", "cancelled"
)
```

- [ ] **Step 2: Run all tests**

Run: `cd backend && go test ./...`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/webhook.go
git commit -m "fix: logging estruturado em todos os eventos de webhook"
```

---

### Task 26: Update Makefile

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Update Makefile with new targets**

Append to `Makefile`:

```makefile
# Backup
backup:
	./scripts/backup.sh

# Lint backend
lint-backend:
	cd backend && go vet ./...

# Full lint
lint: lint-backend
	cd frontend && npm run lint

# Docker Compose Production
docker-prod-up:
	docker compose -f docker-compose.prod.yml up -d

docker-prod-down:
	docker compose -f docker-compose.prod.yml down

# Migrate
migrate-create:
	cd backend && migrate create -ext sql -dir migrations -seq $(name)
```

- [ ] **Step 2: Commit**

```bash
git add Makefile
git commit -m "chore: novos targets no Makefile (backup, lint, docker-prod, migrate-create)"
```

---

## FASE 5: TESTES FINAIS E VERIFICACAO

---

### Task 27: Rodar Todos os Testes

- [ ] **Step 1: Run backend tests**

Run: `cd backend && go test ./... -v -race`
Expected: All PASS

- [ ] **Step 2: Run frontend tests**

Run: `cd frontend && npx vitest run`
Expected: All PASS

- [ ] **Step 3: Build both Docker images**

Run: `docker build -f Dockerfile.backend -t clubepay-api:final . && docker build -f Dockerfile.frontend -t clubepay-web:final .`
Expected: Both build successfully

- [ ] **Step 4: Verify lint**

Run: `cd frontend && npm run lint && cd ../backend && go vet ./...`
Expected: No errors

---

## Resumo das Melhorias por Dimensao

### Seguranca (55% → 100%)
- [x] Task 1: Security headers middleware
- [x] Task 2: CORS com allowlist
- [x] Task 3: Rate limiting
- [x] Task 4: Security headers Next.js
- [x] Task 5: Docker non-root + health checks

### Funcionalidade Core (90% → 100%)
- [x] Task 6: Migration desconto + reset
- [x] Task 7: SQLC queries novas
- [x] Task 8: Desconto referral aplicado no PSP
- [x] Task 9: Email templates HTML
- [x] Task 10: Password reset backend
- [x] Task 11: Profile/account backend
- [x] Task 12: Emails transacionais

### Frontend UX (75% → 100%)
- [x] Task 13: Error boundary + 404
- [x] Task 14: Next.js middleware
- [x] Task 15: Loading skeleton
- [x] Task 16: Password reset frontend
- [x] Task 17: Profile page frontend
- [x] Task 18: API client retry

### Infraestrutura/DevOps (65% → 100%)
- [x] Task 19: CI/CD GitHub Actions
- [x] Task 20: Docker compose prod
- [x] Task 21: .env.example
- [x] Task 22: Auto-migrate startup
- [x] Task 23: Backup script
- [x] Task 24: Logging request ID
- [x] Task 25: Webhook logging
- [x] Task 26: Makefile updates
- [x] Task 27: Verificacao final
