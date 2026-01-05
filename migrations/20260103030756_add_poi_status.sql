-- +goose Up
-- +goose StatementBegin
-- Add POI status workflow for admin approval

-- Add status column with workflow states
ALTER TABLE points_of_interest
ADD COLUMN status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved', 'rejected')),
ADD COLUMN submitted_at TIMESTAMPTZ,
ADD COLUMN rejected_reason TEXT,
ADD COLUMN created_by UUID REFERENCES users(user_id);

-- Index for filtering by status (admin queue, user drafts)
CREATE INDEX idx_poi_status ON points_of_interest(status);
CREATE INDEX idx_poi_created_by ON points_of_interest(created_by);

-- Update materialized view to only show approved POIs
DROP MATERIALIZED VIEW IF EXISTS mv_pois_with_hero;

CREATE MATERIALIZED VIEW mv_pois_with_hero AS
SELECT
    p.poi_id, p.name, p.category_id, p.location, p.has_wifi,
    p.outdoor_seating, p.price_range, p.pet_friendly, p.status,
    a.kelurahan, a.kabupaten, a.provinsi,
    (
        SELECT url FROM photos ph
        WHERE ph.poi_id = p.poi_id
          AND (ph.is_pinned = TRUE OR TRUE)
        ORDER BY ph.is_pinned DESC,
                 (ph.upvotes - ph.downvotes) /
                 POWER(EXTRACT(EPOCH FROM (NOW() - ph.created_at)) / 3600 + 2, 1.5) DESC
        LIMIT 1
    ) as hero_image_url,
    COALESCE(
        (SELECT AVG(rating)::DECIMAL(3,2) FROM reviews r WHERE r.poi_id = p.poi_id),
        0
    ) as rating_avg,
    (SELECT COUNT(*) FROM reviews r WHERE r.poi_id = p.poi_id) as reviews_count
FROM points_of_interest p
LEFT JOIN addresses a ON p.address_id = a.address_id
WHERE p.status = 'approved';

CREATE INDEX idx_mv_pois_location ON mv_pois_with_hero USING GIST (location);
CREATE INDEX idx_mv_pois_filters ON mv_pois_with_hero (has_wifi, outdoor_seating, price_range);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS mv_pois_with_hero;

-- Recreate original materialized view without status filter
CREATE MATERIALIZED VIEW mv_pois_with_hero AS
SELECT
    p.poi_id, p.name, p.category_id, p.location, p.has_wifi,
    p.outdoor_seating, p.price_range, p.pet_friendly,
    a.kelurahan, a.kabupaten, a.provinsi,
    (
        SELECT url FROM photos ph
        WHERE ph.poi_id = p.poi_id
          AND (ph.is_pinned = TRUE OR TRUE)
        ORDER BY ph.is_pinned DESC,
                 (ph.upvotes - ph.downvotes) /
                 POWER(EXTRACT(EPOCH FROM (NOW() - ph.created_at)) / 3600 + 2, 1.5) DESC
        LIMIT 1
    ) as hero_image_url,
    COALESCE(
        (SELECT AVG(rating)::DECIMAL(3,2) FROM reviews r WHERE r.poi_id = p.poi_id),
        0
    ) as rating_avg,
    (SELECT COUNT(*) FROM reviews r WHERE r.poi_id = p.poi_id) as reviews_count
FROM points_of_interest p
LEFT JOIN addresses a ON p.address_id = a.address_id;

CREATE INDEX idx_mv_pois_location ON mv_pois_with_hero USING GIST (location);
CREATE INDEX idx_mv_pois_filters ON mv_pois_with_hero (has_wifi, outdoor_seating, price_range);

ALTER TABLE points_of_interest
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS submitted_at,
DROP COLUMN IF EXISTS rejected_reason,
DROP COLUMN IF EXISTS created_by;

DROP INDEX IF EXISTS idx_poi_status;
DROP INDEX IF EXISTS idx_poi_created_by;
-- +goose StatementEnd
