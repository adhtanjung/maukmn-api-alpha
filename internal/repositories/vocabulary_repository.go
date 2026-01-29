package repositories

import (
	"context"
	"fmt"

	"maukemana-backend/internal/database"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// VocabularyRepository handles vocabulary database operations
type VocabularyRepository struct {
	db *database.DB
}

// NewVocabularyRepository creates a new vocabulary repository
func NewVocabularyRepository(db *database.DB) *VocabularyRepository {
	return &VocabularyRepository{db: db}
}

// Vocabulary represents a vocabulary item from the database
type Vocabulary struct {
	VocabID   uuid.UUID      `db:"vocab_id" json:"vocab_id"`
	VocabType string         `db:"vocab_type" json:"vocab_type"`
	Key       string         `db:"key" json:"key"`
	Aliases   pq.StringArray `db:"aliases" json:"aliases"`
	Icon      *string        `db:"icon" json:"icon,omitempty"`
	IsActive  bool           `db:"is_active" json:"is_active"`
}

// GetActive retrieves active vocabularies, optionally filtered by type
func (r *VocabularyRepository) GetActive(ctx context.Context, vocabType string) ([]Vocabulary, error) {
	query := `
		SELECT vocab_id, vocab_type, key, aliases, icon, is_active
		FROM vocabularies
		WHERE is_active = true
	`
	args := []interface{}{}

	if vocabType != "" {
		query += " AND vocab_type = $1"
		args = append(args, vocabType)
	}
	query += " ORDER BY vocab_type, key"

	var vocabularies []Vocabulary
	err := r.db.SelectContext(ctx, &vocabularies, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get active vocabularies: %w", err)
	}
	return vocabularies, nil
}
