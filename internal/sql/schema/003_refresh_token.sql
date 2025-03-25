-- +goose Up
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    refresh_token VARCHAR(256) NOT NULL,
    is_valid BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE refresh_tokens;