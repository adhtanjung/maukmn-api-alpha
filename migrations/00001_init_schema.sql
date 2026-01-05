-- +goose Up
-- +goose StatementBegin
-- Maukemana PostgreSQL Schema
-- Based on SPEC.md v1.0.1

-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Hierarchical categories
CREATE TABLE categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_category_id UUID REFERENCES categories(category_id),
    name_key VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indonesia administrative hierarchy
CREATE TABLE addresses (
    address_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    street_address TEXT,
    kelurahan VARCHAR(100),
    kecamatan VARCHAR(100),
    kabupaten VARCHAR(100),
    provinsi VARCHAR(100),
    postal_code VARCHAR(10),
    boundary GEOGRAPHY(Polygon, 4326),
    display_name VARCHAR(255) GENERATED ALWAYS AS (
        COALESCE(street_address || ', ', '') || kelurahan || ', ' || kabupaten
    ) STORED
);

-- Admin-controlled vocabularies with aliasing
CREATE TABLE vocabularies (
    vocab_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vocab_type VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    aliases TEXT[],
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE
);

-- Users table for authentication
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    picture_url TEXT,
    google_id VARCHAR(255) UNIQUE,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);

-- Main POI table
CREATE TABLE points_of_interest (
    poi_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    category_id UUID REFERENCES categories(category_id),
    website VARCHAR(255),
    brand VARCHAR(100),
    description TEXT,

    -- Geospatial
    address_id UUID REFERENCES addresses(address_id),
    location GEOGRAPHY(Point, 4326),
    parking_info TEXT,

    -- Amenities (references vocabulary keys)
    amenities TEXT[],
    has_wifi BOOLEAN DEFAULT FALSE,
    outdoor_seating BOOLEAN DEFAULT FALSE,
    is_wheelchair_accessible BOOLEAN DEFAULT FALSE,
    has_delivery BOOLEAN DEFAULT FALSE,

    -- Food & Drink
    cuisine VARCHAR(100),
    price_range INTEGER CHECK (price_range BETWEEN 1 AND 4),
    food_options TEXT[],
    featured_menu_items VARCHAR(100)[],
    specials VARCHAR(100)[],

    -- Logistics
    open_hours JSONB,
    secondary_open_hours JSONB,
    reservation_platform VARCHAR(255),
    reservation_required BOOLEAN DEFAULT FALSE,
    payment_options TEXT[],
    wait_time_estimate INTEGER,

    -- Lifestyle
    kids_friendly BOOLEAN DEFAULT FALSE,
    smoker_friendly BOOLEAN DEFAULT FALSE,
    pet_friendly TEXT[],

    -- Social & Events
    social_media_links JSONB,
    events TEXT[],
    events_calendar JSONB,
    happy_hour_info TEXT,
    loyalty_program TEXT,

    -- Metadata
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    verified_by UUID REFERENCES users(user_id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- POI indexes
CREATE INDEX idx_poi_location ON points_of_interest USING GIST (location);
CREATE INDEX idx_poi_category ON points_of_interest(category_id);
CREATE INDEX idx_poi_has_wifi ON points_of_interest(has_wifi);
CREATE INDEX idx_poi_outdoor_seating ON points_of_interest(outdoor_seating);
CREATE INDEX idx_poi_price_range ON points_of_interest(price_range);
CREATE INDEX idx_poi_amenities ON points_of_interest USING GIN (amenities);
CREATE INDEX idx_poi_food_options ON points_of_interest USING GIN (food_options);
CREATE INDEX idx_poi_payment_options ON points_of_interest USING GIN (payment_options);

-- Separate reviews table (normalized)
CREATE TABLE reviews (
    review_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poi_id UUID REFERENCES points_of_interest(poi_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(user_id) NOT NULL,
    rating INTEGER CHECK (rating BETWEEN 1 AND 5),
    content TEXT,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_reviews_poi ON reviews(poi_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);
CREATE UNIQUE INDEX idx_reviews_user_poi ON reviews(user_id, poi_id);

-- Photos with time-decay scoring
CREATE TABLE photos (
    photo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poi_id UUID REFERENCES points_of_interest(poi_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(user_id),
    url TEXT NOT NULL,
    original_url TEXT,
    is_admin_official BOOLEAN DEFAULT FALSE,
    is_pinned BOOLEAN DEFAULT FALSE,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_photos_poi ON photos(poi_id);
CREATE INDEX idx_photos_user ON photos(user_id);
CREATE INDEX idx_photos_created ON photos(created_at DESC);

-- Vote tracking
CREATE TABLE photo_votes (
    vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    photo_id UUID REFERENCES photos(photo_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(user_id) NOT NULL,
    vote_type SMALLINT CHECK (vote_type IN (-1, 1)),
    reason VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(photo_id, user_id)
);

CREATE INDEX idx_photo_votes_photo ON photo_votes(photo_id);
CREATE INDEX idx_photo_votes_user ON photo_votes(user_id);

-- Offline sync queue tracking
CREATE TABLE sync_queue (
    queue_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id) NOT NULL,
    entity_type VARCHAR(50),
    payload JSONB,
    status VARCHAR(20) DEFAULT 'pending',
    conflict_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    synced_at TIMESTAMPTZ
);

CREATE INDEX idx_sync_queue_user ON sync_queue(user_id);
CREATE INDEX idx_sync_queue_status ON sync_queue(status);

-- Saved POIs
CREATE TABLE saved_pois (
    saved_poi_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id) NOT NULL,
    poi_id UUID REFERENCES points_of_interest(poi_id) ON DELETE CASCADE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, poi_id)
);

CREATE INDEX idx_saved_pois_user ON saved_pois(user_id);
CREATE INDEX idx_saved_pois_poi ON saved_pois(poi_id);

-- Itineraries
CREATE TABLE itineraries (
    itinerary_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE,
    end_date DATE,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_itineraries_user ON itineraries(user_id);

-- Itinerary items
CREATE TABLE itinerary_items (
    item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id UUID REFERENCES itineraries(itinerary_id) ON DELETE CASCADE,
    poi_id UUID REFERENCES points_of_interest(poi_id) ON DELETE CASCADE,
    day INTEGER NOT NULL,
    order_index INTEGER NOT NULL,
    planned_time TIMESTAMPTZ,
    duration INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_itinerary_items_itinerary ON itinerary_items(itinerary_id);
CREATE INDEX idx_itinerary_items_order ON itinerary_items(itinerary_id, day, order_index);

-- Materialized view for performance
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
                 -- Time-decay hot score (Reddit-style)
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

-- Insert default categories
INSERT INTO categories (name_key, icon, parent_category_id) VALUES
    ('category.cafe', '☕', NULL),
    ('category.restaurant', '🍽️', NULL),
    ('category.bar', '🍺', NULL),
    ('category.attraction', '🎭', NULL),
    ('category.hotel', '🏨', NULL),
    ('category.shopping', '🛍️', NULL),
    ('category.activity', '🏃', NULL);

-- Insert default vocabularies
INSERT INTO vocabularies (vocab_type, key, aliases, icon) VALUES
    ('amenity', 'amenity.wifi', ARRAY['WiFi', 'Wifi', 'wi-fi', 'wireless'], '📶'),
    ('amenity', 'amenity.power_outlets', ARRAY['Power Outlets', 'Charging', 'Plugs'], '🔌'),
    ('amenity', 'amenity.outdoor_seating', ARRAY['Outdoor', 'Terrace', 'Patio'], '🌳'),
    ('amenity', 'amenity.parking', ARRAY['Parking', 'Car Park'], '🅿️'),
    ('amenity', 'amenity.wheelchair_accessible', ARRAY['Wheelchair', 'Accessible'], '♿'),
    ('food', 'food.vegan', ARRAY['Vegan'], '🌱'),
    ('food', 'food.vegetarian', ARRAY['Vegetarian'], '🥗'),
    ('food', 'food.halal', ARRAY['Halal'], '🕌'),
    ('food', 'food.gluten_free', ARRAY['Gluten Free', 'GF'], '🌾'),
    ('payment', 'payment.cash', ARRAY['Cash'], '💵'),
    ('payment', 'payment.qris', ARRAY['QRIS', 'QR'], '📱'),
    ('payment', 'payment.credit_card', ARRAY['Credit Card', 'Card'], '💳'),
    ('pet', 'pet.dogs', ARRAY['Dogs', 'Dog Friendly'], '🐕'),
    ('pet', 'pet.cats', ARRAY['Cats', 'Cat Friendly'], '🐈'),
    ('event', 'event.live_music', ARRAY['Live Music', 'Music'], '🎵');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS mv_pois_with_hero;
DROP TABLE IF EXISTS itinerary_items;
DROP TABLE IF EXISTS itineraries;
DROP TABLE IF EXISTS saved_pois;
DROP TABLE IF EXISTS sync_queue;
DROP TABLE IF EXISTS photo_votes;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS points_of_interest;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS vocabularies;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS categories;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS postgis;
-- +goose StatementEnd
