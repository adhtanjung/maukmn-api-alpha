package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Vocabulary represents an admin-controlled vocabulary
type Vocabulary struct {
	VocabID   uuid.UUID      `db:"vocab_id" json:"vocab_id"`
	VocabType string         `db:"vocab_type" json:"vocab_type"`
	Key       string         `db:"key" json:"key"`
	Aliases   pq.StringArray `db:"aliases" json:"aliases"`
	Icon      *string        `db:"icon" json:"icon,omitempty"`
	IsActive  bool           `db:"is_active" json:"is_active"`
}
