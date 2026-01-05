package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"maukemana-backend/internal/repositories"
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Build filters from query params
	filters := make(map[string]interface{})

	if category := c.Query("category_id"); category != "" {
		if catID, err := uuid.Parse(category); err == nil {
			filters["category_id"] = catID
		}
	}

	if hasWifi := c.Query("has_wifi"); hasWifi == "true" {
		filters["has_wifi"] = true
	}

	if priceRange := c.Query("price_range"); priceRange != "" {
		if pr, err := strconv.Atoi(priceRange); err == nil {
			filters["price_range"] = pr
		}
	}

	pois, err := h.repo.Search(ctx, filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   pois,
		"count":  len(pois),
		"limit":  limit,
		"offset": offset,
	})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "POI not found"})
		return
	}

	c.JSON(http.StatusOK, poi)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "POI created successfully",
		"data":    poi,
	})
}

// UpdatePOIRequest represents the JSON input for updating a POI
type UpdatePOIRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	HasWifi        *bool    `json:"has_wifi"`
	OutdoorSeating *bool    `json:"outdoor_seating"`
	PriceRange     *int     `json:"price_range"`
	Amenities      []string `json:"amenities"`
}

// UpdatePOI handles PUT /api/v1/pois/:id
func (h *POIHandler) UpdatePOI(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	poiID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid POI ID format"})
		return
	}

	var input UpdatePOIRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.Update(ctx, poiID, repositories.UpdatePOIInput{
		Name:           input.Name,
		Description:    input.Description,
		HasWifi:        input.HasWifi,
		OutdoorSeating: input.OutdoorSeating,
		PriceRange:     input.PriceRange,
		Amenities:      input.Amenities,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "POI updated successfully",
		"poi_id":  poiID,
	})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "POI deleted successfully"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   pois,
		"count":  len(pois),
		"center": gin.H{"lat": lat, "lng": lng},
		"radius": radius,
	})
}

// GetFilterOptions handles GET /api/v1/pois/filter-options
func (h *POIHandler) GetFilterOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"price_ranges": []gin.H{
			{"value": 1, "label": "$"},
			{"value": 2, "label": "$$"},
			{"value": 3, "label": "$$$"},
			{"value": 4, "label": "$$$$"},
		},
		"amenities": []string{
			"wifi", "outdoor_seating", "power_outlets", "parking",
			"wheelchair_accessible", "pet_friendly", "delivery",
		},
		"food_options": []string{
			"vegan", "vegetarian", "halal", "gluten_free",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid POI ID format"})
		return
	}

	// Verify ownership
	poi, err := h.repo.GetByID(ctx, poiID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "POI not found"})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if poi.CreatedBy == nil || *poi.CreatedBy != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to submit this POI"})
		return
	}

	if poi.Status != "draft" && poi.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only submit draft or rejected POIs"})
		return
	}

	if err := h.repo.UpdateStatus(ctx, poiID, "pending", nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "POI submitted for review"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "POI approved"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "POI rejected"})
}

// GetMyDrafts handles GET /api/v1/pois/my-drafts
func (h *POIHandler) GetMyDrafts(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	pois, err := h.repo.GetByUserAndStatus(ctx, userID.(uuid.UUID), "draft", limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pois, "count": len(pois)})
}

// GetPendingPOIs handles GET /api/v1/pois/pending (admin only)
func (h *POIHandler) GetPendingPOIs(c *gin.Context) {
	ctx := c.Request.Context()

	role, exists := c.Get("user_role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	pois, err := h.repo.GetByStatus(ctx, "pending", limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pois, "count": len(pois)})
}
