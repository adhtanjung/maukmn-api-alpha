-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN clerk_id VARCHAR(255) UNIQUE;
CREATE INDEX idx_users_clerk_id ON users(clerk_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_users_clerk_id;
ALTER TABLE users DROP COLUMN clerk_id;
-- +goose StatementEnd
