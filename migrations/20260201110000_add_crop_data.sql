-- +goose Up
-- +goose StatementBegin

ALTER TABLE image_processing_jobs
ADD COLUMN crop_data JSONB;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE image_processing_jobs
DROP COLUMN IF EXISTS crop_data;

-- +goose StatementEnd
