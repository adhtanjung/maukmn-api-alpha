package imaging

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	_ "golang.org/x/image/webp"
)

// ValidationResult contains the results of image validation
type ValidationResult struct {
	Valid        bool
	Width        int
	Height       int
	Format       string // Detected format from magic bytes
	HasAlpha     bool
	OriginalSize int64
	ContentHash  string // SHA-256 hash
	Error        string
}

// CategoryLimits defines upload limits per category
type CategoryLimits struct {
	MaxBytes     int64
	MaxDimension int // Max width OR height
}

// GetCategoryLimits returns limits for a given category
func GetCategoryLimits(category string) CategoryLimits {
	// All categories now have 15MB limit as per user request to allow high-res uploads
	return CategoryLimits{
		MaxBytes:     15 * 1024 * 1024, // 15MB
		MaxDimension: 6000,             // 6000px
	}
}

// AllowedFormats defines which image formats are accepted
var AllowedFormats = map[string]bool{
	"jpeg": true,
	"png":  true,
	"webp": true,
	"gif":  true, // Static only
	"heic": true, // Will be transcoded
	"avif": true,
}

// Magic bytes for format detection
var magicBytes = map[string][]byte{
	"jpeg": {0xFF, 0xD8, 0xFF},
	"png":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	"gif":  {0x47, 0x49, 0x46, 0x38},
	"webp": nil, // WEBP has RIFF header, checked separately
	"avif": nil, // AVIF has ftyp header, checked separately
	"heic": nil, // HEIC has ftyp header, checked separately
}

// DetectFormat detects image format from magic bytes
func DetectFormat(data []byte) string {
	if len(data) < 12 {
		return ""
	}

	// JPEG
	if bytes.HasPrefix(data, magicBytes["jpeg"]) {
		return "jpeg"
	}

	// PNG
	if bytes.HasPrefix(data, magicBytes["png"]) {
		return "png"
	}

	// GIF
	if bytes.HasPrefix(data, magicBytes["gif"]) {
		return "gif"
	}

	// WebP (RIFF....WEBP)
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "webp"
	}

	// HEIC/AVIF (ftyp box detection)
	if len(data) >= 12 {
		// Check for ftyp box
		if bytes.Equal(data[4:8], []byte("ftyp")) {
			brand := string(data[8:12])
			switch brand {
			case "heic", "heix", "hevc", "hevx", "mif1":
				return "heic"
			case "avif", "avis":
				return "avif"
			}
		}
	}

	return ""
}

// ValidateImage performs comprehensive image validation
func ValidateImage(data []byte, category string) (*ValidationResult, error) {
	limits := GetCategoryLimits(category)
	result := &ValidationResult{
		OriginalSize: int64(len(data)),
	}

	// 1. Check byte size
	if int64(len(data)) > limits.MaxBytes {
		result.Error = fmt.Sprintf("file size %d exceeds maximum %d bytes", len(data), limits.MaxBytes)
		return result, errors.New(result.Error)
	}

	// 2. Detect format from magic bytes (NOT Content-Type header)
	format := DetectFormat(data)
	if format == "" {
		result.Error = "unable to detect image format"
		return result, errors.New(result.Error)
	}

	if !AllowedFormats[format] {
		result.Error = fmt.Sprintf("format %s is not allowed", format)
		return result, errors.New(result.Error)
	}

	result.Format = format

	// 3. Decode image to get dimensions (use Go's image package for basic validation)
	// For production, we'll use libvips for full decoding
	reader := bytes.NewReader(data)
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		// Try to continue - some formats may not be supported by Go's image package
		// In production, libvips will handle all formats
		if format != "heic" && format != "avif" {
			result.Error = fmt.Sprintf("failed to decode image: %v", err)
			return result, errors.New(result.Error)
		}
		// For HEIC/AVIF, we'll validate dimensions during processing
		config.Width = 0
		config.Height = 0
	}

	result.Width = config.Width
	result.Height = config.Height

	// 4. Check dimensions (decompression bomb protection)
	if config.Width > limits.MaxDimension || config.Height > limits.MaxDimension {
		result.Error = fmt.Sprintf("image dimensions %dx%d exceed maximum %d", config.Width, config.Height, limits.MaxDimension)
		return result, errors.New(result.Error)
	}

	// Check for decompression bomb (too many pixels)
	maxPixels := int64(64 * 1024 * 1024) // 64 megapixels
	if int64(config.Width)*int64(config.Height) > maxPixels {
		result.Error = "image too large (potential decompression bomb)"
		return result, errors.New(result.Error)
	}

	// 5. Compute content hash for deduplication
	hash := sha256.Sum256(data)
	result.ContentHash = hex.EncodeToString(hash[:])

	// 6. Detect if image has alpha channel
	reader.Seek(0, io.SeekStart)
	img, _, err := image.Decode(reader)
	if err == nil {
		result.HasAlpha = hasAlphaChannel(img)
	}

	result.Valid = true
	return result, nil
}

// hasAlphaChannel checks if an image has an alpha channel
func hasAlphaChannel(img image.Image) bool {
	switch img.(type) {
	case *image.RGBA, *image.NRGBA, *image.RGBA64, *image.NRGBA64:
		return true
	default:
		return false
	}
}

// ComputeContentHash computes SHA-256 hash of data
func ComputeContentHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
