-- name: CreateBusiness :one
INSERT INTO businesses (owner_id, name, slug, segment, address, logo_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetBusinessByOwnerID :one
SELECT * FROM businesses WHERE owner_id = $1;

-- name: GetBusinessBySlug :one
SELECT * FROM businesses WHERE slug = $1;

-- name: UpdateBusiness :one
UPDATE businesses
SET name = $2, segment = $3, address = $4, logo_url = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;
