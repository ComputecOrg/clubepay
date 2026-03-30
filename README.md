# ClubePay

Plataforma SaaS que permite qualquer negocio fisico criar um **clube de assinatura** e cobrar via **Pix recorrente**, em 5 minutos. MVP focado em cafeterias e padarias.

O dono cadastra o negocio, cria planos (ex: "1 cafe por dia - R$49/mes") e compartilha o link. O assinante se cadastra, paga via Pix (Asaas) e valida o uso escaneando um QR code no balcao.

## Como funciona

```
Dono cadastra negocio ──> Cria planos ──> Compartilha link
                                              │
Assinante acessa link ──> Assina plano ──> Paga via Pix (Asaas)
                                              │
Assinante vai ao local ──> Escaneia QR ──> Uso validado
```

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.22+, chi, sqlc, pgx, JWT |
| Frontend | Next.js 15+, TypeScript, Tailwind CSS 4 |
| Banco | PostgreSQL 16 |
| Pagamento | Asaas API (Pix recorrente) |
| Proxy | Caddy (HTTPS automatico via Let's Encrypt) |
| Deploy | Docker + Coolify na VPS |

## Rodando localmente

### Pre-requisitos

- [Docker](https://docs.docker.com/get-docker/) e Docker Compose
- Ou: Go 1.22+, Node.js 22+, PostgreSQL 16

### Opcao 1: Docker (recomendado)

```bash
# Sobe PostgreSQL + backend + frontend
docker compose up -d

# Frontend: http://localhost:3000
# Backend:  http://localhost:8080
# Parar:    docker compose down
```

### Opcao 2: Local

```bash
# 1. Sobe so o banco
docker compose up -d postgres

# 2. Configura variaveis
cp .env.example .env
# Edita .env com JWT_SECRET (gerar com: openssl rand -hex 32)

# 3. Backend (migra o banco automaticamente)
cd backend && go run ./cmd/api

# 4. Frontend (em outro terminal)
cd frontend && npm install && npm run dev
```

## Testes

```bash
# Backend (Go)
cd backend && go test ./...

# Frontend (Vitest)
cd frontend && npx vitest run

# Tudo junto (requer make)
make test
```

## Estrutura do projeto

```
clubepay/
├── backend/
│   ├── cmd/api/              # Entrypoint + router
│   ├── internal/
│   │   ├── handler/          # HTTP handlers
│   │   ├── middleware/       # Auth, CORS, rate limit, logging
│   │   ├── domain/           # Regras de negocio
│   │   ├── repository/       # Queries (sqlc)
│   │   ├── psp/              # Interface pagamento (Asaas + Mock)
│   │   ├── email/            # SMTP + templates
│   │   └── config/           # Env vars
│   └── migrations/           # SQL (golang-migrate)
├── frontend/
│   ├── src/app/              # Pages (App Router)
│   ├── src/components/       # Componentes reutilizaveis
│   └── src/lib/              # API client, auth
├── Dockerfile.backend
├── Dockerfile.frontend
├── docker-compose.yml        # Dev
├── docker-compose.prod.yml   # Producao (com Caddy)
├── Caddyfile                 # Reverse proxy
└── Makefile
```

## API

### Autenticacao
| Metodo | Rota | Descricao |
|--------|------|-----------|
| POST | `/api/auth/register` | Cadastro do dono |
| POST | `/api/auth/login` | Login (retorna JWT) |
| POST | `/api/auth/register-subscriber` | Cadastro do assinante |
| POST | `/api/auth/request-password-reset` | Solicitar reset de senha |
| POST | `/api/auth/confirm-password-reset` | Confirmar reset de senha |

### Negocio (dono)
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | `/api/business` | Dados do negocio |
| PUT | `/api/business` | Atualizar negocio |

### Planos (dono)
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | `/api/plans` | Listar planos |
| POST | `/api/plans` | Criar plano |
| PUT | `/api/plans/:id` | Atualizar plano |
| DELETE | `/api/plans/:id` | Desativar plano |

### Assinaturas
| Metodo | Rota | Descricao |
|--------|------|-----------|
| POST | `/api/subscribe` | Assinar plano |
| GET | `/api/subscriptions` | Listar assinantes (dono) |
| DELETE | `/api/subscriptions/:id` | Cancelar assinante (dono) |
| GET | `/api/my-plan` | Plano do assinante |
| POST | `/api/cancel` | Assinante cancela propria assinatura |

### Uso
| Metodo | Rota | Descricao |
|--------|------|-----------|
| POST | `/api/validate-usage` | Validar uso (assinante escaneia QR) |
| POST | `/api/validate-usage-owner` | Validar uso (dono busca assinante) |
| GET | `/api/my-usage` | Historico de usos do assinante |

### Indicacoes
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | `/api/my-referral-code` | Codigo de indicacao |
| POST | `/api/referrals/apply` | Aplicar codigo |

### Publico (sem auth)
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | `/api/public/business/:slug` | Dados do negocio (landing page) |
| GET | `/api/public/plans/:slug` | Planos do negocio (landing page) |

### Webhook e Cron
| Metodo | Rota | Descricao |
|--------|------|-----------|
| POST | `/api/psp/webhook` | Webhook do Asaas (HMAC) |
| POST | `/api/cron/reconcile` | Reconciliacao diaria |

## Regras de negocio

- **Limites de uso:** `daily` (ex: 1 cafe/dia) ou `monthly` (ex: 4 cortes/mes)
- **Dunning:** Pagamento falhou → 3 dias de carencia → bloqueia uso → email ao assinante
- **Indicacoes:** 10% desconto para ambos, maximo 3 indicacoes ativas por assinante
- **Free tier:** Ate 15 assinantes + 1 plano. Banner de upgrade a partir de 80% do limite
- **Cancelamento:** Acesso continua ate o fim do periodo pago
- **Validacao:** QR code fixo no balcao. Assinante escaneia logado no celular

## Deploy (producao)

```bash
# 1. Configura variaveis de producao
cp .env.example .env
# Preenche: DATABASE_URL, JWT_SECRET, ASAAS_API_KEY, DOMAIN, SMTP_*, etc.

# 2. Sobe tudo (PostgreSQL + backend + frontend + Caddy com HTTPS)
docker compose -f docker-compose.prod.yml up -d
```

O Caddy provisiona certificados SSL automaticamente via Let's Encrypt.

### CI/CD

Push na `main` ou `master` dispara:
1. Testes backend (Go) + frontend (Vitest + ESLint)
2. Build e push das imagens Docker para GHCR
3. Deploy automatico via webhook do Coolify

## Variaveis de ambiente

Ver [`.env.example`](.env.example) para a lista completa com descricoes.

| Variavel | Obrigatoria | Descricao |
|----------|-------------|-----------|
| `DATABASE_URL` | Sim | Connection string PostgreSQL |
| `JWT_SECRET` | Sim | Secret para assinar tokens JWT |
| `ASAAS_API_KEY` | Producao | Chave da API Asaas (vazio = mock) |
| `ASAAS_WEBHOOK_SECRET` | Producao | Secret HMAC do webhook |
| `DOMAIN` | Producao | Dominio para Caddy/HTTPS |
| `SMTP_HOST` | Producao | Host SMTP para emails (vazio = mock) |

## Licenca

Proprietario. Todos os direitos reservados.
