package repositories

import (
	"context"
	"fmt"
	"maukemana-backend/internal/database"
	"time"

	"github.com/google/uuid"
)

// SavedPOIRepository handles saved POI database operations
type SavedPOIRepository struct {
	db *database.DB
}

// NewSavedPOIRepository creates a new SavedPOI repository
func NewSavedPOIRepository(db *database.DB) *SavedPOIRepository {
	return &SavedPOIRepository{db: db}
}

// SavePOI saves a POI for a user
func (r *SavedPOIRepository) SavePOI(ctx context.Context, userID, poiID uuid.UUID) error {
	// Check if already saved to avoid duplicates or errors (idempotent)
	// Alternatively, rely on unique constraint if one exists.
	// Assuming no explicit unique constraint on (user_id, poi_id) in schema shown earlier,
	// but usually there should be.
	// Let's use ON CONFLICT DO NOTHING if the unique index exists.
	// Based on standard schema design, there should be a unique constraint.
	// If not, we should just insert.

	query := `
		INSERT INTO saved_pois (user_id, poi_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, poi_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, userID, poiID, time.Now())
	if err != nil {
		return fmt.Errorf("save poi: %w", err)
	}
	return nil
}

// UnsavePOI removes a saved POI
func (r *SavedPOIRepository) UnsavePOI(ctx context.Context, userID, poiID uuid.UUID) error {
	query := `DELETE FROM saved_pois WHERE user_id = $1 AND poi_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, poiID)
	if err != nil {
		return fmt.Errorf("unsave poi: %w", err)
	}
	return nil
}

// GetSavedPOIs retrieves a user's saved POIs with minimal POI details
// Assuming we want to return the actual POI data structure for the list
func (r *SavedPOIRepository) GetSavedPOIs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]POI, error) {
	var pois []POI
	query := `
		SELECT p.poi_id, p.name, p.category_id, p.description, p.status, p.created_by,
		       p.cover_image_url, p.has_wifi, p.outdoor_seating, p.price_range,
		       COALESCE((SELECT AVG(rating) FROM reviews r WHERE r.poi_id = p.poi_id), 0) as rating_avg,
		       s.created_at as saved_at
		FROM points_of_interest p
		JOIN saved_pois s ON p.poi_id = s.poi_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &pois, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get saved pois: %w", err)
	}
	return pois, nil
}

// IsSaved checks if a POI is saved by the user
func (r *SavedPOIRepository) IsSaved(ctx context.Context, userID, poiID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM saved_pois WHERE user_id = $1 AND poi_id = $2)`
	err := r.db.QueryRowContext(ctx, query, userID, poiID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check is saved: %w", err)
	}
	return exists, nil
}
