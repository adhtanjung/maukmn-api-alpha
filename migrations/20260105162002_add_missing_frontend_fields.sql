-- +goose Up
-- +goose StatementBegin
-- Add missing frontend fields
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS category_ids TEXT[];
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS parking_options TEXT[];
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS pet_policy TEXT;

CREATE INDEX IF NOT EXISTS idx_poi_category_ids ON points_of_interest USING GIN (category_ids);
CREATE INDEX IF NOT EXISTS idx_poi_parking_options ON points_of_interest USING GIN (parking_options);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_poi_parking_options;
DROP INDEX IF EXISTS idx_poi_category_ids;

ALTER TABLE points_of_interest DROP COLUMN IF EXISTS pet_policy;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS parking_options;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS category_ids;
-- +goose StatementEnd
