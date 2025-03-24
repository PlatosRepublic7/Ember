-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserLoginInfo :one
SELECT id, username, password FROM users WHERE username = $1; 