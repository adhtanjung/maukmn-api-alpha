-- +goose Up
-- +goose StatementBegin

-- 1. Insert Cover Images (as Hero)
INSERT INTO photos (
    photo_id,
    poi_id,
    url,
    is_hero,
    score,
    upvotes,
    downvotes,
    is_pinned,
    is_admin_official,
    created_at
)
SELECT
    gen_random_uuid(),
    poi_id,
    cover_image_url,
    true, -- is_hero
    0, 0, 0, false, false,
    NOW()
FROM points_of_interest
WHERE cover_image_url IS NOT NULL
  AND cover_image_url != ''
  -- Avoid duplicates
  AND NOT EXISTS (
      SELECT 1 FROM photos p2
      WHERE p2.poi_id = points_of_interest.poi_id
        AND p2.url = points_of_interest.cover_image_url
  );

-- 2. Insert Gallery Images (as regular photos)
WITH gallery_items AS (
    SELECT poi_id, unnest(gallery_image_urls) as url
    FROM points_of_interest
    WHERE gallery_image_urls IS NOT NULL
)
INSERT INTO photos (
    photo_id,
    poi_id,
    url,
    is_hero,
    score,
    upvotes,
    downvotes,
    is_pinned,
    is_admin_official,
    created_at
)
SELECT
    gen_random_uuid(),
    poi_id,
    url,
    false, -- is_hero
    0, 0, 0, false, false,
    NOW()
FROM gallery_items
WHERE url IS NOT NULL
  AND url != ''
  -- Avoid duplicates
  AND NOT EXISTS (
      SELECT 1 FROM photos p2
      WHERE p2.poi_id = gallery_items.poi_id
        AND p2.url = gallery_items.url
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Leaving empty as deletion logic is ambiguous.
-- +goose StatementEnd
