package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// Processor handles image processing operations
// In production, this would use libvips (govips) for better performance
// This implementation uses pure Go for initial development
type Processor struct {
	// Configuration
	maxConcurrency int
}

// NewProcessor creates a new image processor
func NewProcessor() *Processor {
	return &Processor{
		maxConcurrency: 4,
	}
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
	// Decode the source image
	reader := bytes.NewReader(data)
	srcImg, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	renditions := GetRenditionsForCategory(category)
	var results []ProcessedImage

	for _, rendition := range renditions {
		// Skip if source is smaller than target
		srcBounds := srcImg.Bounds()
		if rendition.Width > srcBounds.Dx() && rendition.Height > srcBounds.Dy() {
			continue
		}

		// Process the rendition
		processed, err := p.processRendition(srcImg, rendition, hasAlpha)
		if err != nil {
			// Log error but continue with other renditions
			fmt.Printf("Warning: failed to process rendition %s: %v\n", rendition.Name, err)
			continue
		}

		results = append(results, processed...)
	}

	return results, nil
}

// processRendition generates a single rendition in all required formats
func (p *Processor) processRendition(src image.Image, config RenditionConfig, hasAlpha bool) ([]ProcessedImage, error) {
	// Apply cropping and resizing
	resized := p.resizeAndCrop(src, config)

	// Get quality settings
	quality := config.Quality.GetSettings()

	// Get formats to generate
	formats := GetFormatsForRendition(hasAlpha, config.SkipAVIF)

	var results []ProcessedImage
	bounds := resized.Bounds()

	for _, format := range formats {
		var buf bytes.Buffer
		var err error

		switch format {
		case "jpg", "jpeg":
			err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality.JPEG})
		case "png":
			encoder := png.Encoder{CompressionLevel: png.BestCompression}
			err = encoder.Encode(&buf, resized)
		case "webp":
			// Pure Go WebP encoding is limited
			// In production, use libvips/govips for WebP encoding
			// For now, fall back to JPEG
			err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality.JPEG})
			format = "jpg" // Update format since we fell back
		case "avif":
			// AVIF encoding requires libvips/govips
			// For now, skip AVIF in pure Go implementation
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("failed to encode %s: %w", format, err)
		}

		results = append(results, ProcessedImage{
			Name:      config.Name,
			Format:    format,
			Width:     bounds.Dx(),
			Height:    bounds.Dy(),
			Data:      buf.Bytes(),
			SizeBytes: buf.Len(),
		})
	}

	return results, nil
}

// resizeAndCrop applies the specified crop mode and resizing
func (p *Processor) resizeAndCrop(src image.Image, config RenditionConfig) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	switch config.CropMode {
	case CropCenterSquare:
		// Center crop to square, then resize
		size := srcW
		if srcH < size {
			size = srcH
		}
		cropped := imaging.CropCenter(src, size, size)
		return imaging.Resize(cropped, config.Width, config.Height, imaging.Lanczos)

	case CropCenter16x9:
		// Center crop to 16:9, then resize
		targetRatio := 16.0 / 9.0
		currentRatio := float64(srcW) / float64(srcH)

		var cropW, cropH int
		if currentRatio > targetRatio {
			// Image is wider than 16:9, crop width
			cropH = srcH
			cropW = int(float64(srcH) * targetRatio)
		} else {
			// Image is taller than 16:9, crop height
			cropW = srcW
			cropH = int(float64(srcW) / targetRatio)
		}
		cropped := imaging.CropCenter(src, cropW, cropH)
		return imaging.Resize(cropped, config.Width, config.Height, imaging.Lanczos)

	case CropFitWidth:
		// Scale to width, maintain aspect ratio
		return imaging.Resize(src, config.Width, 0, imaging.Lanczos)

	default: // CropNone
		// Fit within dimensions maintaining aspect ratio
		return imaging.Fit(src, config.Width, config.Height, imaging.Lanczos)
	}
}

// StripEXIF removes EXIF metadata from image data
// In production, use libvips for this
func (p *Processor) StripEXIF(data []byte) ([]byte, error) {
	// Decode and re-encode to strip metadata
	reader := bytes.NewReader(data)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(&buf, img)
	default:
		// For other formats, return original
		return data, nil
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
