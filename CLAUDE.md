# ClubePay

Ferramenta que permite qualquer negócio físico criar um clube de assinatura e cobrar via Pix recorrente (Asaas), em 5 minutos. MVP focado em cafeterias/padarias.

## Metodologia

**TDD é obrigatório. Sem exceções.**

Todo código segue Red → Green → Refactor:
1. Escrever o teste que FALHA (red)
2. Implementar o MÍNIMO pra passar (green)
3. Refatorar (refactor)

Nunca escrever implementação antes do teste. Nunca adicionar testes "depois."

## Stack

### Backend (Go)
- **Go 1.22+** — API REST
- **chi** — HTTP router
- **sqlc** — SQL puro → Go type-safe (sem ORM)
- **golang-migrate** — Migrations PostgreSQL
- **pgx** — Driver PostgreSQL
- **golang-jwt/jwt** — Auth JWT
- **bcrypt** — Hash de senhas
- **slog** — Logging estruturado (stdlib)
- **validator/v10** — Validação de input

### Frontend
- **Next.js 15+** (App Router) — SSR + SPA
- **TypeScript**
- **Tailwind CSS 4+**

### Database
- **PostgreSQL 16** — Provisionado via Coolify

### Pagamento
- **Asaas API** — Subscription + QR Pix mensal
- Dono recebe direto no Asaas (sem intermediação)

### Deploy
- **Coolify** na VPS própria
- **Traefik** (via Coolify) — Reverse proxy + HTTPS automático
- Push to deploy via GitHub webhook

## Testes

### Backend (Go)
```bash
cd backend && go test ./...
```
- **testing** (stdlib) + **testify** — Unit tests
- **httptest** (stdlib) — Testes de API
- **testcontainers-go** — Integration tests com PostgreSQL real

### Frontend
```bash
cd frontend && npx vitest run
```
- **Vitest** — Unit tests
- **Playwright** — E2E

### Regra: todos os testes DEVEM passar antes de qualquer commit.

## Estrutura do projeto

```
clubepay/
├── backend/                      # Go API
│   ├── cmd/api/main.go           # Entrypoint
│   ├── internal/
│   │   ├── handler/              # HTTP handlers + _test.go
│   │   ├── middleware/           # Auth, CORS, logging
│   │   ├── domain/              # Business logic pura + _test.go
│   │   ├── repository/          # sqlc generated + _test.go
│   │   ├── psp/                 # Interface + Asaas impl + _test.go
│   │   └── config/              # Env vars
│   ├── migrations/              # SQL (golang-migrate)
│   ├── sqlc.yaml
│   └── go.mod
├── frontend/                     # Next.js
│   ├── src/app/                 # App Router
│   ├── src/components/
│   ├── src/lib/api.ts           # HTTP client pro Go backend
│   └── tests/
├── Dockerfile.backend
├── Dockerfile.frontend
├── Makefile
└── CLAUDE.md
```

## Convenções

### Go
- Packages: singular, lowercase (`handler`, `domain`, `psp`)
- Tabelas SQL: plural snake_case (`businesses`, `subscriptions`)
- Struct fields: PascalCase (Go convention)
- Erros tipados, não strings genéricas
- Todo handler tem `_test.go` ao lado
- sqlc pra queries — SQL puro, sem ORM
- `internal/` pra tudo que não é público

### Frontend
- Componentes: PascalCase (`DashboardStats.tsx`)
- Rotas: App Router com route groups `(auth)`, `(public)`
- API calls: via `src/lib/api.ts` (client HTTP pro Go backend)
- Tailwind utility classes, sem CSS custom

### Commits
- Mensagens em português
- Formato: `tipo: descrição curta`
- Tipos: `feat`, `fix`, `test`, `refactor`, `docs`, `chore`

## API Endpoints

### Auth
- `POST /api/auth/register` — Cadastro do dono (email + senha)
- `POST /api/auth/login` — Login → JWT
- `POST /api/auth/register-subscriber` — Cadastro do assinante (checkout)

### Business
- `GET /api/business` — Dados do negócio do dono logado
- `PUT /api/business` — Atualizar negócio

### Plans
- `GET /api/plans` — Listar planos do negócio
- `POST /api/plans` — Criar plano
- `PUT /api/plans/:id` — Atualizar plano
- `DELETE /api/plans/:id` — Desativar plano

### Subscriptions
- `POST /api/subscribe` — Assinar plano (cria subscription no Asaas)
- `GET /api/subscriptions` — Listar assinantes (dono)
- `DELETE /api/subscriptions/:id` — Cancelar (dono cancela assinante)

### Usage
- `POST /api/validate-usage` — Validar uso (assinante logado)
- `GET /api/my-usage` — Usos do assinante no período

### Subscriber
- `GET /api/my-plan` — Dados do plano do assinante logado
- `POST /api/cancel` — Assinante cancela própria assinatura

### Referrals
- `GET /api/my-referral-code` — Código de indicação do assinante
- `POST /api/referrals/apply` — Aplicar código de indicação

### Webhook
- `POST /api/psp/webhook` — Webhook do Asaas (HMAC validation)

### Cron
- `POST /api/cron/reconcile` — Reconciliação diária (auth por secret)

### Spending (Owner)
- `GET /api/owner/spending/status` — Status atual de gastos (percentual, orçamento, restante)
- `GET /api/owner/spending/history?limit=10&offset=0` — Histórico de gastos por mês
- `GET /api/owner/spending/alerts?limit=10&offset=0` — Histórico de alertas com timestamps

### Public (sem auth)
- `GET /api/public/business/:slug` — Dados do negócio pra landing page (SSR)
- `GET /api/public/plans/:slug` — Planos do negócio pra landing page

## Design Tokens

- **Cor primária:** `#2a7d6e` (verde-água)
- **Accent/CTA:** `#d4a853` (warm gold)
- **Neutros:** Tailwind grays
- **Font headline:** Cabinet Grotesk ou Satoshi
- **Font body:** `system-ui`

## Regras de negócio

- **Dunning:** Pagamento falhou → 3 dias de carência (uso normal) → bloqueia uso → email ao assinante
- **Limites de uso:** `daily` (1 café/dia) ou `monthly` (4 banhos/mês)
- **Referral:** 10% desconto pra ambos, limite de 3 indicações ativas por assinante
- **Free tier:** Até 15 assinantes + 1 plano. Banner de upgrade a partir de 12 (80%).
- **Cancelamento:** Acesso continua até fim do período pago. Retry async se PSP falhar.
- **Validação:** QR do balcão (fixo). Assinante escaneia logado. Fallback: busca por nome/telefone.
