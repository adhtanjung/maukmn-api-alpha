package models

import (
	"time"

	"github.com/google/uuid"
)

// Photo represents a POI photo
type Photo struct {
	PhotoID         uuid.UUID  `db:"photo_id" json:"photo_id"`
	PoiID           uuid.UUID  `db:"poi_id" json:"poi_id"`
	UserID          *uuid.UUID `db:"user_id" json:"user_id,omitempty"`
	URL             string     `db:"url" json:"url"`
	OriginalURL     *string    `db:"original_url" json:"original_url,omitempty"`
	IsAdminOfficial bool       `db:"is_admin_official" json:"is_admin_official"`
	IsPinned        bool       `db:"is_pinned" json:"is_pinned"`
	Upvotes         int        `db:"upvotes" json:"upvotes"`
	Downvotes       int        `db:"downvotes" json:"downvotes"`
	VibeCategory    *string    `db:"vibe_category" json:"vibe_category,omitempty"`
	Score           int        `db:"score" json:"score"`
	IsHero          bool       `db:"is_hero" json:"is_hero"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}
