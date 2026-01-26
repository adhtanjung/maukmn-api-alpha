package imaging

// RenditionConfig defines how to generate a specific image rendition
type RenditionConfig struct {
	Name     string
	Width    int
	Height   int // 0 means maintain aspect ratio
	CropMode CropMode
	Quality  QualityLevel
	SkipAVIF bool // Skip AVIF for very small images
}

// CropMode defines how images should be cropped
type CropMode string

const (
	CropNone         CropMode = "none"          // No cropping, fit within dimensions
	CropCenterSquare CropMode = "center-square" // Center crop to square
	CropCenter16x9   CropMode = "center-16x9"   // Center crop to 16:9
	CropFitWidth     CropMode = "fit-width"     // Scale to width, maintain aspect
)

// QualityLevel defines compression quality presets
type QualityLevel string

const (
	QualityHigh   QualityLevel = "high"
	QualityMedium QualityLevel = "medium"
	QualityLow    QualityLevel = "low"
)

// QualitySettings returns encoder quality values for a given level
func (q QualityLevel) GetSettings() QualitySettings {
	switch q {
	case QualityHigh:
		return QualitySettings{AVIF: 24, WebP: 85, JPEG: 88}
	case QualityMedium:
		return QualitySettings{AVIF: 30, WebP: 78, JPEG: 82}
	case QualityLow:
		return QualitySettings{AVIF: 36, WebP: 70, JPEG: 75}
	default:
		return QualitySettings{AVIF: 30, WebP: 78, JPEG: 82}
	}
}

// QualitySettings holds quality values for each format
type QualitySettings struct {
	AVIF int // 0-63, lower = better quality
	WebP int // 0-100, higher = better quality
	JPEG int // 0-100, higher = better quality
}

// GetRenditionsForCategory returns the image ladder for a category
func GetRenditionsForCategory(category string) []RenditionConfig {
	switch category {
	case "profile":
		return []RenditionConfig{
			{Name: "profile_48", Width: 48, Height: 48, CropMode: CropCenterSquare, Quality: QualityHigh, SkipAVIF: true},
			{Name: "profile_96", Width: 96, Height: 96, CropMode: CropCenterSquare, Quality: QualityHigh, SkipAVIF: true},
			{Name: "profile_200", Width: 200, Height: 200, CropMode: CropCenterSquare, Quality: QualityHigh},
			{Name: "profile_400", Width: 400, Height: 400, CropMode: CropCenterSquare, Quality: QualityMedium},
		}
	case "cover":
		return []RenditionConfig{
			{Name: "cover_320", Width: 320, Height: 180, CropMode: CropCenter16x9, Quality: QualityMedium},
			{Name: "cover_640", Width: 640, Height: 360, CropMode: CropCenter16x9, Quality: QualityMedium},
			{Name: "cover_960", Width: 960, Height: 540, CropMode: CropCenter16x9, Quality: QualityMedium},
			{Name: "cover_1200", Width: 1200, Height: 675, CropMode: CropCenter16x9, Quality: QualityMedium},
			{Name: "cover_1920", Width: 1920, Height: 1080, CropMode: CropCenter16x9, Quality: QualityMedium},
		}
	case "gallery":
		return []RenditionConfig{
			{Name: "gallery_thumb", Width: 150, Height: 150, CropMode: CropCenterSquare, Quality: QualityMedium},
			{Name: "gallery_320", Width: 320, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "gallery_640", Width: 640, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "gallery_960", Width: 960, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "gallery_1200", Width: 1200, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "gallery_1920", Width: 1920, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
		}
	default: // general
		return []RenditionConfig{
			{Name: "general_320", Width: 320, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "general_640", Width: 640, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "general_960", Width: 960, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
			{Name: "general_1200", Width: 1200, Height: 0, CropMode: CropFitWidth, Quality: QualityMedium},
		}
	}
}

// GetFormatsForRendition returns the output formats to generate
// based on whether the image has alpha channel
func GetFormatsForRendition(hasAlpha bool, skipAVIF bool) []string {
	if hasAlpha {
		if skipAVIF {
			return []string{"webp", "png"}
		}
		return []string{"avif", "webp", "png"}
	}
	if skipAVIF {
		return []string{"webp", "jpg"}
	}
	return []string{"avif", "webp", "jpg"}
}
