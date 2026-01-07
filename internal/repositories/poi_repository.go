package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"maukemana-backend/internal/database"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// POIRepository handles POI database operations
type POIRepository struct {
	db *database.DB
}

// NewPOIRepository creates a new POI repository
func NewPOIRepository(db *database.DB) *POIRepository {
	return &POIRepository{db: db}
}

// POI represents a Point of Interest from the database
type POI struct {
	PoiID                  uuid.UUID      `db:"poi_id" json:"poi_id"`
	Name                   string         `db:"name" json:"name"`
	CategoryID             *uuid.UUID     `db:"category_id" json:"category_id,omitempty"`
	CategoryNames          pq.StringArray `db:"category_names" json:"category_names,omitempty"`
	Website                *string        `db:"website" json:"website,omitempty"`
	Brand                  *string        `db:"brand" json:"brand,omitempty"`
	Description            *string        `db:"description" json:"description,omitempty"`
	AddressID              *uuid.UUID     `db:"address_id" json:"address_id,omitempty"`
	Latitude               float64        `db:"latitude" json:"latitude"`
	Longitude              float64        `db:"longitude" json:"longitude"`
	ParkingInfo            *string        `db:"parking_info" json:"parking_info,omitempty"`
	Amenities              pq.StringArray `db:"amenities" json:"amenities,omitempty"`
	HasWifi                bool           `db:"has_wifi" json:"has_wifi"`
	OutdoorSeating         bool           `db:"outdoor_seating" json:"outdoor_seating"`
	IsWheelchairAccessible bool           `db:"is_wheelchair_accessible" json:"is_wheelchair_accessible"`
	HasDelivery            bool           `db:"has_delivery" json:"has_delivery"`
	Cuisine                *string        `db:"cuisine" json:"cuisine,omitempty"`
	PriceRange             *int           `db:"price_range" json:"price_range,omitempty"`
	FoodOptions            pq.StringArray `db:"food_options" json:"food_options,omitempty"`
	PaymentOptions         pq.StringArray `db:"payment_options" json:"payment_options,omitempty"`
	KidsFriendly           bool           `db:"kids_friendly" json:"kids_friendly"`
	SmokerFriendly         bool           `db:"smoker_friendly" json:"smoker_friendly"`
	PetFriendly            pq.StringArray `db:"pet_friendly" json:"pet_friendly,omitempty"`
	// New extended fields
	FloorUnit           *string          `db:"floor_unit" json:"floor_unit,omitempty"`
	PublicTransport     *string          `db:"public_transport" json:"public_transport,omitempty"`
	CoverImageURL       *string          `db:"cover_image_url" json:"cover_image_url,omitempty"`
	GalleryImageURLs    pq.StringArray   `db:"gallery_image_urls" json:"gallery_image_urls,omitempty"`
	WifiQuality         *string          `db:"wifi_quality" json:"wifi_quality,omitempty"`
	PowerOutlets        *string          `db:"power_outlets" json:"power_outlets,omitempty"`
	SeatingOptions      pq.StringArray   `db:"seating_options" json:"seating_options,omitempty"`
	NoiseLevel          *string          `db:"noise_level" json:"noise_level,omitempty"`
	HasAC               bool             `db:"has_ac" json:"has_ac"`
	Vibes               pq.StringArray   `db:"vibes" json:"vibes,omitempty"`
	CrowdType           pq.StringArray   `db:"crowd_type" json:"crowd_type,omitempty"`
	Lighting            *string          `db:"lighting" json:"lighting,omitempty"`
	MusicType           *string          `db:"music_type" json:"music_type,omitempty"`
	Cleanliness         *string          `db:"cleanliness" json:"cleanliness,omitempty"`
	CategoryIDs         pq.StringArray   `db:"category_ids" json:"category_ids,omitempty"`
	ParkingOptions      pq.StringArray   `db:"parking_options" json:"parking_options,omitempty"`
	PetPolicy           *string          `db:"pet_policy" json:"pet_policy,omitempty"`
	DietaryOptions      pq.StringArray   `db:"dietary_options" json:"dietary_options,omitempty"`
	FeaturedItems       pq.StringArray   `db:"featured_menu_items" json:"featured_items,omitempty"`
	Specials            pq.StringArray   `db:"specials" json:"specials,omitempty"`
	OpenHours           *json.RawMessage `db:"open_hours" json:"open_hours,omitempty"`
	ReservationRequired bool             `db:"reservation_required" json:"reservation_required"`
	ReservationPlatform *string          `db:"reservation_platform" json:"reservation_platform,omitempty"`
	WaitTimeEstimate    *int             `db:"wait_time_estimate" json:"wait_time_estimate,omitempty"`
	HappyHourInfo       *string          `db:"happy_hour_info" json:"happy_hour_info,omitempty"`
	LoyaltyProgram      *string          `db:"loyalty_program" json:"loyalty_program,omitempty"`
	Phone               *string          `db:"phone" json:"phone,omitempty"`
	Email               *string          `db:"email" json:"email,omitempty"`
	SocialLinks         *json.RawMessage `db:"social_media_links" json:"social_links,omitempty"`
	// Status workflow fields
	Status         string     `db:"status" json:"status"`
	SubmittedAt    *time.Time `db:"submitted_at" json:"submitted_at,omitempty"`
	RejectedReason *string    `db:"rejected_reason" json:"rejected_reason,omitempty"`
	CreatedBy      *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	// Verification fields
	IsVerified bool       `db:"is_verified" json:"is_verified"`
	VerifiedAt *time.Time `db:"verified_at" json:"verified_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	// Fetched fields (e.g. from joins)
	Address *string `db:"address" json:"address,omitempty"`
}

// POIWithDistance represents a POI with distance from a point
type POIWithDistance struct {
	POI
	DistanceMeters float64 `db:"distance_meters" json:"distance_meters"`
}

// CreatePOIInput represents input for creating a POI
type CreatePOIInput struct {
	// Profile & Visuals
	Name             string
	BrandName        *string
	Categories       []string
	Description      *string
	CoverImageURL    *string
	GalleryImageURLs []string
	CategoryIDs      []string
	// Location
	Address              *string
	FloorUnit            *string
	Latitude             float64
	Longitude            float64
	PublicTransport      *string
	ParkingOptions       []string
	WheelchairAccessible bool
	// Work & Prod
	WifiQuality    *string
	PowerOutlets   *string
	SeatingOptions []string
	NoiseLevel     *string
	HasAC          bool
	// Atmosphere
	Vibes       []string
	CrowdType   []string
	Lighting    *string
	MusicType   *string
	Cleanliness *string
	// Food & Drink
	Cuisine        *string
	PriceRange     *int
	DietaryOptions []string
	FeaturedItems  []string
	Specials       []string
	// Operations
	OpenHours           map[string]interface{}
	ReservationRequired bool
	ReservationPlatform *string
	PaymentOptions      []string
	WaitTimeEstimate    *int
	// Social & Lifestyle
	KidsFriendly   bool
	PetFriendly    []string
	PetPolicy      *string
	SmokerFriendly bool
	HappyHourInfo  *string
	LoyaltyProgram *string
	// Contact
	Phone       *string
	Email       *string
	Website     *string
	SocialLinks map[string]interface{}
	// Metadata
	CreatedBy *uuid.UUID
}

// UpdatePOIInput represents input for updating a POI (partial update)
type UpdatePOIInput struct {
	Name           *string
	Description    *string
	HasWifi        *bool
	OutdoorSeating *bool
	PriceRange     *int
	Amenities      []string
}

// UpdateFullInput represents input for full POI update (same fields as create)
type UpdateFullInput struct {
	// Profile & Visuals
	Name             string
	BrandName        *string
	Categories       []string
	Description      *string
	CoverImageURL    *string
	GalleryImageURLs []string
	CategoryIDs      []string
	// Location
	Address              *string
	FloorUnit            *string
	Latitude             float64
	Longitude            float64
	PublicTransport      *string
	ParkingOptions       []string
	WheelchairAccessible bool
	// Work & Prod
	WifiQuality    *string
	PowerOutlets   *string
	SeatingOptions []string
	NoiseLevel     *string
	HasAC          bool
	// Atmosphere
	Vibes       []string
	CrowdType   []string
	Lighting    *string
	MusicType   *string
	Cleanliness *string
	// Food & Drink
	Cuisine        *string
	PriceRange     *int
	DietaryOptions []string
	FeaturedItems  []string
	Specials       []string
	// Operations
	OpenHours           map[string]interface{}
	ReservationRequired bool
	ReservationPlatform *string
	PaymentOptions      []string
	WaitTimeEstimate    *int
	// Social & Lifestyle
	KidsFriendly   bool
	PetFriendly    []string
	PetPolicy      *string
	SmokerFriendly bool
	HappyHourInfo  *string
	LoyaltyProgram *string
	// Contact
	Phone       *string
	Email       *string
	Website     *string
	SocialLinks map[string]interface{}
}

// GetByID retrieves a POI by its ID
func (r *POIRepository) GetByID(ctx context.Context, poiID uuid.UUID) (*POI, error) {
	var poi POI
	query := `
		SELECT poi_id, name, category_id, website, brand, description,
		       points_of_interest.address_id, parking_info, amenities, has_wifi, outdoor_seating,
		       is_wheelchair_accessible, has_delivery, cuisine, price_range,
		       food_options, payment_options, kids_friendly, smoker_friendly,
		       pet_friendly, status, is_verified, verified_at, created_at, updated_at,
		       floor_unit, public_transport, cover_image_url, gallery_image_urls,
		       wifi_quality, power_outlets, seating_options, noise_level, has_ac,
		       vibes, crowd_type, lighting, music_type, cleanliness, dietary_options,
		       featured_menu_items, specials, open_hours, reservation_required,
		       reservation_platform, wait_time_estimate, happy_hour_info, loyalty_program,
		       phone, email, social_media_links, category_ids, parking_options, pet_policy,
		       ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude,
		       (
		           SELECT array_agg(name_key)
		           FROM categories
		           WHERE category_id = points_of_interest.category_id
		              OR category_id::text = ANY(points_of_interest.category_ids)
		       ) as category_names,
		       a.street_address as address
		FROM points_of_interest
		LEFT JOIN addresses a ON points_of_interest.address_id = a.address_id
		WHERE poi_id = $1
	`

	err := r.db.GetContext(ctx, &poi, query, poiID)
	if err != nil {
		return nil, err
	}

	return &poi, nil
}

// GetNearby retrieves POIs within a radius (in meters) from a point
func (r *POIRepository) GetNearby(ctx context.Context, lat, lng float64, radiusMeters int, limit int) ([]POIWithDistance, error) {
	var pois []POIWithDistance

	query := `
		SELECT
			poi_id, name, category_id, website, brand, description,
			address_id, parking_info, amenities, has_wifi, outdoor_seating,
			is_wheelchair_accessible, has_delivery, cuisine, price_range,
			food_options, payment_options, kids_friendly, smoker_friendly,
			pet_friendly, is_verified, verified_at, created_at, updated_at,
			cover_image_url, gallery_image_urls, status,
			ST_Y(location::geometry) as latitude,
			ST_X(location::geometry) as longitude,
			ST_Distance(
				location,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
			) as distance_meters
		FROM points_of_interest
		WHERE location IS NOT NULL
		  AND ST_DWithin(
			location,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)
		ORDER BY distance_meters
		LIMIT $4
	`

	err := r.db.SelectContext(ctx, &pois, query, lng, lat, radiusMeters, limit)
	if err != nil {
		return nil, err
	}

	return pois, nil
}

// Search searches POIs with filters
func (r *POIRepository) Search(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]POI, error) {
	var pois []POI

	query := `
		SELECT poi_id, name, category_id, website, brand, description,
		       address_id, parking_info, amenities, has_wifi, outdoor_seating,
		       is_wheelchair_accessible, has_delivery, cuisine, price_range,
		       food_options, payment_options, kids_friendly, smoker_friendly,
		       pet_friendly, status, cover_image_url, gallery_image_urls,
		       is_verified, verified_at, created_at, updated_at,
		       ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude
		FROM points_of_interest
		WHERE 1=1
	`
	args := []interface{}{}
	paramIdx := 1

	if categoryID, ok := filters["category_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND category_id = $%d", paramIdx)
		args = append(args, categoryID)
		paramIdx++
	}

	if hasWifi, ok := filters["has_wifi"].(bool); ok {
		query += fmt.Sprintf(" AND has_wifi = $%d", paramIdx)
		args = append(args, hasWifi)
		paramIdx++
	}

	if priceRange, ok := filters["price_range"].(int); ok {
		query += fmt.Sprintf(" AND price_range = $%d", paramIdx)
		args = append(args, priceRange)
		paramIdx++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", paramIdx)
		args = append(args, status)
		paramIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &pois, query, args...)
	if err != nil {
		return nil, err
	}

	return pois, nil
}

// Create creates a new POI from input
func (r *POIRepository) Create(ctx context.Context, input CreatePOIInput) (*POI, error) {
	// Convert open_hours map to JSONB
	var openHoursJSON []byte
	var err error
	if input.OpenHours != nil {
		openHoursJSON, err = json.Marshal(input.OpenHours)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal open_hours: %w", err)
		}
	}

	// Convert social_links map to JSONB
	var socialLinksJSON []byte
	if input.SocialLinks != nil {
		socialLinksJSON, err = json.Marshal(input.SocialLinks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal social_links: %w", err)
		}
	}

	query := `
		INSERT INTO points_of_interest (
			name, brand, description, website, location,
			address_id, floor_unit, public_transport,
			cover_image_url, gallery_image_urls, category_ids,
			amenities, has_wifi, outdoor_seating, is_wheelchair_accessible,
			wifi_quality, power_outlets, seating_options, noise_level, has_ac,
			vibes, crowd_type, lighting, music_type, cleanliness,
			cuisine, price_range, food_options, dietary_options,
			featured_menu_items, specials,
			open_hours, reservation_required, reservation_platform,
			payment_options, wait_time_estimate,
			kids_friendly, pet_friendly, pet_policy, smoker_friendly,
			happy_hour_info, loyalty_program,
			phone, email, social_media_links,
			status, created_by, parking_options
		) VALUES (
			$1, $2, $3, $4,
			ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography,
			NULL, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26, $27, $28, $29,
			$30, $31,
			$32, $33, $34,
			$35, $36,
			$37, $38, $39, $40,
			$41, $42,
			$43, $44, $45,
			'draft', $46, $47
		)
		RETURNING poi_id, name, brand, description, status, created_by,
		          is_verified, created_at, updated_at
	`

	var poi POI
	err = r.db.QueryRowxContext(
		ctx,
		query,
		input.Name,
		input.BrandName,
		input.Description,
		input.Website,
		input.Longitude,
		input.Latitude,
		input.FloorUnit,
		input.PublicTransport,
		input.CoverImageURL,
		pq.StringArray(input.GalleryImageURLs), pq.StringArray(input.CategoryIDs),
		pq.StringArray([]string{}), // amenities (empty for now, specific fields used)
		input.WifiQuality != nil && *input.WifiQuality != "" && *input.WifiQuality != "none",
		len(input.SeatingOptions) > 0 && contains(input.SeatingOptions, "outdoor"),
		input.WheelchairAccessible,
		input.WifiQuality,
		input.PowerOutlets,
		pq.StringArray(input.SeatingOptions),
		input.NoiseLevel,
		input.HasAC,
		pq.StringArray(input.Vibes),
		pq.StringArray(input.CrowdType),
		input.Lighting,
		input.MusicType,
		input.Cleanliness,
		input.Cuisine,
		input.PriceRange,
		pq.StringArray(input.DietaryOptions), // food_options
		pq.StringArray(input.DietaryOptions),
		pq.StringArray(input.FeaturedItems),
		pq.StringArray(input.Specials),
		openHoursJSON,
		input.ReservationRequired,
		input.ReservationPlatform,
		pq.StringArray(input.PaymentOptions),
		input.WaitTimeEstimate,
		input.KidsFriendly,
		pq.StringArray(input.PetFriendly), input.PetPolicy,
		input.SmokerFriendly,
		input.HappyHourInfo,
		input.LoyaltyProgram,
		input.Phone,
		input.Email,
		socialLinksJSON,
		input.CreatedBy,
		pq.StringArray(input.ParkingOptions),
	).StructScan(&poi)

	if err != nil {
		return nil, err
	}

	return &poi, nil
}

// Helper function to check if slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Update updates specific fields of a POI
func (r *POIRepository) Update(ctx context.Context, poiID uuid.UUID, input UpdatePOIInput) error {
	query := `
		UPDATE points_of_interest SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			has_wifi = COALESCE($4, has_wifi),
			outdoor_seating = COALESCE($5, outdoor_seating),
			price_range = COALESCE($6, price_range),
			amenities = COALESCE($7, amenities),
			updated_at = NOW()
		WHERE poi_id = $1
	`

	var amenities interface{}
	if len(input.Amenities) > 0 {
		amenities = pq.StringArray(input.Amenities)
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		poiID,
		input.Name,
		input.Description,
		input.HasWifi,
		input.OutdoorSeating,
		input.PriceRange,
		amenities,
	)

	return err
}

// Delete deletes a POI by ID
func (r *POIRepository) Delete(ctx context.Context, poiID uuid.UUID) error {
	query := `DELETE FROM points_of_interest WHERE poi_id = $1`
	_, err := r.db.ExecContext(ctx, query, poiID)
	return err
}

// UpdateFull updates all fields of a POI
func (r *POIRepository) UpdateFull(ctx context.Context, poiID uuid.UUID, input UpdateFullInput) error {
	// Convert open_hours map to JSONB
	var openHoursJSON []byte
	var err error
	if input.OpenHours != nil {
		openHoursJSON, err = json.Marshal(input.OpenHours)
		if err != nil {
			return fmt.Errorf("failed to marshal open_hours: %w", err)
		}
	}

	// Convert social_links map to JSONB
	var socialLinksJSON []byte
	if input.SocialLinks != nil {
		socialLinksJSON, err = json.Marshal(input.SocialLinks)
		if err != nil {
			return fmt.Errorf("failed to marshal social_links: %w", err)
		}
	}

	query := `
		UPDATE points_of_interest SET
			name = $2, brand = $3, description = $4, website = $5,
			location = ST_SetSRID(ST_MakePoint($6, $7), 4326)::geography,
			floor_unit = $8, public_transport = $9,
			cover_image_url = $10, gallery_image_urls = $11,
			amenities = $12, has_wifi = $13, outdoor_seating = $14, is_wheelchair_accessible = $15,
			wifi_quality = $16, power_outlets = $17, seating_options = $18, noise_level = $19, has_ac = $20,
			vibes = $21, crowd_type = $22, lighting = $23, music_type = $24, cleanliness = $25,
			cuisine = $26, price_range = $27, food_options = $28, dietary_options = $29,
			featured_menu_items = $30, specials = $31,
			open_hours = $32, reservation_required = $33, reservation_platform = $34,
			payment_options = $35, wait_time_estimate = $36,
			kids_friendly = $37, pet_friendly = $38, smoker_friendly = $39,
			happy_hour_info = $40, loyalty_program = $41,
			phone = $42, email = $43, social_media_links = $44,
			updated_at = NOW()
		WHERE poi_id = $1
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		poiID,
		input.Name,
		input.BrandName,
		input.Description,
		input.Website,
		input.Longitude,
		input.Latitude,
		input.FloorUnit,
		input.PublicTransport,
		input.CoverImageURL,
		pq.StringArray(input.GalleryImageURLs),
		pq.StringArray(input.ParkingOptions), // amenities
		input.WifiQuality != nil && *input.WifiQuality != "" && *input.WifiQuality != "none",
		len(input.SeatingOptions) > 0 && contains(input.SeatingOptions, "outdoor"),
		input.WheelchairAccessible,
		input.WifiQuality,
		input.PowerOutlets,
		pq.StringArray(input.SeatingOptions),
		input.NoiseLevel,
		input.HasAC,
		pq.StringArray(input.Vibes),
		pq.StringArray(input.CrowdType),
		input.Lighting,
		input.MusicType,
		input.Cleanliness,
		input.Cuisine,
		input.PriceRange,
		pq.StringArray(input.DietaryOptions), // food_options
		pq.StringArray(input.DietaryOptions),
		pq.StringArray(input.FeaturedItems),
		pq.StringArray(input.Specials),
		openHoursJSON,
		input.ReservationRequired,
		input.ReservationPlatform,
		pq.StringArray(input.PaymentOptions),
		input.WaitTimeEstimate,
		input.KidsFriendly,
		pq.StringArray(input.PetFriendly),
		input.SmokerFriendly,
		input.HappyHourInfo,
		input.LoyaltyProgram,
		input.Phone,
		input.Email,
		socialLinksJSON,
	)

	return err
}

// GetByUser retrieves all POIs created by a specific user
func (r *POIRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]POI, int, error) {
	var pois []POI
	query := `
		SELECT poi_id, name, category_id, description, status, created_by,
		       cover_image_url, has_wifi, outdoor_seating, price_range, created_at, updated_at
		FROM points_of_interest
		WHERE created_by = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &pois, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM points_of_interest WHERE created_by = $1`
	err = r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return pois, 0, err
	}

	return pois, total, nil
}

// GetWithHeroImages retrieves POIs from the materialized view
func (r *POIRepository) GetWithHeroImages(ctx context.Context, limit, offset int) ([]map[string]interface{}, error) {
	var pois []map[string]interface{}

	query := `
		SELECT * FROM mv_pois_with_hero
		ORDER BY rating_avg DESC, reviews_count DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		poi := make(map[string]interface{})
		if err := rows.MapScan(poi); err != nil {
			return nil, err
		}
		pois = append(pois, poi)
	}

	return pois, nil
}

// BeginTx starts a new transaction
func (r *POIRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTx(ctx)
}

// UpdateStatus updates the status of a POI
func (r *POIRepository) UpdateStatus(ctx context.Context, poiID uuid.UUID, status string, rejectedReason *string) error {
	var query string
	var args []interface{}

	if status == "pending" {
		query = `UPDATE points_of_interest SET status = $2, submitted_at = NOW(), updated_at = NOW() WHERE poi_id = $1`
		args = []interface{}{poiID, status}
	} else if status == "rejected" {
		query = `UPDATE points_of_interest SET status = $2, rejected_reason = $3, updated_at = NOW() WHERE poi_id = $1`
		args = []interface{}{poiID, status, rejectedReason}
	} else {
		query = `UPDATE points_of_interest SET status = $2, updated_at = NOW() WHERE poi_id = $1`
		args = []interface{}{poiID, status}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// UpdateProfile updates profile and visual fields
func (r *POIRepository) UpdateProfile(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	query := `
		UPDATE points_of_interest SET
			name = $2, brand = $3, description = $4,
			cover_image_url = $5, gallery_image_urls = $6,
			category_id = (SELECT category_id FROM categories WHERE name_key = ANY($7) LIMIT 1),
			category_ids = (SELECT array_agg(category_id) FROM categories WHERE name_key = ANY($7)),
			updated_at = NOW()
		WHERE poi_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, poiID, input.Name, input.BrandName, input.Description, input.CoverImageURL, pq.StringArray(input.GalleryImageURLs), pq.StringArray(input.Categories))
	return err
}

// UpdateLocation updates location specific fields
func (r *POIRepository) UpdateLocation(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Handle Address
	var addressID *uuid.UUID

	// Check if POI already has an address
	var existingAddressID *uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT address_id FROM points_of_interest WHERE poi_id = $1", poiID).Scan(&existingAddressID)
	if err != nil {
		return err
	}

	if input.Address != nil && *input.Address != "" {
		if existingAddressID != nil {
			// Update existing address
			_, err = tx.ExecContext(ctx, "UPDATE addresses SET street_address = $1 WHERE address_id = $2", *input.Address, existingAddressID)
			if err != nil {
				return err
			}
			addressID = existingAddressID
		} else {
			// Insert new address
			var newAddrID uuid.UUID
			err = tx.QueryRowContext(ctx, "INSERT INTO addresses (street_address) VALUES ($1) RETURNING address_id", *input.Address).Scan(&newAddrID)
			if err != nil {
				return err
			}
			addressID = &newAddrID
		}
	} else {
		addressID = existingAddressID
	}

	// 2. Update POI Location
	query := `
		UPDATE points_of_interest SET
			location = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
			floor_unit = $4, public_transport = $5,
			address_id = COALESCE($8, address_id),
			parking_options = $6,
			is_wheelchair_accessible = $7,
			updated_at = NOW()
		WHERE poi_id = $1
	`
	_, err = tx.ExecContext(ctx, query, poiID, input.Longitude, input.Latitude, input.FloorUnit, input.PublicTransport, pq.StringArray(input.ParkingOptions), input.WheelchairAccessible, addressID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateOperations updates operational fields
func (r *POIRepository) UpdateOperations(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	openHoursJSON, _ := json.Marshal(input.OpenHours)
	query := `
		UPDATE points_of_interest SET
			open_hours = $1, reservation_required = $2, reservation_platform = $3,
			payment_options = $4, wait_time_estimate = $5,
			updated_at = NOW()
		WHERE poi_id = $6
	`
	_, err := r.db.ExecContext(ctx, query, openHoursJSON, input.ReservationRequired, input.ReservationPlatform, pq.StringArray(input.PaymentOptions), input.WaitTimeEstimate, poiID)
	return err
}

// UpdateWorkProd updates work and productivity fields
func (r *POIRepository) UpdateWorkProd(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	query := `
		UPDATE points_of_interest SET
			wifi_quality = $1, power_outlets = $2, seating_options = $3,
			noise_level = $4, has_ac = $5,
			updated_at = NOW()
		WHERE poi_id = $6
	`
	_, err := r.db.ExecContext(ctx, query, input.WifiQuality, input.PowerOutlets, pq.StringArray(input.SeatingOptions), input.NoiseLevel, input.HasAC, poiID)
	return err
}

// UpdateAtmosphere updates atmosphere fields
func (r *POIRepository) UpdateAtmosphere(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	query := `
		UPDATE points_of_interest SET
			vibes = $1, crowd_type = $2, lighting = $3,
			music_type = $4, cleanliness = $5,
			updated_at = NOW()
		WHERE poi_id = $6
	`
	_, err := r.db.ExecContext(ctx, query, pq.StringArray(input.Vibes), pq.StringArray(input.CrowdType), input.Lighting, input.MusicType, input.Cleanliness, poiID)
	return err
}

// UpdateFoodDrink updates food and drink fields
func (r *POIRepository) UpdateFoodDrink(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	query := `
		UPDATE points_of_interest SET
			cuisine = $1, price_range = $2, food_options = $3,
			featured_menu_items = $4, specials = $5,
			updated_at = NOW()
		WHERE poi_id = $6
	`
	// Note: mapping DietaryOptions to food_options column
	_, err := r.db.ExecContext(ctx, query, input.Cuisine, input.PriceRange, pq.StringArray(input.DietaryOptions), pq.StringArray(input.FeaturedItems), pq.StringArray(input.Specials), poiID)
	return err
}

// UpdateSocial updates social and lifestyle fields
func (r *POIRepository) UpdateSocial(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	query := `
		UPDATE points_of_interest SET
			kids_friendly = $1, pet_friendly = $2, smoker_friendly = $3,
			happy_hour_info = $4, loyalty_program = $5,
			pet_policy = $6,
			updated_at = NOW()
		WHERE poi_id = $7
	`
	_, err := r.db.ExecContext(ctx, query, input.KidsFriendly, pq.StringArray(input.PetFriendly), input.SmokerFriendly, input.HappyHourInfo, input.LoyaltyProgram, input.PetPolicy, poiID)
	return err
}

// UpdateContact updates contact fields
func (r *POIRepository) UpdateContact(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	socialLinksJSON, _ := json.Marshal(input.SocialLinks)
	query := `
		UPDATE points_of_interest SET
			phone = $1, email = $2, website = $3,
			social_media_links = $4,
			updated_at = NOW()
		WHERE poi_id = $5
	`
	_, err := r.db.ExecContext(ctx, query, input.Phone, input.Email, input.Website, socialLinksJSON, poiID)
	return err
}
func (r *POIRepository) GetByUserAndStatus(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]POI, error) {
	var pois []POI
	query := `
		SELECT poi_id, name, category_id, description, status, created_by,
		       has_wifi, outdoor_seating, price_range, created_at, updated_at
		FROM points_of_interest
		WHERE created_by = $1 AND status = $2
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	err := r.db.SelectContext(ctx, &pois, query, userID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	return pois, nil
}

// GetByStatus retrieves POIs by status (for admin queue)
func (r *POIRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]POI, error) {
	var pois []POI
	query := `
		SELECT poi_id, name, category_id, description, status, created_by,
		       cover_image_url, has_wifi, outdoor_seating, price_range, submitted_at, created_at, updated_at
		FROM points_of_interest
		WHERE status = $1
		ORDER BY submitted_at ASC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &pois, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	return pois, nil
}
