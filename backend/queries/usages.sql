-- name: CreateUsage :one
INSERT INTO usages (subscription_id, validated_at)
VALUES ($1, NOW())
RETURNING *;

-- name: CountDailyUsage :one
SELECT COUNT(*) FROM usages
WHERE subscription_id = $1
AND validated_at >= $2::timestamptz
AND validated_at < $3::timestamptz;

-- name: CountMonthlyUsage :one
SELECT COUNT(*) FROM usages
WHERE subscription_id = $1
AND validated_at >= $2::timestamptz
AND validated_at < $3::timestamptz;

-- name: ListUsagesBySubscription :many
SELECT * FROM usages
WHERE subscription_id = $1
AND validated_at >= $2::timestamptz
AND validated_at < $3::timestamptz
ORDER BY validated_at DESC;
