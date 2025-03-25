-- +goose Up
ALTER TABLE refresh_tokens
ADD user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE refresh_tokens DROP COLUMN user_id;