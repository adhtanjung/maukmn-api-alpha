-- +goose Up
-- +goose StatementBegin
ALTER TYPE processing_status ADD VALUE IF NOT EXISTS 'downloading';
ALTER TYPE processing_status ADD VALUE IF NOT EXISTS 'uploading';
-- +goose StatementEnd

-- +goose Down
-- Note: PostgreSQL does not support removing values from an ENUM type.
-- The down migration is a no-op. To fully revert, recreate the type.
