-- +goose Up
ALTER TABLE users
ADD hashed_password TEXT NOT NULL DEFAULT 'unset'; -- always use '' for string

-- +goose Down
ALTER TABLE users
DROP COLUMN hashed_password;
