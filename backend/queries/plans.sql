-- name: CreatePlan :one
INSERT INTO plans (business_id, name, description, price_cents, limit_type, limit_count)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListPlansByBusinessID :many
SELECT * FROM plans WHERE business_id = $1 AND active = true ORDER BY created_at;

-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1;

-- name: UpdatePlan :one
UPDATE plans
SET name = $2, description = $3, price_cents = $4, limit_type = $5, limit_count = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivatePlan :exec
UPDATE plans SET active = false, updated_at = NOW() WHERE id = $1;

-- name: CountPlansByBusinessID :one
SELECT COUNT(*) FROM plans WHERE business_id = $1 AND active = true;
