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
	Website                *string        `db:"website" json:"website,omitempty"`
	Brand                  *string        `db:"brand" json:"brand,omitempty"`
	Description            *string        `db:"description" json:"description,omitempty"`
	AddressID              *uuid.UUID     `db:"address_id" json:"address_id,omitempty"`
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
	FloorUnit        *string        `db:"floor_unit" json:"floor_unit,omitempty"`
	PublicTransport  *string        `db:"public_transport" json:"public_transport,omitempty"`
	CoverImageURL    *string        `db:"cover_image_url" json:"cover_image_url,omitempty"`
	GalleryImageURLs pq.StringArray `db:"gallery_image_urls" json:"gallery_image_urls,omitempty"`
	WifiQuality      *string        `db:"wifi_quality" json:"wifi_quality,omitempty"`
	PowerOutlets     *string        `db:"power_outlets" json:"power_outlets,omitempty"`
	SeatingOptions   pq.StringArray `db:"seating_options" json:"seating_options,omitempty"`
	NoiseLevel       *string        `db:"noise_level" json:"noise_level,omitempty"`
	HasAC            bool           `db:"has_ac" json:"has_ac"`
	Vibes            pq.StringArray `db:"vibes" json:"vibes,omitempty"`
	CrowdType        pq.StringArray `db:"crowd_type" json:"crowd_type,omitempty"`
	Lighting         *string        `db:"lighting" json:"lighting,omitempty"`
	MusicType        *string        `db:"music_type" json:"music_type,omitempty"`
	Cleanliness      *string        `db:"cleanliness" json:"cleanliness,omitempty"`
	DietaryOptions   pq.StringArray `db:"dietary_options" json:"dietary_options,omitempty"`
	Phone            *string        `db:"phone" json:"phone,omitempty"`
	Email            *string        `db:"email" json:"email,omitempty"`
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

// UpdatePOIInput represents input for updating a POI
type UpdatePOIInput struct {
	Name           *string
	Description    *string
	HasWifi        *bool
	OutdoorSeating *bool
	PriceRange     *int
	Amenities      []string
}

// GetByID retrieves a POI by its ID
func (r *POIRepository) GetByID(ctx context.Context, poiID uuid.UUID) (*POI, error) {
	var poi POI
	query := `
		SELECT poi_id, name, category_id, website, brand, description,
		       address_id, parking_info, amenities, has_wifi, outdoor_seating,
		       is_wheelchair_accessible, has_delivery, cuisine, price_range,
		       food_options, payment_options, kids_friendly, smoker_friendly,
		       pet_friendly, is_verified, verified_at, created_at, updated_at
		FROM points_of_interest
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
		       pet_friendly, is_verified, verified_at, created_at, updated_at
		FROM points_of_interest
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if categoryID, ok := filters["category_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, categoryID)
		argCount++
	}

	if hasWifi, ok := filters["has_wifi"].(bool); ok {
		query += fmt.Sprintf(" AND has_wifi = $%d", argCount)
		args = append(args, hasWifi)
		argCount++
	}

	if priceRange, ok := filters["price_range"].(int); ok {
		query += fmt.Sprintf(" AND price_range = $%d", argCount)
		args = append(args, priceRange)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
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
			cover_image_url, gallery_image_urls,
			amenities, has_wifi, outdoor_seating, is_wheelchair_accessible,
			wifi_quality, power_outlets, seating_options, noise_level, has_ac,
			vibes, crowd_type, lighting, music_type, cleanliness,
			cuisine, price_range, food_options, dietary_options,
			featured_menu_items, specials,
			open_hours, reservation_required, reservation_platform,
			payment_options, wait_time_estimate,
			kids_friendly, pet_friendly, smoker_friendly,
			happy_hour_info, loyalty_program,
			phone, email, social_media_links,
			status, created_by
		) VALUES (
			$1, $2, $3, $4,
			ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography,
			NULL, $7, $8,
			$9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24,
			$25, $26, $27, $28,
			$29, $30,
			$31, $32, $33,
			$34, $35,
			$36, $37, $38,
			$39, $40,
			$41, $42, $43,
			'draft', $44
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
		input.CreatedBy,
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

// GetByUserAndStatus retrieves POIs by creator and status
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
		       has_wifi, outdoor_seating, price_range, submitted_at, created_at, updated_at
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
