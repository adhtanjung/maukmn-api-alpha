package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"maukemana-backend/internal/storage"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	r2 *storage.R2Client
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(r2 *storage.R2Client) *UploadHandler {
	return &UploadHandler{r2: r2}
}

// PresignRequest represents the request for a presigned URL
type PresignRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Category    string `json:"category"` // "cover", "gallery", "profile"
}

// PresignResponse contains the presigned URL and final public URL
type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	Key       string `json:"key"`
}

// GetPresignedURL generates a presigned URL for direct upload to R2
func (h *UploadHandler) GetPresignedURL(c *gin.Context) {
	ctx := c.Request.Context()

	var req PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate content type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	if !allowedTypes[req.ContentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content type, must be image/jpeg, image/png, image/webp, or image/gif"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Generate unique key with path structure
	// Format: uploads/{user_id}/{category}/{timestamp}_{uuid}.{ext}
	ext := filepath.Ext(req.Filename)
	if ext == "" {
		ext = ".webp" // Default to webp for compressed images
	}

	category := req.Category
	if category == "" {
		category = "general"
	}

	key := fmt.Sprintf("uploads/%s/%s/%d_%s%s",
		userID.(uuid.UUID).String(),
		category,
		time.Now().Unix(),
		uuid.New().String()[:8],
		ext,
	)

	// Generate presigned URL
	uploadURL, err := h.r2.GeneratePresignedURL(ctx, key, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload URL"})
		return
	}

	publicURL := h.r2.GetPublicURL(key)

	c.JSON(http.StatusOK, PresignResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
		Key:       key,
	})
}

// DeleteUpload removes a file from R2
func (h *UploadHandler) DeleteUpload(c *gin.Context) {
	ctx := c.Request.Context()
	key := c.Query("key")

	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	// Verify user owns this file (key starts with their user ID)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	expectedPrefix := fmt.Sprintf("uploads/%s/", userID.(uuid.UUID).String())
	if !strings.HasPrefix(key, expectedPrefix) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this file"})
		return
	}

	if err := h.r2.DeleteObject(ctx, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
}
