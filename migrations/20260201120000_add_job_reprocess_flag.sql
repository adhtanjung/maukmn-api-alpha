-- +goose Up
-- +goose StatementBegin
ALTER TABLE image_processing_jobs ADD COLUMN is_reprocess BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE image_processing_jobs DROP COLUMN is_reprocess;
-- +goose StatementEnd
