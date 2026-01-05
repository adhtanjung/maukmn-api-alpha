package models

import (
	"time"

	"github.com/google/uuid"
)

// Itinerary represents a user's itinerary
type Itinerary struct {
	ItineraryID uuid.UUID  `db:"itinerary_id" json:"itinerary_id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	Title       string     `db:"title" json:"title"`
	Description *string    `db:"description" json:"description,omitempty"`
	StartDate   *time.Time `db:"start_date" json:"start_date,omitempty"`
	EndDate     *time.Time `db:"end_date" json:"end_date,omitempty"`
	IsPublic    bool       `db:"is_public" json:"is_public"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// ItineraryItem represents an item in an itinerary
type ItineraryItem struct {
	ItemID      uuid.UUID  `db:"item_id" json:"item_id"`
	ItineraryID uuid.UUID  `db:"itinerary_id" json:"itinerary_id"`
	PoiID       uuid.UUID  `db:"poi_id" json:"poi_id"`
	Day         int        `db:"day" json:"day"`
	OrderIndex  int        `db:"order_index" json:"order_index"`
	PlannedTime *time.Time `db:"planned_time" json:"planned_time,omitempty"`
	Duration    *int       `db:"duration" json:"duration,omitempty"`
	Notes       *string    `db:"notes" json:"notes,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
