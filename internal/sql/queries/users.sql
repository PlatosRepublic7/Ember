-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, email, password)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserLoginInfo :one
SELECT id, username, email, password FROM users WHERE email = $1; 

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(refresh_token, is_valid, created_at, updated_at, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE refresh_token = $1;

-- name: GetAllUserRefreshTokens :many
SELECT * FROM refresh_tokens WHERE user_id = $1;

-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens SET 
is_valid = $1, updated_at = $2 
WHERE refresh_token = $3;