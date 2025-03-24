-- +goose Up
ALTER TABLE users
ADD password VARCHAR(100) NOT NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN password;