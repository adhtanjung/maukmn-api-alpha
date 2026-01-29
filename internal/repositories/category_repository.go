package repositories

import (
	"context"
	"fmt"
	"time"

	"maukemana-backend/internal/database"

	"github.com/google/uuid"
)

// CategoryRepository handles category database operations
type CategoryRepository struct {
	db *database.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *database.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Category represents a category from the database
type Category struct {
	CategoryID       uuid.UUID  `db:"category_id" json:"category_id"`
	ParentCategoryID *uuid.UUID `db:"parent_category_id" json:"parent_category_id,omitempty"`
	NameKey          string     `db:"name_key" json:"name_key"`
	Icon             *string    `db:"icon" json:"icon,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}

// GetAll retrieves all categories
func (r *CategoryRepository) GetAll(ctx context.Context) ([]Category, error) {
	query := `
		SELECT category_id, parent_category_id, name_key, icon, created_at
		FROM categories
		ORDER BY name_key
	`

	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("get all categories: %w", err)
	}
	return categories, nil
}
