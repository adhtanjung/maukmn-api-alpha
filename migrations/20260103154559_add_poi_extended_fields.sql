-- +goose Up
-- +goose StatementBegin
-- Add extended POI fields for full tab integration

-- Location fields
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS floor_unit VARCHAR(100);
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS public_transport TEXT;

-- Profile & Visuals fields
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS cover_image_url TEXT;
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS gallery_image_urls TEXT[];

-- Work & Productivity fields
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS wifi_quality VARCHAR(20) CHECK (wifi_quality IN ('', 'none', 'slow', 'moderate', 'fast', 'excellent'));
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS power_outlets VARCHAR(20) CHECK (power_outlets IN ('', 'none', 'limited', 'moderate', 'plenty'));
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS seating_options TEXT[];
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS noise_level VARCHAR(20) CHECK (noise_level IN ('', 'silent', 'quiet', 'moderate', 'lively', 'loud'));
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS has_ac BOOLEAN DEFAULT FALSE;

-- Atmosphere fields
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS vibes TEXT[];
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS crowd_type TEXT[];
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS lighting VARCHAR(20) CHECK (lighting IN ('', 'dim', 'moderate', 'bright', 'natural'));
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS music_type VARCHAR(100);
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS cleanliness VARCHAR(20) CHECK (cleanliness IN ('', 'poor', 'average', 'clean', 'spotless'));

-- Food & Drink fields (featured_menu_items and specials already exist)
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS dietary_options TEXT[];

-- Contact fields
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE points_of_interest ADD COLUMN IF NOT EXISTS email VARCHAR(255);

-- Create indexes for commonly queried fields
CREATE INDEX IF NOT EXISTS idx_poi_wifi_quality ON points_of_interest(wifi_quality);
CREATE INDEX IF NOT EXISTS idx_poi_noise_level ON points_of_interest(noise_level);
CREATE INDEX IF NOT EXISTS idx_poi_has_ac ON points_of_interest(has_ac);
CREATE INDEX IF NOT EXISTS idx_poi_vibes ON points_of_interest USING GIN (vibes);
CREATE INDEX IF NOT EXISTS idx_poi_seating_options ON points_of_interest USING GIN (seating_options);
CREATE INDEX IF NOT EXISTS idx_poi_dietary_options ON points_of_interest USING GIN (dietary_options);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- Remove extended POI fields

DROP INDEX IF EXISTS idx_poi_dietary_options;
DROP INDEX IF EXISTS idx_poi_seating_options;
DROP INDEX IF EXISTS idx_poi_vibes;
DROP INDEX IF EXISTS idx_poi_has_ac;
DROP INDEX IF EXISTS idx_poi_noise_level;
DROP INDEX IF EXISTS idx_poi_wifi_quality;

ALTER TABLE points_of_interest DROP COLUMN IF EXISTS email;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS phone;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS dietary_options;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS cleanliness;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS music_type;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS lighting;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS crowd_type;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS vibes;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS has_ac;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS noise_level;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS seating_options;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS power_outlets;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS wifi_quality;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS gallery_image_urls;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS cover_image_url;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS public_transport;
ALTER TABLE points_of_interest DROP COLUMN IF EXISTS floor_unit;

-- +goose StatementEnd
