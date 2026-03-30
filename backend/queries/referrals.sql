-- name: CreateReferral :one
INSERT INTO referrals (referrer_id, referred_id, business_id, code)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetReferralByCode :one
SELECT * FROM referrals WHERE code = $1 AND active = true;

-- name: GetReferralCodeBySubscriberAndBusiness :one
SELECT code FROM referrals
WHERE referrer_id = $1 AND business_id = $2
LIMIT 1;

-- name: CountActiveReferralsByReferrer :one
SELECT COUNT(*) FROM referrals WHERE referrer_id = $1 AND business_id = $2 AND active = true;

-- name: GetReferralByReferredAndBusiness :one
SELECT * FROM referrals
WHERE referred_id = $1 AND business_id = $2 AND referrer_id != referred_id AND active = true
LIMIT 1;
