package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"maukemana-backend/internal/repositories"
	"maukemana-backend/internal/utils"
)

// POIHandler handles POI-related HTTP requests
type POIHandler struct {
	repo *repositories.POIRepository
}

// NewPOIHandler creates a new POI handler
func NewPOIHandler(repo *repositories.POIRepository) *POIHandler {
	return &POIHandler{repo: repo}
}

// SearchPOIs handles GET /api/v1/pois
func (h *POIHandler) SearchPOIs(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, limit := utils.GetPagination(c)
	offset := utils.GetOffset(page, limit)

	// Build filters from query params
	filters := make(map[string]interface{})

	// Category filter
	if category := c.Query("category_id"); category != "" {
		if catID, err := uuid.Parse(category); err == nil {
			filters["category_id"] = catID
		}
	}

	// Legacy has_wifi boolean filter
	if hasWifi := c.Query("has_wifi"); hasWifi == "true" {
		filters["has_wifi"] = true
	}

	// Price range filter
	if priceRange := c.Query("price_range"); priceRange != "" {
		if pr, err := strconv.Atoi(priceRange); err == nil {
			filters["price_range"] = pr
		}
	}

	// Status filter - defaults to "approved" for public feed
	status := c.Query("status")
	if status == "" {
		status = "approved"
	}
	filters["status"] = status

	// WiFi quality filter (string: none|slow|moderate|fast|excellent)
	if wifiQuality := c.Query("wifi_quality"); wifiQuality != "" {
		filters["wifi_quality"] = wifiQuality
	}

	// Noise level filter (string: silent|quiet|moderate|lively|loud)
	if noiseLevel := c.Query("noise_level"); noiseLevel != "" {
		filters["noise_level"] = noiseLevel
	}

	// Power outlets filter (string: none|limited|moderate|plenty)
	if powerOutlets := c.Query("power_outlets"); powerOutlets != "" {
		filters["power_outlets"] = powerOutlets
	}

	// Cuisine filter (string)
	if cuisine := c.Query("cuisine"); cuisine != "" {
		filters["cuisine"] = cuisine
	}

	// Has AC filter (boolean)
	if hasAC := c.Query("has_ac"); hasAC == "true" {
		filters["has_ac"] = true
	} else if hasAC == "false" {
		filters["has_ac"] = false
	}

	// Vibes filter (comma-separated array)
	if vibes := c.Query("vibes"); vibes != "" {
		filters["vibes"] = parseCommaSeparated(vibes)
	}

	// Crowd type filter (comma-separated array)
	if crowdType := c.Query("crowd_type"); crowdType != "" {
		filters["crowd_type"] = parseCommaSeparated(crowdType)
	}

	// Dietary options filter (comma-separated array)
	if dietaryOptions := c.Query("dietary_options"); dietaryOptions != "" {
		filters["dietary_options"] = parseCommaSeparated(dietaryOptions)
	}

	// Seating options filter (comma-separated array)
	if seatingOptions := c.Query("seating_options"); seatingOptions != "" {
		filters["seating_options"] = parseCommaSeparated(seatingOptions)
	}

	// Parking options filter (comma-separated array)
	if parkingOptions := c.Query("parking_options"); parkingOptions != "" {
		filters["parking_options"] = parseCommaSeparated(parkingOptions)
	}

	// Sort by filter (string: recommended|nearest|top_rated)
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters["sort_by"] = sortBy

		// For "nearest" sorting, we need lat/lng
		if sortBy == "nearest" {
			if lat, err := strconv.ParseFloat(c.Query("lat"), 64); err == nil {
				filters["lat"] = lat
			}
			if lng, err := strconv.ParseFloat(c.Query("lng"), 64); err == nil {
				filters["lng"] = lng
			}
		}
	}

	pois, err := h.repo.Search(ctx, filters, limit, offset)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	// Note: We currently don't have a total count from the repo, so we use the slice length + offset as a proxy or just the length.
	// Ideally, the repo should return total count. For now, this standardizes the structure.
	utils.SendPaginated(c, "POIs retrieved successfully", pois, page, limit, len(pois)+offset)
}

// parseCommaSeparated splits a comma-separated string into a slice of strings
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetPOI handles GET /api/v1/pois/:id
func (h *POIHandler) GetPOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid POI ID format"})
		return
	}

	poi, err := h.repo.GetByID(ctx, poiID)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "POI not found", err)
		return
	}

	utils.SendSuccess(c, "POI details retrieved", poi)
}

// CreatePOIRequest represents the JSON input for creating a POI
type CreatePOIRequest struct {
	// Profile & Visuals
	Name             string   `json:"name" binding:"required"`
	BrandName        *string  `json:"brand_name"`
	Categories       []string `json:"categories"`
	Description      *string  `json:"description"`
	CoverImageURL    *string  `json:"cover_image_url"`
	GalleryImageURLs []string `json:"gallery_image_urls"`
	// Location
	Address              *string  `json:"address"`
	FloorUnit            *string  `json:"floor_unit"`
	Latitude             float64  `json:"latitude"`
	Longitude            float64  `json:"longitude"`
	PublicTransport      *string  `json:"public_transport"`
	ParkingOptions       []string `json:"parking_options"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
	// Work & Prod
	WifiQuality    *string  `json:"wifi_quality"`
	PowerOutlets   *string  `json:"power_outlets"`
	SeatingOptions []string `json:"seating_options"`
	NoiseLevel     *string  `json:"noise_level"`
	HasAC          bool     `json:"has_ac"`
	// Atmosphere
	Vibes       []string `json:"vibes"`
	CrowdType   []string `json:"crowd_type"`
	Lighting    *string  `json:"lighting"`
	MusicType   *string  `json:"music_type"`
	Cleanliness *string  `json:"cleanliness"`
	// Food & Drink
	Cuisine        *string  `json:"cuisine"`
	PriceRange     *int     `json:"price_range"`
	DietaryOptions []string `json:"dietary_options"`
	FeaturedItems  []string `json:"featured_items"`
	Specials       []string `json:"specials"`
	// Operations
	OpenHours           map[string]interface{} `json:"open_hours"`
	ReservationRequired bool                   `json:"reservation_required"`
	ReservationPlatform *string                `json:"reservation_platform"`
	PaymentOptions      []string               `json:"payment_options"`
	WaitTimeEstimate    *int                   `json:"wait_time_estimate"`
	// Social & Lifestyle
	KidsFriendly   bool     `json:"kids_friendly"`
	PetFriendly    []string `json:"pet_friendly"`
	SmokerFriendly bool     `json:"smoker_friendly"`
	HappyHourInfo  *string  `json:"happy_hour_info"`
	LoyaltyProgram *string  `json:"loyalty_program"`
	// Contact
	Phone       *string                `json:"phone"`
	Email       *string                `json:"email"`
	Website     *string                `json:"website"`
	SocialLinks map[string]interface{} `json:"social_links"`
}

// CreatePOI handles POST /api/v1/pois
func (h *POIHandler) CreatePOI(c *gin.Context) {
	ctx := c.Request.Context()

	var input CreatePOIRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	var createdBy *uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			createdBy = &uid
		}
	}

	poi, err := h.repo.Create(ctx, repositories.CreatePOIInput{
		// Profile & Visuals
		Name:             input.Name,
		BrandName:        input.BrandName,
		Categories:       input.Categories,
		Description:      input.Description,
		CoverImageURL:    input.CoverImageURL,
		GalleryImageURLs: input.GalleryImageURLs,
		// Location
		Address:              input.Address,
		FloorUnit:            input.FloorUnit,
		Latitude:             input.Latitude,
		Longitude:            input.Longitude,
		PublicTransport:      input.PublicTransport,
		ParkingOptions:       input.ParkingOptions,
		WheelchairAccessible: input.WheelchairAccessible,
		// Work & Prod
		WifiQuality:    input.WifiQuality,
		PowerOutlets:   input.PowerOutlets,
		SeatingOptions: input.SeatingOptions,
		NoiseLevel:     input.NoiseLevel,
		HasAC:          input.HasAC,
		// Atmosphere
		Vibes:       input.Vibes,
		CrowdType:   input.CrowdType,
		Lighting:    input.Lighting,
		MusicType:   input.MusicType,
		Cleanliness: input.Cleanliness,
		// Food & Drink
		Cuisine:        input.Cuisine,
		PriceRange:     input.PriceRange,
		DietaryOptions: input.DietaryOptions,
		FeaturedItems:  input.FeaturedItems,
		Specials:       input.Specials,
		// Operations
		OpenHours:           input.OpenHours,
		ReservationRequired: input.ReservationRequired,
		ReservationPlatform: input.ReservationPlatform,
		PaymentOptions:      input.PaymentOptions,
		WaitTimeEstimate:    input.WaitTimeEstimate,
		// Social & Lifestyle
		KidsFriendly:   input.KidsFriendly,
		PetFriendly:    input.PetFriendly,
		SmokerFriendly: input.SmokerFriendly,
		HappyHourInfo:  input.HappyHourInfo,
		LoyaltyProgram: input.LoyaltyProgram,
		// Contact
		Phone:       input.Phone,
		Email:       input.Email,
		Website:     input.Website,
		SocialLinks: input.SocialLinks,
		// Metadata
		CreatedBy: createdBy,
	})
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendCreated(c, "POI created successfully", poi)
}

// UpdatePOIRequest represents the JSON input for updating a POI (full update)
type UpdatePOIRequest struct {
	// Profile & Visuals
	Name             string   `json:"name" binding:"required"`
	BrandName        *string  `json:"brand_name"`
	Categories       []string `json:"categories"`
	Description      *string  `json:"description"`
	CoverImageURL    *string  `json:"cover_image_url"`
	GalleryImageURLs []string `json:"gallery_image_urls"`
	// Location
	Address              *string  `json:"address"`
	FloorUnit            *string  `json:"floor_unit"`
	Latitude             float64  `json:"latitude"`
	Longitude            float64  `json:"longitude"`
	PublicTransport      *string  `json:"public_transport"`
	ParkingOptions       []string `json:"parking_options"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
	// Work & Prod
	WifiQuality    *string  `json:"wifi_quality"`
	PowerOutlets   *string  `json:"power_outlets"`
	SeatingOptions []string `json:"seating_options"`
	NoiseLevel     *string  `json:"noise_level"`
	HasAC          bool     `json:"has_ac"`
	// Atmosphere
	Vibes       []string `json:"vibes"`
	CrowdType   []string `json:"crowd_type"`
	Lighting    *string  `json:"lighting"`
	MusicType   *string  `json:"music_type"`
	Cleanliness *string  `json:"cleanliness"`
	// Food & Drink
	Cuisine        *string  `json:"cuisine"`
	PriceRange     *int     `json:"price_range"`
	DietaryOptions []string `json:"dietary_options"`
	FeaturedItems  []string `json:"featured_items"`
	Specials       []string `json:"specials"`
	// Operations
	OpenHours           map[string]interface{} `json:"open_hours"`
	ReservationRequired bool                   `json:"reservation_required"`
	ReservationPlatform *string                `json:"reservation_platform"`
	PaymentOptions      []string               `json:"payment_options"`
	WaitTimeEstimate    *int                   `json:"wait_time_estimate"`
	// Social & Lifestyle
	KidsFriendly   bool     `json:"kids_friendly"`
	PetFriendly    []string `json:"pet_friendly"`
	SmokerFriendly bool     `json:"smoker_friendly"`
	HappyHourInfo  *string  `json:"happy_hour_info"`
	LoyaltyProgram *string  `json:"loyalty_program"`
	// Contact
	Phone       *string                `json:"phone"`
	Email       *string                `json:"email"`
	Website     *string                `json:"website"`
	SocialLinks map[string]interface{} `json:"social_links"`
}

// UpdatePOI handles PUT /api/v1/pois/:id
// Authorized for: POI owner OR admin
func (h *POIHandler) UpdatePOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid POI ID format", err)
		return
	}

	// Get the POI to check ownership
	poi, err := h.repo.GetByID(ctx, poiID)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "POI not found", err)
		return
	}

	// Get user info from context
	userID, userIDExists := c.Get("user_id")
	role, _ := c.Get("user_role")
	isAdmin := role == "admin"

	// Authorization check: owner or admin
	// Special case: if created_by is NULL (orphan POI), allow any authenticated user to claim it
	isOwner := false
	if poi.CreatedBy != nil && userIDExists {
		isOwner = *poi.CreatedBy == userID.(uuid.UUID)
	}

	// If POI has no owner (created_by is null), allow any authenticated user to edit
	// This handles legacy POIs that were created before ownership tracking
	isOrphanPOI := poi.CreatedBy == nil

	if !isOwner && !isAdmin && !isOrphanPOI {
		utils.SendError(c, http.StatusForbidden, "not authorized to edit this POI", nil)
		return
	}

	// TODO: Optionally claim ownership of orphan POIs by updating created_by
	// if isOrphanPOI && userIDExists {
	//     h.repo.SetOwner(ctx, poiID, userID.(uuid.UUID))
	// }

	var input UpdatePOIRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err = h.repo.UpdateFull(ctx, poiID, repositories.UpdateFullInput{
		Name:                 input.Name,
		BrandName:            input.BrandName,
		Categories:           input.Categories,
		Description:          input.Description,
		CoverImageURL:        input.CoverImageURL,
		GalleryImageURLs:     input.GalleryImageURLs,
		Address:              input.Address,
		FloorUnit:            input.FloorUnit,
		Latitude:             input.Latitude,
		Longitude:            input.Longitude,
		PublicTransport:      input.PublicTransport,
		ParkingOptions:       input.ParkingOptions,
		WheelchairAccessible: input.WheelchairAccessible,
		WifiQuality:          input.WifiQuality,
		PowerOutlets:         input.PowerOutlets,
		SeatingOptions:       input.SeatingOptions,
		NoiseLevel:           input.NoiseLevel,
		HasAC:                input.HasAC,
		Vibes:                input.Vibes,
		CrowdType:            input.CrowdType,
		Lighting:             input.Lighting,
		MusicType:            input.MusicType,
		Cleanliness:          input.Cleanliness,
		Cuisine:              input.Cuisine,
		PriceRange:           input.PriceRange,
		DietaryOptions:       input.DietaryOptions,
		FeaturedItems:        input.FeaturedItems,
		Specials:             input.Specials,
		OpenHours:            input.OpenHours,
		ReservationRequired:  input.ReservationRequired,
		ReservationPlatform:  input.ReservationPlatform,
		PaymentOptions:       input.PaymentOptions,
		WaitTimeEstimate:     input.WaitTimeEstimate,
		KidsFriendly:         input.KidsFriendly,
		PetFriendly:          input.PetFriendly,
		SmokerFriendly:       input.SmokerFriendly,
		HappyHourInfo:        input.HappyHourInfo,
		LoyaltyProgram:       input.LoyaltyProgram,
		Phone:                input.Phone,
		Email:                input.Email,
		Website:              input.Website,
		SocialLinks:          input.SocialLinks,
	})
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "POI updated successfully", gin.H{"poi_id": poiID})
}

// DeletePOI handles DELETE /api/v1/pois/:id
func (h *POIHandler) DeletePOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid POI ID format"})
		return
	}

	if err := h.repo.Delete(ctx, poiID); err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "POI deleted successfully", nil)
}

// GetMyPOIs handles GET /api/v1/pois/my
// Returns all POIs created by the authenticated user
func (h *POIHandler) GetMyPOIs(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	page, limit := utils.GetPagination(c)
	offset := utils.GetOffset(page, limit)

	pois, total, err := h.repo.GetByUser(ctx, userID.(uuid.UUID), limit, offset)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendPaginated(c, "User POIs retrieved", pois, page, limit, total)
}

// GetNearbyPOIs handles GET /api/v1/pois/nearby
func (h *POIHandler) GetNearbyPOIs(c *gin.Context) {
	ctx := c.Request.Context()

	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	radius, _ := strconv.Atoi(c.DefaultQuery("radius", "5000"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	pois, err := h.repo.GetNearby(ctx, lat, lng, radius, limit)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "Nearby POIs retrieved", gin.H{
		"data":   pois,
		"count":  len(pois),
		"center": gin.H{"lat": lat, "lng": lng},
		"radius": radius,
	})
}

// GetFilterOptions handles GET /api/v1/pois/filter-options
func (h *POIHandler) GetFilterOptions(c *gin.Context) {
	utils.SendSuccess(c, "Filter options retrieved", gin.H{
		"sort_options": []gin.H{
			{"value": "recommended", "label": "Recommended"},
			{"value": "nearest", "label": "Nearest"},
			{"value": "top_rated", "label": "Top Rated"},
		},
		"price_ranges": []gin.H{
			{"value": 1, "label": "$"},
			{"value": 2, "label": "$$"},
			{"value": 3, "label": "$$$"},
			{"value": 4, "label": "$$$$"},
		},
		"wifi_quality": []gin.H{
			{"value": "any", "label": "Any"},
			{"value": "slow", "label": "Slow"},
			{"value": "moderate", "label": "Mid"},
			{"value": "fast", "label": "Fast"},
			{"value": "excellent", "label": "Best"},
		},
		"noise_levels": []gin.H{
			{"value": "silent", "label": "Silent"},
			{"value": "quiet", "label": "Quiet"},
			{"value": "moderate", "label": "Mid"},
			{"value": "lively", "label": "Lively"},
			{"value": "loud", "label": "Loud"},
		},
		"power_outlets": []gin.H{
			{"value": "any", "label": "Any"},
			{"value": "limited", "label": "Low"},
			{"value": "moderate", "label": "Mid"},
			{"value": "plenty", "label": "Many"},
		},
		"vibes": []gin.H{
			{"value": "industrial", "label": "Industrial", "icon": "factory"},
			{"value": "cozy", "label": "Cozy", "icon": "chair"},
			{"value": "tropical", "label": "Tropical", "icon": "potted_plant"},
			{"value": "minimalist", "label": "Minimalist", "icon": "crop_square"},
			{"value": "luxury", "label": "Luxury", "icon": "diamond"},
			{"value": "retro", "label": "Retro", "icon": "radio"},
			{"value": "nature", "label": "Nature", "icon": "park"},
		},
		"crowd_types": []gin.H{
			{"value": "quiet_study", "label": "Quiet / Study"},
			{"value": "social_lively", "label": "Social / Lively"},
			{"value": "business", "label": "Business"},
		},
		"dietary_options": []gin.H{
			{"value": "vegan", "label": "Vegan"},
			{"value": "vegetarian", "label": "Vegetarian"},
			{"value": "halal", "label": "Halal"},
			{"value": "gluten_free", "label": "Gluten-Free"},
			{"value": "nut_free", "label": "Nut-Free"},
		},
		"seating_options": []gin.H{
			{"value": "ergonomic", "label": "Ergonomic", "icon": "chair"},
			{"value": "communal", "label": "Communal", "icon": "table_restaurant"},
			{"value": "high-tops", "label": "High-tops", "icon": "countertops"},
			{"value": "outdoor", "label": "Outdoor", "icon": "deck"},
			{"value": "private-booths", "label": "Private Booths", "icon": "meeting_room"},
		},
		"parking_options": []gin.H{
			{"value": "car", "label": "Car Parking"},
			{"value": "motorcycle", "label": "Motorcycle"},
			{"value": "valet", "label": "Valet Service"},
		},
		"cuisines": []gin.H{
			{"value": "italian", "label": "Italian"},
			{"value": "japanese", "label": "Japanese"},
			{"value": "mexican", "label": "Mexican"},
			{"value": "fusion", "label": "Fusion"},
			{"value": "cafe", "label": "Cafe"},
		},
		"quick_filters": []gin.H{
			{
				"id":    "deep_work",
				"label": "Deep Work",
				"icon":  "rocket_launch",
				"filters": gin.H{
					"wifi_quality":  "fast",
					"noise_level":   "quiet",
					"power_outlets": "plenty",
				},
			},
			{
				"id":    "client_meeting",
				"label": "Client Meeting",
				"icon":  "handshake",
				"filters": gin.H{
					"noise_level": "moderate",
					"vibes":       []string{"luxury", "minimalist"},
				},
			},
			{
				"id":    "date_night",
				"label": "Date Night",
				"icon":  "wine_bar",
				"filters": gin.H{
					"vibes": []string{"cozy", "luxury"},
				},
			},
		},
		"timestamp": time.Now().Unix(),
	})
}

// SubmitPOI handles POST /api/v1/pois/:id/submit
func (h *POIHandler) SubmitPOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid POI ID format", err)
		return
	}

	// Verify ownership
	poi, err := h.repo.GetByID(ctx, poiID)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "POI not found", err)
		return
	}

	// Get user from context
	_, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Ownership check removed as per requirement "anyone can submit POI"
	// if poi.CreatedBy == nil || *poi.CreatedBy != userID.(uuid.UUID) {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to submit this POI"})
	// 	return
	// }

	// Allow draft, rejected, approved, and pending POIs to be submitted/re-submitted
	if poi.Status != "draft" && poi.Status != "rejected" && poi.Status != "approved" && poi.Status != "pending" {
		utils.SendError(c, http.StatusBadRequest, "can only submit draft, rejected, pending, or approved POIs", nil)
		return
	}

	if err := h.repo.UpdateStatus(ctx, poiID, "pending", nil); err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "POI submitted for review", nil)
}

// ApprovePOI handles POST /api/v1/pois/:id/approve (admin only)
func (h *POIHandler) ApprovePOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid POI ID format"})
		return
	}

	// Check admin role from context
	role, exists := c.Get("user_role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	if err := h.repo.UpdateStatus(ctx, poiID, "approved", nil); err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "POI approved", nil)
}

// RejectPOIRequest for rejection reason
type RejectPOIRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RejectPOI handles POST /api/v1/pois/:id/reject (admin only)
func (h *POIHandler) RejectPOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid POI ID format"})
		return
	}

	// Check admin role
	role, exists := c.Get("user_role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var input RejectPOIRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateStatus(ctx, poiID, "rejected", &input.Reason); err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "POI rejected", nil)
}

// GetMyDrafts handles GET /api/v1/pois/my-drafts
func (h *POIHandler) GetMyDrafts(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, limit := utils.GetPagination(c)
	offset := utils.GetOffset(page, limit)

	pois, err := h.repo.GetByUserAndStatus(ctx, userID.(uuid.UUID), "draft", limit, offset)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendPaginated(c, "Drafts retrieved", pois, page, limit, len(pois)+offset)
}

// GetPendingPOIs handles GET /api/v1/pois/pending (admin only)
func (h *POIHandler) GetPendingPOIs(c *gin.Context) {
	ctx := c.Request.Context()

	role, exists := c.Get("user_role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	page, limit := utils.GetPagination(c)
	offset := utils.GetOffset(page, limit)

	pois, err := h.repo.GetByStatus(ctx, "pending", limit, offset)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendPaginated(c, "Pending POIs retrieved", pois, page, limit, len(pois)+offset)
}

// GetAdminPOIs handles GET /api/v1/pois/admin-list?status=... (admin only)
func (h *POIHandler) GetAdminPOIs(c *gin.Context) {
	ctx := c.Request.Context()

	role, exists := c.Get("user_role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	status := c.DefaultQuery("status", "pending")
	page, limit := utils.GetPagination(c)
	offset := utils.GetOffset(page, limit)

	pois, err := h.repo.GetByStatus(ctx, status, limit, offset)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendPaginated(c, "Admin POI list retrieved", pois, page, limit, len(pois)+offset)
}
