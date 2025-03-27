-- name: CreateMessage :one
INSERT INTO messages (id, sender_id, recipient_id, content, created_at, ttl_seconds, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSentMessages :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE sender_id = $1;

-- name: GetRecievedMessages :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE recipient_id = $1;