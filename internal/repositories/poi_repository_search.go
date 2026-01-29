package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Search searches POIs with filters
func (r *POIRepository) Search(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]POI, error) {
	var pois []POI

	// Check if we need distance calculation for sorting
	sortBy, _ := filters["sort_by"].(string)
	lat, hasLat := filters["lat"].(float64)
	lng, hasLng := filters["lng"].(float64)
	needsDistance := sortBy == "nearest" && hasLat && hasLng

	selectClause := `
		SELECT p.poi_id, p.name, p.category_id, p.website, p.brand, p.description,
		       p.address_id, p.parking_info, p.amenities, p.has_wifi, p.outdoor_seating,
		       p.is_wheelchair_accessible, p.has_delivery, p.cuisine, p.price_range,
		       p.food_options, p.payment_options, p.kids_friendly, p.smoker_friendly,
		       p.pet_friendly, p.status, p.cover_image_url, p.gallery_image_urls,
		       (
		           SELECT COALESCE(json_agg(
		               json_build_object(
		                   'photo_id', ph.photo_id,
		                   'poi_id', ph.poi_id,
		                   'url', ph.url,
		                   'is_hero', ph.is_hero,
		                   'score', ph.score,
		                   'upvotes', ph.upvotes,
		                   'downvotes', ph.downvotes,
		                   'is_pinned', ph.is_pinned,
		                   'is_admin_official', ph.is_admin_official,
		                   'created_at', ph.created_at
		               ) ORDER BY ph.is_pinned DESC, ph.is_hero DESC, ph.score DESC
		           ), '[]'::json)
		           FROM photos ph
		           WHERE ph.poi_id = p.poi_id
		       ) as gallery_images,
		       p.is_verified, p.verified_at, p.created_at, p.updated_at,
		       p.wifi_quality, p.power_outlets, p.noise_level, p.vibes, p.crowd_type,
		       p.seating_options, p.parking_options, p.has_ac, p.dietary_options,
		       p.founding_user_id, p.wifi_speed_mbps, p.wifi_verified_at, p.ergonomic_seating, p.power_sockets_reach,
		       ST_Y(p.location::geometry) as latitude, ST_X(p.location::geometry) as longitude,
		       u.name as founding_user_username,
		       COALESCE(
		           (SELECT AVG(rating)::float8 FROM reviews r WHERE r.poi_id = p.poi_id),
		           0
		       ) as rating_avg,
		       (SELECT COUNT(*)::int FROM reviews r WHERE r.poi_id = p.poi_id) as reviews_count`

	if needsDistance {
		selectClause += ",\n		       ST_Distance(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_meters"
	}

	query := selectClause + `
		FROM points_of_interest p
		LEFT JOIN users u ON COALESCE(p.founding_user_id, p.created_by) = u.user_id
		WHERE 1=1
	`

	args := []interface{}{}
	paramIdx := 1

	// If we need distance, add lat/lng as the first two parameters
	if needsDistance {
		args = append(args, lng, lat)
		paramIdx = 3
	}

	// Category filter
	if categoryID, ok := filters["category_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND category_id = $%d", paramIdx)
		args = append(args, categoryID)
		paramIdx++
	}

	// Legacy has_wifi boolean filter
	if hasWifi, ok := filters["has_wifi"].(bool); ok {
		query += fmt.Sprintf(" AND has_wifi = $%d", paramIdx)
		args = append(args, hasWifi)
		paramIdx++
	}

	// Price range filter
	if priceRange, ok := filters["price_range"].(int); ok {
		query += fmt.Sprintf(" AND price_range = $%d", paramIdx)
		args = append(args, priceRange)
		paramIdx++
	}

	// Status filter
	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", paramIdx)
		args = append(args, status)
		paramIdx++
	}

	// WiFi quality filter (string)
	if wifiQuality, ok := filters["wifi_quality"].(string); ok && wifiQuality != "" && wifiQuality != "any" {
		query += fmt.Sprintf(" AND wifi_quality = $%d", paramIdx)
		args = append(args, wifiQuality)
		paramIdx++
	}

	// Noise level filter (string)
	if noiseLevel, ok := filters["noise_level"].(string); ok && noiseLevel != "" {
		query += fmt.Sprintf(" AND noise_level = $%d", paramIdx)
		args = append(args, noiseLevel)
		paramIdx++
	}

	// Power outlets filter (string)
	if powerOutlets, ok := filters["power_outlets"].(string); ok && powerOutlets != "" && powerOutlets != "any" {
		query += fmt.Sprintf(" AND power_outlets = $%d", paramIdx)
		args = append(args, powerOutlets)
		paramIdx++
	}

	// Cuisine filter (string)
	if cuisine, ok := filters["cuisine"].(string); ok && cuisine != "" {
		query += fmt.Sprintf(" AND cuisine = $%d", paramIdx)
		args = append(args, cuisine)
		paramIdx++
	}

	// Has AC filter (boolean)
	if hasAC, ok := filters["has_ac"].(bool); ok {
		query += fmt.Sprintf(" AND has_ac = $%d", paramIdx)
		args = append(args, hasAC)
		paramIdx++
	}

	// Vibes filter (array - match any)
	if vibes, ok := filters["vibes"].([]string); ok && len(vibes) > 0 {
		query += fmt.Sprintf(" AND vibes && $%d", paramIdx)
		args = append(args, pq.StringArray(vibes))
		paramIdx++
	}

	// Crowd type filter (array - match any)
	if crowdType, ok := filters["crowd_type"].([]string); ok && len(crowdType) > 0 {
		query += fmt.Sprintf(" AND crowd_type && $%d", paramIdx)
		args = append(args, pq.StringArray(crowdType))
		paramIdx++
	}

	// Dietary options filter (array - match any)
	if dietaryOptions, ok := filters["dietary_options"].([]string); ok && len(dietaryOptions) > 0 {
		query += fmt.Sprintf(" AND dietary_options && $%d", paramIdx)
		args = append(args, pq.StringArray(dietaryOptions))
		paramIdx++
	}

	// Seating options filter (array - match any)
	if seatingOptions, ok := filters["seating_options"].([]string); ok && len(seatingOptions) > 0 {
		query += fmt.Sprintf(" AND seating_options && $%d", paramIdx)
		args = append(args, pq.StringArray(seatingOptions))
		paramIdx++
	}

	// Parking options filter (array - match any)
	if parkingOptions, ok := filters["parking_options"].([]string); ok && len(parkingOptions) > 0 {
		query += fmt.Sprintf(" AND parking_options && $%d", paramIdx)
		args = append(args, pq.StringArray(parkingOptions))
		paramIdx++
	}

	// WiFi speed min filter
	if wifiSpeedMin, ok := filters["wifi_speed_min"].(int); ok {
		query += fmt.Sprintf(" AND wifi_speed_mbps >= $%d", paramIdx)
		args = append(args, wifiSpeedMin)
		paramIdx++
	}

	// Radius filter (requires lat/lng)
	radius, hasRadius := filters["radius"].(float64)
	if hasRadius && hasLat && hasLng {
		query += fmt.Sprintf(" AND ST_DWithin(location, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)", paramIdx, paramIdx+1, paramIdx+2)
		args = append(args, lng, lat, radius)
		paramIdx += 3
	}

	// Dynamic ordering based on sort_by
	switch sortBy {
	case "nearest":
		if needsDistance {
			query += " ORDER BY distance_meters ASC"
		} else {
			query += " ORDER BY created_at DESC" // Fallback if no location provided
		}
	case "top_rated":
		// TODO: Add rating column when available, for now fallback to created_at
		query += " ORDER BY created_at DESC"
	default: // "recommended" or empty
		query += " ORDER BY created_at DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &pois, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search pois: %w", err)
	}

	return pois, nil
}

// GetByID retrieves a POI by its ID
func (r *POIRepository) GetByID(ctx context.Context, poiID uuid.UUID) (*POI, error) {
	var poi POI
	query := `
		SELECT poi_id, points_of_interest.name, category_id, points_of_interest.website, brand, description,
		       points_of_interest.address_id, parking_info, amenities, has_wifi, outdoor_seating,
		       is_wheelchair_accessible, has_delivery, cuisine, price_range,
		       food_options, payment_options, kids_friendly, smoker_friendly,
		       pet_friendly, points_of_interest.status, is_verified, verified_at, points_of_interest.created_at, points_of_interest.updated_at, points_of_interest.created_by,
		       floor_unit, public_transport, cover_image_url, gallery_image_urls,
		       (
		           SELECT COALESCE(json_agg(
		               json_build_object(
		                   'photo_id', ph.photo_id,
		                   'poi_id', ph.poi_id,
		                   'url', ph.url,
		                   'is_hero', ph.is_hero,
		                   'score', ph.score,
		                   'upvotes', ph.upvotes,
		                   'downvotes', ph.downvotes,
		                   'is_pinned', ph.is_pinned,
		                   'is_admin_official', ph.is_admin_official,
		                   'created_at', ph.created_at
		               ) ORDER BY ph.is_pinned DESC, ph.is_hero DESC, ph.score DESC
		           ), '[]'::json)
		           FROM photos ph
		           WHERE ph.poi_id = points_of_interest.poi_id
		       ) as gallery_images,
		       wifi_quality, power_outlets, seating_options, noise_level, has_ac,
		       vibes, crowd_type, lighting, music_type, cleanliness, dietary_options,
		       featured_menu_items, specials, open_hours, reservation_required,
		       reservation_platform, wait_time_estimate, happy_hour_info, loyalty_program,
		       points_of_interest.phone, points_of_interest.email, social_media_links, category_ids, parking_options, pet_policy,
		       founding_user_id, wifi_speed_mbps, wifi_verified_at, ergonomic_seating, power_sockets_reach,
		       ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude,
		       (
		           SELECT array_agg(name_key)
		           FROM categories
		           WHERE category_id = points_of_interest.category_id
		              OR category_id::text = ANY(points_of_interest.category_ids)
		       ) as category_names,
		       a.street_address as address,
		       u.name as founding_user_username,
		       COALESCE(
		           (SELECT AVG(rating)::float8 FROM reviews r WHERE r.poi_id = points_of_interest.poi_id),
		           0
		       ) as rating_avg,
		       (SELECT COUNT(*)::int FROM reviews r WHERE r.poi_id = points_of_interest.poi_id) as reviews_count
		FROM points_of_interest
		LEFT JOIN addresses a ON points_of_interest.address_id = a.address_id
		LEFT JOIN users u ON COALESCE(points_of_interest.founding_user_id, points_of_interest.created_by) = u.user_id
		WHERE poi_id = $1
	`

	err := r.db.GetContext(ctx, &poi, query, poiID)
	if err != nil {
		return nil, fmt.Errorf("get poi by id: %w", err)
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
			(
			   SELECT COALESCE(json_agg(
				   json_build_object(
					   'photo_id', ph.photo_id,
					   'poi_id', ph.poi_id,
					   'url', ph.url,
					   'is_hero', ph.is_hero,
					   'score', ph.score,
					   'upvotes', ph.upvotes,
					   'downvotes', ph.downvotes,
					   'is_pinned', ph.is_pinned,
					   'is_admin_official', ph.is_admin_official,
					   'created_at', ph.created_at
				   ) ORDER BY ph.is_pinned DESC, ph.is_hero DESC, ph.score DESC
			   ), '[]'::json)
			   FROM photos ph
			   WHERE ph.poi_id = points_of_interest.poi_id
			) as gallery_images,
			founding_user_id, wifi_speed_mbps, wifi_verified_at, ergonomic_seating, power_sockets_reach,
			ST_Y(location::geometry) as latitude,
			ST_X(location::geometry) as longitude,
			ST_Distance(
				location,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
			) as distance_meters,
			u.name as founding_user_username,
			COALESCE(
				(SELECT AVG(rating)::float8 FROM reviews r WHERE r.poi_id = points_of_interest.poi_id),
				0
			) as rating_avg,
			(SELECT COUNT(*)::int FROM reviews r WHERE r.poi_id = points_of_interest.poi_id) as reviews_count
		FROM points_of_interest
		LEFT JOIN users u ON COALESCE(points_of_interest.founding_user_id, points_of_interest.created_by) = u.user_id
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
		return nil, fmt.Errorf("get nearby pois: %w", err)
	}

	return pois, nil
}

// GetByUser retrieves all POIs created by a specific user
func (r *POIRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]POI, int, error) {
	var pois []POI
	query := `
		SELECT poi_id, name, category_id, description, status, created_by,
		       cover_image_url, has_wifi, outdoor_seating, price_range, created_at, updated_at,
			   (
			   SELECT COALESCE(json_agg(
				   json_build_object(
					   'photo_id', ph.photo_id,
					   'poi_id', ph.poi_id,
					   'url', ph.url,
					   'is_hero', ph.is_hero,
					   'score', ph.score,
					   'upvotes', ph.upvotes,
					   'downvotes', ph.downvotes,
					   'is_pinned', ph.is_pinned,
					   'is_admin_official', ph.is_admin_official,
					   'created_at', ph.created_at
				   ) ORDER BY ph.is_pinned DESC, ph.is_hero DESC, ph.score DESC
			   ), '[]'::json)
			   FROM photos ph
			   WHERE ph.poi_id = points_of_interest.poi_id
			) as gallery_images
		FROM points_of_interest
		WHERE created_by = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &pois, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get pois by user: %w", err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM points_of_interest WHERE created_by = $1`
	err = r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return pois, 0, fmt.Errorf("count pois by user: %w", err)
	}

	return pois, total, nil
}

// GetWithHeroImages retrieves POIs from the materialized view
func (r *POIRepository) GetWithHeroImages(ctx context.Context, limit, offset int) ([]map[string]interface{}, error) {
	pois := make([]map[string]interface{}, 0, limit)

	query := `
		SELECT * FROM mv_pois_with_hero
		ORDER BY rating_avg DESC, reviews_count DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query with hero images: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		poi := make(map[string]interface{})
		if err := rows.MapScan(poi); err != nil {
			return nil, fmt.Errorf("map scan with hero images: %w", err)
		}
		pois = append(pois, poi)
	}

	return pois, nil
}
