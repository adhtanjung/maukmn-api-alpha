-- +goose Up
-- +goose StatementBegin
-- Function to sync the highest scored/hero photo to the POI cover_image_url
CREATE OR REPLACE FUNCTION sync_poi_hero_image() RETURNS TRIGGER AS $$
DECLARE
    new_hero_url TEXT;
BEGIN
    -- Logic:
    -- 1. If a photo is set to is_hero=TRUE, it wins immediately.
    -- 2. Fallback: The highest scored photo (score DESC, created_at DESC).

    SELECT url INTO new_hero_url
    FROM photos
    WHERE poi_id = NEW.poi_id
    ORDER BY is_hero DESC, score DESC, created_at DESC
    LIMIT 1;

    -- Update the cache column on points_of_interest
    IF new_hero_url IS NOT NULL THEN
        UPDATE points_of_interest
        SET cover_image_url = new_hero_url,
            updated_at = NOW()
        WHERE poi_id = NEW.poi_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Fires when a photo is inserted, updated (score/is_hero), or deleted
CREATE TRIGGER trg_sync_hero_image
AFTER INSERT OR UPDATE OF score, is_hero OR DELETE
ON photos
FOR EACH ROW
EXECUTE FUNCTION sync_poi_hero_image();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_sync_hero_image ON photos;
DROP FUNCTION IF EXISTS sync_poi_hero_image;
-- +goose StatementEnd
