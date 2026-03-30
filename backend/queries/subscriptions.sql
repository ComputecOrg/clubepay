-- name: CreateSubscription :one
INSERT INTO subscriptions (plan_id, subscriber_id, business_id, psp_subscription_id, status, period_end, referred_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions WHERE id = $1;

-- name: GetSubscriptionByPSPID :one
SELECT * FROM subscriptions WHERE psp_subscription_id = $1;

-- name: ListSubscriptionsByBusinessID :many
SELECT s.*, u.name as subscriber_name, u.email as subscriber_email, u.phone as subscriber_phone, p.name as plan_name
FROM subscriptions s
JOIN users u ON u.id = s.subscriber_id
JOIN plans p ON p.id = s.plan_id
WHERE s.business_id = $1
ORDER BY s.created_at DESC;

-- name: GetActiveSubscription :one
SELECT * FROM subscriptions
WHERE subscriber_id = $1 AND business_id = $2 AND status IN ('active', 'grace')
LIMIT 1;

-- name: UpdateSubscriptionStatus :exec
UPDATE subscriptions SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateSubscriptionGrace :exec
UPDATE subscriptions SET status = 'grace', grace_deadline = $2, updated_at = NOW() WHERE id = $1;

-- name: CancelSubscription :exec
UPDATE subscriptions SET status = 'cancelled', updated_at = NOW() WHERE id = $1;

-- name: CountActiveSubscriptionsByBusinessID :one
SELECT COUNT(*) FROM subscriptions WHERE business_id = $1 AND status IN ('active', 'grace');

-- name: ListGraceExpiredSubscriptions :many
SELECT * FROM subscriptions WHERE status = 'grace' AND grace_deadline < NOW();

-- name: ListPendingSubscriptions :many
SELECT * FROM subscriptions WHERE status = 'pending';

-- name: CreateSubscriptionWithDiscount :one
INSERT INTO subscriptions (plan_id, subscriber_id, business_id, psp_subscription_id, status, period_end, referred_by, discount_percent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
