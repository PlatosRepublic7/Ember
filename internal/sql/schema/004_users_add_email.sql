-- +goose Up
ALTER TABLE users
ADD email VARCHAR(128) UNIQUE NOT NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN email;