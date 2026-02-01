package imaging

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/davidbyttow/govips/v2/vips"
	"golang.org/x/sync/errgroup"
)

// Processor handles image processing operations
type Processor struct {
	maxConcurrency int
}

// NewProcessor creates a new image processor
func NewProcessor() *Processor {
	// Initialize libvips
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
		CacheTrace:       false,
		CollectStats:     true,
	})

	return &Processor{
		maxConcurrency: runtime.NumCPU(),
	}
}

// Shutdown cleans up libvips resources
func (p *Processor) Shutdown() {
	vips.Shutdown()
}

// ProcessedImage represents a processed image rendition
type ProcessedImage struct {
	Name      string
	Format    string
	Width     int
	Height    int
	Data      []byte
	SizeBytes int
}

// ProcessImage generates all renditions for an image in parallel
func (p *Processor) ProcessImage(ctx context.Context, data []byte, category string, hasAlpha bool, cropConfig *CropConfig) ([]ProcessedImage, error) {
	// Initialize source to check dimensions
	srcParams := vips.NewImportParams()
	srcParams.FailOnError.Set(true)

	tmpImage, err := vips.LoadImageFromBuffer(data, srcParams)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	srcW := tmpImage.Width()
	srcH := tmpImage.Height()
	tmpImage.Close()

	renditions := GetRenditionsForCategory(category)

	// Use errgroup for parallel processing across available CPU cores
	g, ctx := errgroup.WithContext(ctx)
	// Limit concurrency if needed, but errgroup usually handles it via goroutines
	// libvips has its own internal pools too.

	resultsChan := make(chan []ProcessedImage, len(renditions))

	for _, rendition := range renditions {
		r := rendition // capture for goroutine

		// Skip if source is smaller than target (avoid upscaling)
		if r.Width > srcW && (r.Height == 0 || r.Height > srcH) {
			continue
		}

		g.Go(func() error {
			processed, err := p.processRendition(ctx, data, r, hasAlpha, cropConfig)
			if err != nil {
				slog.Error("failed to process rendition", "rendition", r.Name, "error", err)
				return nil // Don't fail the whole job if one rendition fails
			}
			resultsChan <- processed
			return nil
		})
	}

	// Wait for all renditions to finish
	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(resultsChan)

	var allResults []ProcessedImage
	for res := range resultsChan {
		allResults = append(allResults, res...)
	}

	return allResults, nil
}

// processRendition generates a single rendition in all required formats
func (p *Processor) processRendition(ctx context.Context, srcData []byte, config RenditionConfig, hasAlpha bool, cropConfig *CropConfig) ([]ProcessedImage, error) {
	formats := GetFormatsForRendition(hasAlpha, config.SkipAVIF)

	// Load the source image ONCE for this rendition
	baseImg, err := vips.LoadImageFromBuffer(srcData, vips.NewImportParams())
	if err != nil {
		return nil, fmt.Errorf("failed to load source: %w", err)
	}
	defer baseImg.Close()

	// 1. Apply shared transformations (Resize/Crop) once
	if err := p.resizeAndCrop(baseImg, config, cropConfig); err != nil {
		return nil, fmt.Errorf("failed to resize/crop: %w", err)
	}

	var results []ProcessedImage

	// Sequential processing of formats to avoid exploding concurrency
	for _, format := range formats {
		// Check context before each heavy operation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Clone for this format
		img, err := baseImg.Copy()
		if err != nil {
			return nil, fmt.Errorf("failed to copy image for %s: %w", format, err)
		}

		var bytes []byte
		var exportErr error
		q := config.Quality.GetSettings()

		switch format {
		case "jpg", "jpeg":
			ep := vips.NewJpegExportParams()
			ep.Quality = q.JPEG
			ep.StripMetadata = true
			bytes, _, exportErr = img.ExportJpeg(ep)
		case "png":
			ep := vips.NewPngExportParams()
			ep.Compression = 6
			ep.StripMetadata = true
			bytes, _, exportErr = img.ExportPng(ep)
		case "webp":
			ep := vips.NewWebpExportParams()
			ep.Quality = q.WebP
			ep.StripMetadata = true
			bytes, _, exportErr = img.ExportWebp(ep)
		case "avif":
			ep := vips.NewAvifExportParams()
			ep.Quality = q.AVIF
			ep.StripMetadata = true
			ep.Speed = 8 // Increased speed for faster encoding (range 0-9)
			bytes, _, exportErr = img.ExportAvif(ep)
		}

		img.Close() // Explicitly close the copy immediately

		if exportErr != nil {
			return nil, exportErr
		}

		results = append(results, ProcessedImage{
			Name:      config.Name,
			Format:    format,
			Width:     config.Width,
			Height:    baseImg.Height(), // Use baseImg height as it's already resized
			Data:      bytes,
			SizeBytes: len(bytes),
		})
	}

	return results, nil
}

// resizeAndCrop applies the specified crop mode and resizing
func (p *Processor) resizeAndCrop(img *vips.ImageRef, config RenditionConfig, cropConfig *CropConfig) error {
	// Apply custom crop first if enabled and available
	if config.UseCustomCrop && cropConfig != nil {
		width := img.Width()
		height := img.Height()

		// Calculate absolute coordinates
		left := int(float64(width) * cropConfig.X)
		top := int(float64(height) * cropConfig.Y)
		cropWidth := int(float64(width) * cropConfig.Width)
		cropHeight := int(float64(height) * cropConfig.Height)

		// Validate bounds
		if left < 0 {
			left = 0
		}
		if top < 0 {
			top = 0
		}
		if left+cropWidth > width {
			cropWidth = width - left
		}
		if top+cropHeight > height {
			cropHeight = height - top
		}

		// Perform extraction if dimensions are valid
		if cropWidth > 0 && cropHeight > 0 {
			if err := img.ExtractArea(left, top, cropWidth, cropHeight); err != nil {
				return fmt.Errorf("extract area: %w", err)
			}
		}
	}

	switch config.CropMode {
	case CropCenterSquare:
		// Smart thumbnail crop to square
		return img.Thumbnail(config.Width, config.Width, vips.InterestingCentre)

	case CropCenter16x9:
		// Smart thumbnail crop to 16:9
		return img.Thumbnail(config.Width, config.Height, vips.InterestingCentre)

	case CropFitWidth:
		// Resize to width, maintain aspect, no upscaling
		return img.ThumbnailWithSize(config.Width, 20000, vips.InterestingNone, vips.SizeDown)

	default: // CropNone
		// Resize to fit within box
		return img.Thumbnail(config.Width, config.Height, vips.InterestingNone)
	}
}

// StripEXIF removes EXIF metadata from image data
func (p *Processor) StripEXIF(data []byte) ([]byte, error) {
	img, err := vips.LoadImageFromBuffer(data, vips.NewImportParams())
	if err != nil {
		return nil, err
	}
	defer img.Close()

	// Export as JPEG (most common source) to strip metadata
	// If source was PNG/WebP, we might want to respect that, but
	// for now this is mostly used for raw uploads which are mostly JPEGs.
	// Actually, just re-exporting in same format is better.

	// Determine format from magic bytes or vips loader
	loader := vips.DetermineImageType(data)

	switch loader {
	case vips.ImageTypePNG:
		ep := vips.NewPngExportParams()
		ep.StripMetadata = true
		b, _, err := img.ExportPng(ep)
		return b, err
	case vips.ImageTypeWEBP:
		ep := vips.NewWebpExportParams()
		ep.StripMetadata = true
		b, _, err := img.ExportWebp(ep)
		return b, err
	default:
		// Default to JPEG
		ep := vips.NewJpegExportParams()
		ep.StripMetadata = true
		ep.Quality = 95
		b, _, err := img.ExportJpeg(ep)
		return b, err
	}
}
