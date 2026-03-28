-- name: SearchSubscribersByBusiness :many
SELECT u.id, u.name, u.email, u.phone,
       s.id as subscription_id, s.status,
       p.name as plan_name
FROM users u
JOIN subscriptions s ON s.subscriber_id = u.id
JOIN plans p ON p.id = s.plan_id
WHERE s.business_id = $1
  AND s.status IN ('active', 'grace')
  AND (u.name ILIKE '%' || $2 || '%' OR u.phone ILIKE '%' || $2 || '%')
ORDER BY u.name
LIMIT 20;
