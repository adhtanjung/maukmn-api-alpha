package handlers

import (
	"net/http"

	"maukemana-backend/internal/repositories"
	"maukemana-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PhotoHandler struct {
	repo *repositories.PhotoRepository
}

func NewPhotoHandler(repo *repositories.PhotoRepository) *PhotoHandler {
	return &PhotoHandler{repo: repo}
}

// VotePhoto handles upvoting/downvoting a photo with Reddit-style toggle
func (h *PhotoHandler) VotePhoto(c *gin.Context) {
	photoIDStr := c.Param("photo_id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid photo ID", err)
		return
	}

	// Extract user_id from auth context (set by AuthMiddleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		// Try string conversion
		userIDStr, ok := userIDVal.(string)
		if !ok || userIDStr == "" {
			utils.SendError(c, http.StatusInternalServerError, "Invalid user ID type", nil)
			return
		}
		var err error
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "Invalid user ID", err)
			return
		}
	}

	var input struct {
		VoteType string `json:"vote_type" binding:"required,oneof=up down"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Convert vote_type string to int (-1 or 1)
	voteInt := 1
	if input.VoteType == "down" {
		voteInt = -1
	}

	newScore, userVote, err := h.repo.VoteWithToggle(c.Request.Context(), photoID, userID, voteInt)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to register vote", err)
		return
	}

	utils.SendSuccess(c, "Vote registered", gin.H{
		"photo_id":  photoID,
		"new_score": newScore,
		"user_vote": userVote, // 1=upvoted, -1=downvoted, 0=no vote
	})
}
