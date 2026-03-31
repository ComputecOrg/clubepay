-- name: GetOrCreateMonthlyCost :one
SELECT id, business_id, month, infrastructure_cost_cents, claude_api_tokens, total_cost_cents, monthly_budget_cents, created_at, updated_at
FROM monthly_costs
WHERE business_id = $1 AND month = $2
LIMIT 1;

-- name: CreateMonthlyCost :one
INSERT INTO monthly_costs (business_id, month, monthly_budget_cents)
VALUES ($1, $2, $3)
RETURNING id, business_id, month, infrastructure_cost_cents, claude_api_tokens, total_cost_cents, monthly_budget_cents, created_at, updated_at;

-- name: UpdateMonthlyCostInfrastructure :exec
UPDATE monthly_costs
SET infrastructure_cost_cents = infrastructure_cost_cents + $2,
    total_cost_cents = total_cost_cents + $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateMonthlyCostTokens :exec
UPDATE monthly_costs
SET claude_api_tokens = claude_api_tokens + $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetMonthlyCostByID :one
SELECT id, business_id, month, infrastructure_cost_cents, claude_api_tokens, total_cost_cents, monthly_budget_cents, created_at, updated_at
FROM monthly_costs
WHERE id = $1
LIMIT 1;

-- name: CreateSpendingAlert :one
INSERT INTO spending_alerts (business_id, monthly_cost_id, alert_level, threshold_percent)
VALUES ($1, $2, $3, $4)
RETURNING id, business_id, monthly_cost_id, alert_level, threshold_percent, sent_at, created_at;

-- name: GetRecentSpendingAlert :one
SELECT id, business_id, monthly_cost_id, alert_level, threshold_percent, sent_at, created_at
FROM spending_alerts
WHERE business_id = $1 AND monthly_cost_id = $2 AND alert_level = $3
ORDER BY sent_at DESC
LIMIT 1;

-- name: ListMonthlyCosts :many
SELECT id, business_id, month, infrastructure_cost_cents, claude_api_tokens, total_cost_cents, monthly_budget_cents, created_at, updated_at
FROM monthly_costs
WHERE business_id = $1
ORDER BY month DESC
LIMIT $2 OFFSET $3;

-- name: ListSpendingAlerts :many
SELECT id, business_id, monthly_cost_id, alert_level, threshold_percent, sent_at, created_at
FROM spending_alerts
WHERE business_id = $1
ORDER BY sent_at DESC
LIMIT $2 OFFSET $3;
