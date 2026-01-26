package imaging

import (
	"fmt"
	"log"
	"runtime"

	"github.com/davidbyttow/govips/v2/vips"
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

// ProcessImage generates all renditions for an image
func (p *Processor) ProcessImage(data []byte, category string, hasAlpha bool) ([]ProcessedImage, error) {
	// Import image from buffer
	// govips uses streaming processing where possible
	srcParams := vips.NewImportParams()
	srcParams.FailOnError.Set(true)

	srcImage, err := vips.LoadImageFromBuffer(data, srcParams)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	// Important: we can't reuse the same *vips.ImageRef for concurrent operations
	// or sequential ops that modify it. We must clone or reload.
	// However, LoadImageFromBuffer creates a new ref.
	// For multiple renditions, it's efficient to keep one "source" ref open
	// and copy/clone it for each operation.
	// Defer closing the source image.
	defer srcImage.Close()

	srcW := srcImage.Width()
	srcH := srcImage.Height()

	renditions := GetRenditionsForCategory(category)
	var results []ProcessedImage

	for _, rendition := range renditions {
		// Skip if source is smaller than target (avoid upscaling)
		if rendition.Width > srcW && (rendition.Height == 0 || rendition.Height > srcH) {
			continue
		}

		// Process the rendition
		processed, err := p.processRendition(data, rendition, hasAlpha)
		if err != nil {
			log.Printf("Warning: failed to process rendition %s: %v", rendition.Name, err)
			continue
		}

		results = append(results, processed...)
	}

	return results, nil
}

// processRendition generates a single rendition in all required formats
func (p *Processor) processRendition(srcData []byte, config RenditionConfig, hasAlpha bool) ([]ProcessedImage, error) {
	// Get formats to generate
	formats := GetFormatsForRendition(hasAlpha, config.SkipAVIF)
	var results []ProcessedImage

	for _, format := range formats {
		// We must create a fresh pipeline from the source buffer for each format
		// or copy the vips image ref safely.
		// For safety and simplicity in this implementation, we reload from buffer.
		// libvips is very fast at this.
		img, err := vips.LoadImageFromBuffer(srcData, vips.NewImportParams())
		if err != nil {
			return nil, fmt.Errorf("failed to load source for %s: %w", format, err)
		}

		// Apply resize/crop
		if err := p.resizeAndCrop(img, config); err != nil {
			img.Close()
			return nil, fmt.Errorf("failed to resize: %w", err)
		}

		var bytes []byte
		var exportErr error

		// Export based on format
		// Using standard quality settings
		q := config.Quality.GetSettings()

		switch format {
		case "jpg", "jpeg":
			ep := vips.NewJpegExportParams()
			ep.Quality = q.JPEG
			ep.StripMetadata = true
			bytes, _, exportErr = img.ExportJpeg(ep)
		case "png":
			ep := vips.NewPngExportParams()
			ep.Compression = 6 // Default best
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
			ep.Speed = 5 // Balanced speed/size
			bytes, _, exportErr = img.ExportAvif(ep)
		}

		// Clean up the image ref immediately
		img.Close()

		if exportErr != nil {
			log.Printf("Warning: failed to export %s: %v", format, exportErr)
			continue
		}

		results = append(results, ProcessedImage{
			Name:      config.Name,
			Format:    format,
			Width:     config.Width,
			Height:    config.Height, // Note: actual height might differ if auto-height
			Data:      bytes,
			SizeBytes: len(bytes),
		})
	}

	return results, nil
}

// resizeAndCrop applies the specified crop mode and resizing
func (p *Processor) resizeAndCrop(img *vips.ImageRef, config RenditionConfig) error {
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
