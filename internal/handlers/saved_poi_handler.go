package handlers

import (
	"context"
	"maukemana-backend/internal/logger"
	"maukemana-backend/internal/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SavedPOIRepository interface {
	SavePOI(ctx context.Context, userID, poiID uuid.UUID) error
	UnsavePOI(ctx context.Context, userID, poiID uuid.UUID) error
	GetSavedPOIs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repositories.POI, error)
	IsSaved(ctx context.Context, userID, poiID uuid.UUID) (bool, error)
}

type SavedPOIHandler struct {
	repo SavedPOIRepository
}

func NewSavedPOIHandler(repo SavedPOIRepository) *SavedPOIHandler {
	return &SavedPOIHandler{repo: repo}
}

// ToggleSave handles POST /api/v1/pois/:id/save
// It checks if the POI is already saved. If yes, unsaves it. If no, saves it.
// Returns the new state { "is_saved": boolean }
func (h *SavedPOIHandler) ToggleSave(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// 2. Get POI ID from URL param
	poiIDStr := c.Param("id")
	poiID, err := uuid.Parse(poiIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid POI ID"})
		return
	}

	// 3. Check current status
	isSaved, err := h.repo.IsSaved(c.Request.Context(), userID, poiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 4. Toggle
	if isSaved {
		err = h.repo.UnsavePOI(c.Request.Context(), userID, poiID)
		if err != nil {
			logger.L().Error("Failed to unsave POI", "error", err, "user_id", userID, "poi_id", poiID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsave POI"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"is_saved": false, "message": "POI unsaved"})
	} else {
		err = h.repo.SavePOI(c.Request.Context(), userID, poiID)
		if err != nil {
			logger.L().Error("Failed to save POI", "error", err, "user_id", userID, "poi_id", poiID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save POI"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"is_saved": true, "message": "POI saved"})
	}
}

// GetMySavedPOIs handles GET /api/v1/me/saved-pois
func (h *SavedPOIHandler) GetMySavedPOIs(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	pois, err := h.repo.GetSavedPOIs(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.L().Error("Failed to fetch saved POIs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch saved POIs"})
		return
	}

	// Return empty array instead of null for empty results
	if pois == nil {
		pois = []repositories.POI{}
	}

	c.JSON(http.StatusOK, gin.H{"pois": pois})
}
