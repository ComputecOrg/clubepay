-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, phone, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password_hash, name, phone, role, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, phone, role, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, phone, role, created_at
FROM users
WHERE id = $1;
