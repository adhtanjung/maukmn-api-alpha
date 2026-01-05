package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Email     string    `db:"email" json:"email"`
	Name      *string   `db:"name" json:"name,omitempty"`
	Picture   *string   `db:"picture_url" json:"picture,omitempty"`
	GoogleID  *string   `db:"google_id" json:"google_id,omitempty"`
	ClerkID   *string   `db:"clerk_id" json:"clerk_id,omitempty"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
