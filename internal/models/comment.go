package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	CommentID uuid.UUID  `db:"comment_id" json:"comment_id"`
	PoiID     uuid.UUID  `db:"poi_id" json:"poi_id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Content   string     `db:"content" json:"content"`
	ParentID  *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`

	// Joined fields
	User    *User     `db:"user" json:"user,omitempty"`
	Replies []Comment `db:"replies" json:"replies,omitempty"`
}
