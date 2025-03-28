-- name: CreateMessage :one
INSERT INTO messages (id, sender_id, recipient_id, content, created_at, ttl_seconds, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSentMessagesFromThisUser :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE sender_id = $1 AND deleted = false ORDER BY created_at DESC;

-- name: GetSentMessagesToNamedUser :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE sender_id = $1 AND recipient_id = $2 AND deleted = false ORDER BY created_at DESC;

-- name: GetReceivedMessagesFromNamedUser :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE recipient_id = $1 AND sender_id = $2 AND deleted = false ORDER BY created_at DESC;

-- name: GetReceivedMessagesToThisUser :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE recipient_id = $1 AND deleted = false ORDER BY created_at DESC;

-- name: GetUserMessageHistory :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE sender_id = $1 OR recipient_id = $1 AND deleted = false ORDER BY created_at DESC;

-- name: GetMessageHistoryWithNamedUser :many
SELECT id, sender_id, recipient_id, content, created_at, read_at, ttl_seconds, expires_at, deleted FROM
messages WHERE (sender_id = $1 AND recipient_id = $2) OR (sender_id = $2 AND recipient_id = $1) 
AND deleted = false ORDER BY created_at DESC;