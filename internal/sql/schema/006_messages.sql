-- +goose Up
CREATE TABLE messages (
    id              UUID PRIMARY KEY,
    sender_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content         TEXT NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    read_at         TIMESTAMP,
    ttl_seconds     INT,
    expires_at      TIMESTAMP,
    deleted         BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE messages;