package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Create creates a new POI from input
func (r *POIRepository) Create(ctx context.Context, input CreatePOIInput) (*POI, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Handle Address Creation
	var addressID *uuid.UUID
	if input.Address != nil || input.District != nil || input.City != nil {
		// Try to find existing address or create new?
		// For now, always create new address for new POI to avoid complexity,
		// or maybe check? Let's just create new.
		var newAddrID uuid.UUID
		addrQuery := `
			INSERT INTO addresses (street_address, kecamatan, kabupaten, kelurahan, postal_code)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING address_id
		`
		// Prepare args handling nil pointers
		err = tx.QueryRowContext(ctx, addrQuery,
			input.Address,
			input.District,
			input.City,
			input.Village,
			input.PostalCode,
		).Scan(&newAddrID)

		if err != nil {
			return nil, fmt.Errorf("failed to create address: %w", err)
		}
		addressID = &newAddrID
	}

	// Convert open_hours map to JSONB
	var openHoursJSON interface{} = nil
	if input.OpenHours != nil {
		b, err := json.Marshal(input.OpenHours)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal open_hours: %w", err)
		}
		openHoursJSON = string(b)
	}

	// Convert social_links map to JSONB
	var socialLinksJSON interface{} = nil
	if input.SocialLinks != nil {
		b, err := json.Marshal(input.SocialLinks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal social_links: %w", err)
		}
		socialLinksJSON = string(b)
	}

	query := `
		INSERT INTO points_of_interest (
			name, brand, description, website, location,
			floor_unit, public_transport,
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
			status, created_by, parking_options,
			founding_user_id, wifi_speed_mbps, ergonomic_seating, power_sockets_reach,
			address_id
		) VALUES (
			$1, $2, $3, $4,
			ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography,
			$7, $8,
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
			COALESCE($53, 'draft'), $46, $47,
			$48, $49, $50, $51,
			$52
		)
		RETURNING poi_id, name, brand, description, status, created_by,
		          is_verified, created_at, updated_at, founding_user_id,
		          wifi_speed_mbps, ergonomic_seating, power_sockets_reach
	`

	var poi POI
	err = tx.QueryRowxContext(
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
		input.FoundingUserID,
		input.WifiSpeedMbps,
		input.ErgonomicSeating,
		input.PowerSocketsReach,
		addressID,
		input.InitialStatus,
	).StructScan(&poi)

	if err != nil {
		return nil, fmt.Errorf("create poi query: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &poi, nil
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

	if err != nil {
		return fmt.Errorf("update poi: %w", err)
	}
	return nil
}

// Delete deletes a POI by ID
func (r *POIRepository) Delete(ctx context.Context, poiID uuid.UUID) error {
	query := `DELETE FROM points_of_interest WHERE poi_id = $1`
	_, err := r.db.ExecContext(ctx, query, poiID)
	if err != nil {
		return fmt.Errorf("delete poi: %w", err)
	}
	return nil
}

// UpdateFull updates all fields of a POI
func (r *POIRepository) UpdateFull(ctx context.Context, poiID uuid.UUID, input UpdateFullInput) error {
	// Convert open_hours map to JSONB
	var openHoursJSON interface{} = nil
	if input.OpenHours != nil {
		b, err := json.Marshal(input.OpenHours)
		if err != nil {
			return fmt.Errorf("failed to marshal open_hours: %w", err)
		}
		// Explicitly cast to string to ensure driver treats it as text/json
		openHoursJSON = string(b)
	}

	// Convert social_links map to JSONB
	var socialLinksJSON interface{} = nil
	if input.SocialLinks != nil {
		b, err := json.Marshal(input.SocialLinks)
		if err != nil {
			return fmt.Errorf("failed to marshal social_links: %w", err)
		}
		socialLinksJSON = string(b)
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
			wifi_speed_mbps = $45, ergonomic_seating = $46, power_sockets_reach = $47,
			updated_at = NOW()
		WHERE poi_id = $1
	`

	_, err := r.db.ExecContext(
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
		input.WifiSpeedMbps,
		input.ErgonomicSeating,
		input.PowerSocketsReach,
	)

	if err != nil {
		return fmt.Errorf("update full poi: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}
