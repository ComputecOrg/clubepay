# ClubePay Spending Alerts Configuration

**Last Updated**: 2026-03-31
**Responsible**: Fin (CFO)
**For**: CEO, CFO, Finance Team

---

## Overview

ClubePay implements automated spending alerts at two thresholds:
- **80% (Warning)**: Email to CEO
- **95% (Critical)**: Email to CEO + board escalation

Alerts are calculated daily via cron reconciliation (`POST /api/cron/reconcile`).

---

## Configuration Variables

Set these environment variables in your `.env` or deployment:

| Variable | Default | Type | Description |
|----------|---------|------|-------------|
| `MONTHLY_BUDGET_CENTS` | 500000 | int64 | Monthly infrastructure budget in cents (R$5000 = 500000¢) |
| `SPENDING_ALERT_EMAIL` | ceo@clubepay.com | string | Email to receive alerts |
| `WARN_THRESHOLD_PCT` | 80 | int | Warning threshold (%) |
| `CRITICAL_THRESHOLD_PCT` | 95 | int | Critical threshold (%) |

### Example Configuration

```bash
# .env
MONTHLY_BUDGET_CENTS=500000              # R$5000 budget
SPENDING_ALERT_EMAIL=fin@clubepay.com    # Fin receives alerts
WARN_THRESHOLD_PCT=75                    # Lower threshold for early warning
CRITICAL_THRESHOLD_PCT=90                # Early critical alert
```

### Validation Rules

- `WARN_THRESHOLD_PCT` must be < `CRITICAL_THRESHOLD_PCT`
- Both must be between 0-100
- `MONTHLY_BUDGET_CENTS` must be > 0

---

## How Alerts Work

### 1. Daily Reconciliation

Every day, the cron job runs:
```bash
POST /api/cron/reconcile
Header: X-Cron-Secret: <CRON_SECRET>
```

This triggers `SendSpendingAlerts()` which:
1. Iterates all businesses
2. Gets/creates `monthly_costs` record for current month
3. Calculates spending %: `(total_cost / budget) * 100`
4. Checks if alert should be sent (no duplicates per month)
5. Sends email to `SPENDING_ALERT_EMAIL`

### 2. Alert Deduplication

Alerts are deduplicated per threshold per month:
- Only **one warning email** per business per month at 80%
- Only **one critical email** per business per month at 95%
- If spending reaches 95%, only **critical** alert is sent (not warning)

### 3. Cost Tracking

Costs are stored in `monthly_costs` table:

```sql
INSERT INTO monthly_costs (
    business_id,
    month,
    infrastructure_cost_cents,
    claude_api_tokens,
    total_cost_cents,
    monthly_budget_cents
) VALUES (...);
```

Fields:
- `infrastructure_cost_cents`: VPS, domain, email (static costs)
- `claude_api_tokens`: API usage (tracked separately)
- `total_cost_cents`: Sum of above
- `monthly_budget_cents`: Budget for this month

---

## Manual Cost Updates

Use the provided script to manually track infrastructure costs:

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/clubepay_dev"
./scripts/track-infrastructure-costs.sh
```

Or update directly via SQL:

```sql
UPDATE monthly_costs
SET infrastructure_cost_cents = 107000  -- $200 Claude + R$13 VPS + R$3 domain
WHERE business_id = 1 AND month = '2026-03-01';
```

---

## API Endpoints

### Get Current Spending Status

```bash
GET /api/owner/spending/status
Authorization: Bearer <JWT_TOKEN>

Response:
{
  "current_cost_cents": 50000,
  "budget_cents": 500000,
  "spending_percent": 10,
  "remaining_cents": 450000,
  "month": "2026-03-01T00:00:00Z",
  "alert_status": "normal",  // "normal" | "warning" | "critical"
  "last_alert_sent_at": null
}
```

### Get Monthly History

```bash
GET /api/owner/spending/history?limit=3&offset=0
Authorization: Bearer <JWT_TOKEN>

Response:
{
  "items": [
    {
      "month": "2026-03-01",
      "infrastructure_cost_cents": 107000,
      "claude_tokens": 0,
      "total_cost_cents": 107000,
      "budget_cents": 500000,
      "spending_percent": 21
    }
  ],
  "total": 1,
  "limit": 3,
  "offset": 0
}
```

### Get Alert History

```bash
GET /api/owner/spending/alerts?limit=10&offset=0
Authorization: Bearer <JWT_TOKEN>

Response:
{
  "items": [
    {
      "id": 1,
      "alert_level": "warning",      // "warning" | "critical"
      "threshold_percent": 80,
      "sent_at": "2026-03-15T14:30:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

---

## Troubleshooting

### Alerts Not Being Sent

1. **Check cron secret**: Ensure `X-Cron-Secret` header matches `CRON_SECRET` in deployment
2. **Check email config**: Ensure `SPENDING_ALERT_EMAIL` is valid and SMTP is configured
3. **Check logs**: Look for errors in cron reconciliation logs
4. **Check threshold**: Ensure spending actually exceeds `WARN_THRESHOLD_PCT`
5. **Check deduplication**: Confirm no alert was already sent this month (check `spending_alerts` table)

### Alert Already Sent This Month

Alerts are deduplicated:
```sql
SELECT * FROM spending_alerts
WHERE business_id = 1
  AND alert_level = 'warning'
  AND DATE_TRUNC('month', sent_at) = '2026-03-01';
```

To reset (testing only):
```sql
DELETE FROM spending_alerts
WHERE business_id = 1 AND alert_level = 'warning'
  AND DATE_TRUNC('month', sent_at) = '2026-03-01';
```

### Email Not Received

1. Check if `SPENDING_ALERT_EMAIL` is correct
2. Check if SMTP is configured (look for `SMTP_HOST`, `SMTP_USERNAME`, `SMTP_PASSWORD`)
3. Check spam folder
4. Verify email logs: `grep "ALERT" logs/backend.log`

---

## Next Steps (CFO Roadmap)

- [ ] **Phase 2**: Automate infrastructure cost updates (from Hostinger API, Claude usage)
- [ ] **Phase 3**: CEO dashboard + threshold adjustment UI
- [ ] **Phase 4**: Token consumption tracking
- [ ] **Phase 5**: Weekly financial reports for board

---

## Related Files

- **Schema**: `backend/migrations/000003_add_spending_tables.{up,down}.sql`
- **Handler**: `backend/internal/handler/handler.go` (SendSpendingAlerts)
- **Domain Logic**: `backend/internal/domain/spending.go`
- **Config**: `backend/internal/config/config.go`
- **Cron Integration**: `backend/internal/handler/cron.go`
- **Cost Tracking Script**: `scripts/track-infrastructure-costs.sh`

---

**Questions?** Contact Fin or Clippy (CEO)
