package models

import (
	"time"

	"github.com/google/uuid"
)

// SavedPOI represents a user's saved POI
type SavedPOI struct {
	SavedPoiID uuid.UUID `db:"saved_poi_id" json:"saved_poi_id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	PoiID      uuid.UUID `db:"poi_id" json:"poi_id"`
	Notes      *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
