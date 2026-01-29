package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"maukemana-backend/internal/database"
	"maukemana-backend/internal/models"

	"github.com/google/uuid"
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

// PhotosJSON handles JSON scanning for photos
type PhotosJSON []models.Photo

// Scan implements the sql.Scanner interface
func (p *PhotosJSON) Scan(value interface{}) error {
	if value == nil {
		*p = []models.Photo{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, p)
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
	GalleryImages       PhotosJSON       `db:"gallery_images" json:"gallery_images,omitempty"`
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
	// Gamification & Granular Data
	FoundingUserID       *uuid.UUID `db:"founding_user_id" json:"founding_user_id,omitempty"`
	WifiSpeedMbps        *int       `db:"wifi_speed_mbps" json:"wifi_speed_mbps,omitempty"`
	WifiVerifiedAt       *time.Time `db:"wifi_verified_at" json:"wifi_verified_at,omitempty"`
	ErgonomicSeating     bool       `db:"ergonomic_seating" json:"ergonomic_seating"`
	PowerSocketsReach    *string    `db:"power_sockets_reach" json:"power_sockets_reach,omitempty"`
	FoundingUserUsername *string    `db:"founding_user_username" json:"founding_user_username"`
	RatingAvg            float64    `db:"rating_avg" json:"rating_avg"`
	ReviewsCount         int        `db:"reviews_count" json:"reviews_count"`
	SavedAt              *time.Time `db:"saved_at" json:"saved_at,omitempty"`
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
	District             *string // Kecamatan
	City                 *string // Kabupaten
	Village              *string // Kelurahan
	PostalCode           *string
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
	CreatedBy     *uuid.UUID
	InitialStatus *string // Defaults to 'draft' if nil

	// Gamification & Granular Data (New in v1.3.0)
	WifiSpeedMbps     *int
	ErgonomicSeating  bool
	PowerSocketsReach *string
	FoundingUserID    *uuid.UUID
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
	District             *string // Kecamatan
	City                 *string // Kabupaten
	Village              *string // Kelurahan
	PostalCode           *string
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

	// Gamification & Granular Data (New in v1.3.0)
	WifiSpeedMbps     *int
	ErgonomicSeating  bool
	PowerSocketsReach *string
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
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// UpdateLocation updates location specific fields
func (r *POIRepository) UpdateLocation(ctx context.Context, poiID uuid.UUID, input CreatePOIInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("update location begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Handle Address
	var addressID *uuid.UUID

	// Check if POI already has an address
	var existingAddressID *uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT address_id FROM points_of_interest WHERE poi_id = $1", poiID).Scan(&existingAddressID)
	if err != nil {
		return fmt.Errorf("update location check address: %w", err)
	}

	if input.Address != nil && *input.Address != "" {
		if existingAddressID != nil {
			// Update existing address
			_, err = tx.ExecContext(ctx, "UPDATE addresses SET street_address = $1 WHERE address_id = $2", *input.Address, existingAddressID)
			if err != nil {
				return fmt.Errorf("update address: %w", err)
			}
			addressID = existingAddressID
		} else {
			// Insert new address
			var newAddrID uuid.UUID
			err = tx.QueryRowContext(ctx, "INSERT INTO addresses (street_address) VALUES ($1) RETURNING address_id", *input.Address).Scan(&newAddrID)
			if err != nil {
				return fmt.Errorf("insert address: %w", err)
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
		return fmt.Errorf("update location update poi: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update location commit: %w", err)
	}

	return nil
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
	if err != nil {
		return fmt.Errorf("update operations: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("update work prod: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("update atmosphere: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("update food drink: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("update social: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("update contact: %w", err)
	}
	return nil
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
		return nil, fmt.Errorf("get user pois: %w", err)
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
		return nil, fmt.Errorf("get pois by status: %w", err)
	}
	return pois, nil
}
