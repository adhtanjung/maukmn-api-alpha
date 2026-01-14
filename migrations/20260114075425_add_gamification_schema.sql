-- +goose Up
-- +goose StatementBegin
-- Maukemana v1.3.0 Gamification Schema

-- 1. User Profiles (The Glass Passport)
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    username VARCHAR(50) UNIQUE,
    avatar_url TEXT,
    scout_level INTEGER DEFAULT 1,
    global_xp INTEGER DEFAULT 0,
    impact_score INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_username ON user_profiles(username);
CREATE INDEX IF NOT EXISTS idx_user_profiles_xp ON user_profiles(global_xp DESC);

-- 2. Territory Stats (Juragan Leaderboards)
CREATE TABLE IF NOT EXISTS user_territory_stats (
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    district_name VARCHAR(100) NOT NULL, -- Matched against addresses.kecamatan usually
    total_xp INTEGER DEFAULT 0,
    is_juragan BOOLEAN DEFAULT FALSE,
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, district_name)
);

CREATE INDEX IF NOT EXISTS idx_territory_stats_district_xp ON user_territory_stats(district_name, total_xp DESC);
CREATE INDEX IF NOT EXISTS idx_territory_stats_juragan ON user_territory_stats(district_name, is_juragan) WHERE is_juragan = TRUE;

-- 3. POI Gamification & Granular Data
-- Adding fields that support "The Game of Discovery"
ALTER TABLE points_of_interest
    ADD COLUMN IF NOT EXISTS founding_user_id UUID REFERENCES users(user_id),
    ADD COLUMN IF NOT EXISTS wifi_speed_mbps INTEGER,
    ADD COLUMN IF NOT EXISTS wifi_verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ergonomic_seating BOOLEAN DEFAULT FALSE,
    -- 'Everywhere', 'Wall-only', 'None' as per PRD.
    -- Naming it power_sockets_reach to distinguish from existing power_outlets enum
    ADD COLUMN IF NOT EXISTS power_sockets_reach VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_poi_founding_user ON points_of_interest(founding_user_id);
CREATE INDEX IF NOT EXISTS idx_poi_wifi_speed ON points_of_interest(wifi_speed_mbps) WHERE wifi_speed_mbps IS NOT NULL;

-- 4. Photos Gamification
ALTER TABLE photos
    ADD COLUMN IF NOT EXISTS vibe_category VARCHAR(50), -- 'Chair', 'Ambience', 'Menu'
    ADD COLUMN IF NOT EXISTS score INTEGER DEFAULT 0,   -- Denormalized Upvotes - Downvotes
    ADD COLUMN IF NOT EXISTS is_hero BOOLEAN DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_photos_score ON photos(poi_id, score DESC);
CREATE INDEX IF NOT EXISTS idx_photos_hero ON photos(poi_id) WHERE is_hero = TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert Maukemana v1.3.0 Gamification Schema

ALTER TABLE photos
    DROP COLUMN IF EXISTS is_hero,
    DROP COLUMN IF EXISTS score,
    DROP COLUMN IF EXISTS vibe_category;

ALTER TABLE points_of_interest
    DROP COLUMN IF EXISTS power_sockets_reach,
    DROP COLUMN IF EXISTS ergonomic_seating,
    DROP COLUMN IF EXISTS wifi_verified_at,
    DROP COLUMN IF EXISTS wifi_speed_mbps,
    DROP COLUMN IF EXISTS founding_user_id;

DROP TABLE IF EXISTS user_territory_stats;
DROP TABLE IF EXISTS user_profiles;
-- +goose StatementEnd
