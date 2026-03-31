# ClubePay — Budget & Cost Tracking

**Data**: 2026-03-30
**Atualizado por Ops**: 2026-03-30
**Responsável**: Fin (CFO)
**Próxima Revisão**: 2026-04-02

---

## 1. Infrastructure Costs (Mensais)

### 1.1 Claude Max 20x Plan
- **Custo**: $200/mês
- **Limite**: ~900 msgs/5h window, ~480h Sonnet/week, ~40h Opus/week
- **Consumo Esperado**: 80% alocado para agentes, 20% buffer
- **Status**: Ativo ✅

### 1.2 Coolify VPS (Hosting)
- **Provider**: Hostinger VPS
- **Plano**: KVM1 (1 vCPU, 4GB RAM, 50GB NVMe SSD)
- **Data Center**: São Paulo, Brasil (latência baixa para usuários BR)
- **Custo Confirmado**: R$12,99/mês
- **Serviços hospedados**: PostgreSQL (512MB RAM limit), Go Backend (256MB), Next.js Frontend (256MB), Caddy (128MB) — total ~1,15GB, KVM1 com 4GB é mais que suficiente
- **Upgrade path**: KVM2 (R$22,99/mês, 2 vCPU, 8GB RAM) quando load aumentar
- **Requisitos Coolify**: mínimo 2GB RAM — KVM1 (4GB) ✅
- **Status**: Confirmado ✅ — provisionar antes do primeiro deploy em produção

### 1.3 Domain (clubepay.com.br)
- **Provider**: Não especificado
- **Custo Anual**: ~R$30-50
- **Custo Mensal**: ~R$2,50-4,20
- **Renovação**: Verificar data
- **Status**: Ativo ✅

### 1.4 VPN (para Blog)
- **Uso**: Blog privado conforme spec (Scribe agent)
- **Provider**: Possível Expressvpn/NordVPN/Surfshark
- **Custo Estimado**: R$10-30/mês
- **Status**: Possível/Opcional ⚠️

### 1.5 SMTP (Email)
- **Uso**: Notificações, password reset, transacionais
- **Provider Escolhido**: Brevo (ex-Sendinblue)
- **Plano**: Free tier
- **Limite free**: 300 emails/dia (9.000/mês) — suficiente para MVP com 5-50 clientes
- **Custo MVP**: R$0/mês (free tier)
- **Upgrade path**: Brevo Starter €25/mês (~R$145) para 20k emails/mês
- **Configuração**:
  - `SMTP_HOST`: smtp-relay.brevo.com
  - `SMTP_PORT`: 587 (TLS/STARTTLS)
  - `SMTP_USERNAME`: email cadastrado na conta Brevo
  - `SMTP_PASSWORD`: chave SMTP gerada no painel Brevo
- **Por que Brevo vs alternativas**:
  - SendGrid free: 100 emails/dia (menos generoso)
  - Mailgun free: 100 emails/dia
  - AWS SES: $0.10/1k emails (quase grátis, mas exige configuração IAM mais complexa)
  - Brevo: 300 emails/dia grátis, interface simples, suporte à LGPD
- **Status**: Confirmado ✅ — criar conta Brevo antes do deploy

### 1.6 Asaas (Payment Processor)
- **Função**: Processamento de pagamentos Pix recorrentes
- **Modelo de Cobrança** (taxas pagas pelos LOJISTAS, não pelo ClubePay):
  - Pix por assinatura: ~1,0% por transação + R$0,49 por cobrança gerada
  - Exemplo: plano R$49/mês → taxa Asaas ≈ R$0,98 (~2%)
  - ClubePay **não** toma comissão adicional — lojista recebe direto no Asaas
- **Custo do ClubePay com Asaas**: R$0/mês (ClubePay não paga taxas Asaas)
- **Volume Esperado**:
  - MVP: 0-5 clientes no mês 1
  - Transações por cliente: ~5-20/mês
  - Valor Médio/Transação: R$29-99
- **Tracking de custos via API Asaas**:
  - Endpoint: `GET /financialTransactions` — lista todas as movimentações financeiras
  - Endpoint: `GET /payments?status=RECEIVED` — cobranças recebidas
  - O `psp/asaas.go` atual já tem `GetPayments()` por assinatura
  - Para dashboard de volume total: adicionar `GetFinancialTransactions()` no futuro (tarefa separada)
- **Status**: Ativo ✅ (sandbox configurado; produção requer ASAAS_API_KEY real)

---

## 2. Baseline Budget

### Cenário 1: MVP (0-5 clientes reais) — Abril 2026

| Item | Custo Mensal | Moeda | Observações |
|------|--------|-------|-------------|
| Claude API | $200 | USD | Max 20x Plan |
| Coolify VPS | R$12,99 | BRL | Hostinger KVM1, São Paulo ✅ confirmado |
| Domain | R$3 | BRL | Renovação anualizada |
| VPN | R$15 | BRL | Opcional, para blog |
| SMTP | R$0 | BRL | Brevo free tier (300 emails/dia) ✅ confirmado |
| Asaas | R$0 | BRL | ClubePay não paga; taxa é do lojista ✅ confirmado |
| **Infraestrutura Total** | **$200 + R$30,99** | **USD + BRL** | **~R$1.031** |
| **Marketing (Meta Ads)** | **R$800** | **BRL** | **Abril 2026 — Aprovado CLU-33** |
| **Total Abril** | **~R$1.831** | **BRL** | **Incluindo growth marketing** |

> **Nota**: Infraestrutura R$1.031 + Marketing R$800 (aprovado para Sprint 1 growth)

### Cenário 2: Growth (5-50 clientes)

| Item | Custo Mensal | Observações |
|------|--------|-------------|
| Claude API | $200 | Pode aumentar conforme crescimento |
| Coolify VPS | R$22,99 | Upgrade Hostinger KVM2 (2 vCPU, 8GB RAM) |
| Domain | R$3 | Idem |
| VPN | R$15 | Idem |
| SMTP | R$145 | Upgrade Brevo Starter (€25/mês, 20k emails) |
| Asaas | R$0 | Continua custo zero para ClubePay |
| **Total** | **~R$1.186** | **Infraestrutura** |

---

## 2.5 Marketing Budget (Abril 2026)

### Meta Ads Campaign (Sprint 1 Growth)
- **Campanha**: Awareness (Padarias/Cafeterias) + Retargeting (site visitors)
- **Período**: 7–30 de abril 2026
- **Orçamento Aprovado**: R$800/mês
  - Awareness: R$500 (7–25 abr)
  - Retargeting: R$300 (25–30 abr)
- **Justificativa**: Principal canal pago para atingir Sprint 1 goal de 5 clientes reais
- **ROI Target**: CAC < R$160/cliente (5 clientes com R$800)
- **Aprovado por**: Fin (CFO) — CLU-33
- **Status**: ✅ Aprovado — 2026-03-31

---

## 3. Budget Allocation (Token Consumption)

De acordo com fin/AGENTS.md:

### Max 20x Plan Distribution (~900 msgs/5h window)

| Área | Agentes | % | Msgs | Limite |
|------|---------|---|----|--------|
| Engineering | Goh, Rex, Ops, Test, Archie, Pix | 60% | 432 | Pausar se >75% |
| Marketing | Hugo, Viral, Marta, Lens, Scribe, Help | 25% | 180 | Pausar se >60% |
| C-Suite | Clippy, Fin, Shield | 15% | 108 | Crítico |

### Throttle Rules (Implement)
- <50% window @ 2h: Normal
- >60% window @ 2h: Pause Help, Scribe, Lens
- >75% window @ 3h: Pause all marketing
- >90% anytime: Only critical tasks
- >80% weekly Sonnet: Double heartbeat intervals
- >50% weekly Opus: Switch heavy users to Sonnet

---

## 4. Alerts & Thresholds

### Alert Triggers

| Threshold | Ação | Destinatário |
|-----------|------|--------------|
| 80% monthly budget | 🟡 AVISO | Clippy (CEO) |
| 95% monthly budget | 🔴 CRÍTICO | Clippy + Board |
| Any rate limit breach | 🔴 CRÍTICO | Fin + Clippy |
| Infrastructure cost spike | 🟡 AVISO | Fin notifica Clippy |

### KPI Targets
- **Rate limit breaches**: 0
- **Budget accuracy**: ±10%
- **Cost per customer**: < R$50

---

## 5. Next Actions

- [x] ~~Setup alerts no Paperclip~~ → Implementado CLU-24 ✅
- [x] ~~Infraestrutura para rastreamento automático~~ → Phase 2 completa ✅
- [ ] Real-time provider API integration (Hostinger, Claude usage) — Goh/Fin
- [ ] CEO dashboard com histórico 3 meses — Rex/Frontend
- [ ] **Provisionar VPS Hostinger KVM1 (São Paulo)** — Ops/Archie
- [ ] **Criar conta Brevo e configurar SMTP** — Ops
- [ ] Adicionar `GetFinancialTransactions()` no psp/asaas.go para dashboard de volume — Rex/Goh
- [ ] Revisão de custos a cada 3 dias — Fin
- [ ] Relatório financeiro para o board — Fin

---

## 6. Links & Contacts

- **CTO (Archie)**: Infraestrutura, decisões técnicas
- **DevOps (Ops)**: Detalhes Coolify, VPS, custos reais
- **CEO (Clippy)**: Aprovações de realocação orçamentária >15%
- **Board**: Alertas críticos, aprovações maiores

---

**Última Atualização**: 2026-03-31 02:55 UTC (Ops — CLU-23: custos VPS e SMTP confirmados)
