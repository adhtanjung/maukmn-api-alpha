package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"maukemana-backend/internal/imaging"
	"maukemana-backend/internal/storage"
	"maukemana-backend/internal/utils"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	r2             *storage.R2Client
	imagingService *imaging.Service
}

// NewUploadHandler creates a new upload handler
// NewUploadHandler creates a new upload handler
func NewUploadHandler(r2 *storage.R2Client, imagingService *imaging.Service) *UploadHandler {
	return &UploadHandler{
		r2:             r2,
		imagingService: imagingService,
	}
}

// PresignRequest represents the request for a presigned URL
type PresignRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Category    string `json:"category"` // "cover", "gallery", "profile", "general"
}

// PresignResponse contains the presigned URL and upload information
type PresignResponse struct {
	UploadID        string   `json:"upload_id"`
	UploadURL       string   `json:"upload_url"`
	UploadExpiresAt string   `json:"upload_expires_at"`
	MaxSizeBytes    int64    `json:"max_size_bytes"`
	AllowedTypes    []string `json:"allowed_content_types"`
	Key             string   `json:"key"`
	// Legacy fields for backward compatibility
	PublicURL string `json:"public_url,omitempty"`
}

// FinalizeRequest represents the request to finalize an upload
type FinalizeRequest struct {
	UploadKey string `json:"upload_key" binding:"required"`
	Category  string `json:"category"`
}

// FinalizeResponse contains the result of finalizing an upload
type FinalizeResponse struct {
	AssetID                    string `json:"asset_id"`
	ContentHash                string `json:"content_hash"`
	Status                     string `json:"status"`
	EstimatedCompletionSeconds int    `json:"estimated_completion_seconds"`
	StatusURL                  string `json:"status_url"`
}

// AssetStatusResponse contains the status and derivatives of an asset
type AssetStatusResponse struct {
	AssetID     string                     `json:"asset_id"`
	ContentHash string                     `json:"content_hash"`
	Status      string                     `json:"status"`
	Original    *OriginalInfo              `json:"original,omitempty"`
	Derivatives map[string]*DerivativeInfo `json:"derivatives,omitempty"`
	CreatedAt   string                     `json:"created_at"`
	ProcessedAt string                     `json:"processed_at,omitempty"`
	Error       string                     `json:"error,omitempty"`
}

// OriginalInfo contains information about the original image
type OriginalInfo struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	SizeBytes int64  `json:"size_bytes"`
}

// DerivativeInfo contains information about a derivative
type DerivativeInfo struct {
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Formats    []string `json:"formats"`
	URLPattern string   `json:"url_pattern"`
}

// GetPresignedURL generates a presigned URL for direct upload to R2
func (h *UploadHandler) GetPresignedURL(c *gin.Context) {
	ctx := c.Request.Context()

	var req PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	// Validate content type - expanded list
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
		"image/heic": true,
		"image/heif": true,
		"image/avif": true,
	}
	if !allowedTypes[req.ContentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid content type",
			"allowed": []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/heic", "image/avif"},
		})
		return
	}

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Generate unique upload key
	// Format: uploads/tmp/{user_id}/{timestamp}_{uuid}.{ext}
	ext := filepath.Ext(req.Filename)
	if ext == "" {
		// Infer extension from content type
		switch req.ContentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		case "image/heic", "image/heif":
			ext = ".heic"
		case "image/avif":
			ext = ".avif"
		default:
			ext = ".bin"
		}
	}

	category := req.Category
	if category == "" {
		category = "general"
	}

	uploadID := uuid.New()
	key := fmt.Sprintf("uploads/tmp/%s/%s/%d_%s%s",
		userID.String(),
		category,
		time.Now().Unix(),
		uploadID.String()[:8],
		ext,
	)

	// Get size limits for category
	limits := imaging.GetCategoryLimits(category)

	// Generate presigned URL
	uploadURL, err := h.r2.GeneratePresignedURL(ctx, key, req.ContentType)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	utils.SendSuccess(c, "Presigned URL generated", PresignResponse{
		UploadID:        uploadID.String(),
		UploadURL:       uploadURL,
		UploadExpiresAt: expiresAt.Format(time.RFC3339),
		MaxSizeBytes:    limits.MaxBytes,
		AllowedTypes:    []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/heic", "image/avif"},
		Key:             key,
		// Legacy: also include public_url for backward compatibility
		PublicURL: h.r2.GetPublicURL(key),
	})
}

// FinalizeUpload triggers async processing of an uploaded image
func (h *UploadHandler) FinalizeUpload(c *gin.Context) {
	var req FinalizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Verify the upload key belongs to this user
	expectedPrefix := fmt.Sprintf("uploads/tmp/%s/", userID.String())
	if !strings.HasPrefix(req.UploadKey, expectedPrefix) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized for this upload"})
		return
	}

	category := req.Category
	if category == "" {
		category = "general"
	}

	// Queue for async processing
	jobID, err := h.imagingService.QueueProcessing(req.UploadKey, category, userID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "processing queue is full, try again later"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": FinalizeResponse{
			AssetID:                    jobID.String(),
			Status:                     "processing",
			EstimatedCompletionSeconds: 5,
			StatusURL:                  fmt.Sprintf("/api/v1/assets/%s", jobID.String()),
		},
	})
}

// GetAssetStatus returns the processing status and derivatives of an asset
func (h *UploadHandler) GetAssetStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	var asset *imaging.ImageAsset
	var job *imaging.ProcessingJob
	var exists bool

	// 1. Try to find asset by ID
	asset, exists = h.imagingService.GetAssetByID(id)

	// 2. If not found, try to find job by ID
	if !exists {
		job, exists = h.imagingService.GetJobByID(id)
		if exists && job.AssetID != nil {
			// Job finished, check the linked asset
			asset, exists = h.imagingService.GetAssetByID(*job.AssetID)
		}
	}

	if !exists {
		if job != nil {
			// Job exists but no asset yet (pending or failed)
			utils.SendSuccess(c, "Job status retrieved", AssetStatusResponse{
				AssetID:   job.ID.String(),
				Status:    job.Status,
				CreatedAt: job.CreatedAt.Format(time.RFC3339),
				Error:     job.LastError,
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// 3. We have an asset, build full response with derivatives
	response := AssetStatusResponse{
		AssetID:     asset.ID.String(),
		ContentHash: asset.ContentHash,
		Status:      string(asset.Status),
		CreatedAt:   asset.CreatedAt.Format(time.RFC3339),
		Error:       asset.Error,
	}

	if asset.ProcessedAt != nil {
		response.ProcessedAt = asset.ProcessedAt.Format(time.RFC3339)
	}

	if asset.Status == imaging.StatusReady {
		response.Original = &OriginalInfo{
			Width:     asset.OriginalWidth,
			Height:    asset.OriginalHeight,
			Format:    asset.OriginalFormat,
			SizeBytes: asset.OriginalSize,
		}

		// Group derivatives by rendition name
		derivativeMap := make(map[string]*DerivativeInfo)
		for _, d := range asset.Derivatives {
			if existing, ok := derivativeMap[d.RenditionName]; ok {
				existing.Formats = append(existing.Formats, d.Format)
			} else {
				derivativeMap[d.RenditionName] = &DerivativeInfo{
					Width:      d.Width,
					Height:     d.Height,
					Formats:    []string{d.Format},
					URLPattern: h.imagingService.GetDerivativeURL(asset.ContentHash, d.RenditionName),
				}
			}
		}
		response.Derivatives = derivativeMap
	}

	utils.SendSuccess(c, "Asset status retrieved", response)
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
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	expectedPrefix := fmt.Sprintf("uploads/%s/", userID.String())
	tmpPrefix := fmt.Sprintf("uploads/tmp/%s/", userID.String())
	if !strings.HasPrefix(key, expectedPrefix) && !strings.HasPrefix(key, tmpPrefix) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this file"})
		return
	}

	if err := h.r2.DeleteObject(ctx, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
}

// ServeImage proxies image requests from R2
func (h *UploadHandler) ServeImage(c *gin.Context) {
	hash := c.Param("hash")
	rendition := c.Param("rendition")

	// Check Accept header
	accept := c.GetHeader("Accept")
	preferredFormat := ""
	if strings.Contains(accept, "image/avif") {
		preferredFormat = "avif"
	} else if strings.Contains(accept, "image/webp") {
		preferredFormat = "webp"
	}

	key, _, err := h.imagingService.GetDerivativeKey(hash, rendition, preferredFormat)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	publicURL := h.r2.GetPublicURL(key)

	// Add cache headers
	// Immutable cache for 1 year
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Vary", "Accept")

	// We're redirecting to the actual file
	c.Redirect(http.StatusFound, publicURL)
}
