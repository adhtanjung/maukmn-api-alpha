package models

import (
	"time"

	"github.com/google/uuid"
)

// Review represents a user review
type Review struct {
	ReviewID  uuid.UUID `db:"review_id" json:"review_id"`
	PoiID     uuid.UUID `db:"poi_id" json:"poi_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Rating    *int      `db:"rating" json:"rating,omitempty"`
	Content   *string   `db:"content" json:"content,omitempty"`
	Upvotes   int       `db:"upvotes" json:"upvotes"`
	Downvotes int       `db:"downvotes" json:"downvotes"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
