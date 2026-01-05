package models

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a POI category
type Category struct {
	CategoryID       uuid.UUID  `db:"category_id" json:"category_id"`
	ParentCategoryID *uuid.UUID `db:"parent_category_id" json:"parent_category_id,omitempty"`
	NameKey          string     `db:"name_key" json:"name_key"`
	Icon             *string    `db:"icon" json:"icon,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}
