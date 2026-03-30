-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, phone, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password_hash, name, phone, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, phone, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, phone, role, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users SET name = $2, phone = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW()
WHERE id = $1;
