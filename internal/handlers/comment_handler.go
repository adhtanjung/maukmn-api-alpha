package handlers

import (
	"errors"
	"maukemana-backend/internal/models"
	"maukemana-backend/internal/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CommentHandler struct {
	commentRepo *repositories.CommentRepository
}

func NewCommentHandler(commentRepo *repositories.CommentRepository) *CommentHandler {
	return &CommentHandler{commentRepo: commentRepo}
}

// Helper to get user ID from context
func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}
	// It might be stored as string or UUID depending on middleware
	// AuthMiddleware in this project stores it as uuid.UUID directly based on my read
	// But let's handle both just in case or check auth_middleware again.
	// Re-reading auth_middleware.go: c.Set("user_id", userID) where userID is uuid.UUID.
	if id, ok := userIDStr.(uuid.UUID); ok {
		return id, nil
	}
	// Fallback if it's a string
	if idStr, ok := userIDStr.(string); ok {
		return uuid.Parse(idStr)
	}
	return uuid.Nil, errors.New("invalid user_id type")
}

// CreateInput defines the expected JSON payload for creating a comment
type CreateCommentInput struct {
	Content  string     `json:"content" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	poiIDStr := c.Param("id")
	poiID, err := uuid.Parse(poiIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid POI ID"})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := &models.Comment{
		PoiID:    poiID,
		UserID:   userID,
		Content:  input.Content,
		ParentID: input.ParentID,
	}

	if err := h.commentRepo.Create(c.Request.Context(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) GetCommentsByPOI(c *gin.Context) {
	poiIDStr := c.Param("id")
	poiID, err := uuid.Parse(poiIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid POI ID"})
		return
	}

	limit := 50
	offset := 0

	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	comments, err := h.commentRepo.GetByPOI(c.Request.Context(), poiID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Comment ID"})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err = h.commentRepo.Delete(c.Request.Context(), commentID, userID)
	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not found or permission denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
