package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"maukemana-backend/internal/database"
	"maukemana-backend/internal/imaging"

	"github.com/google/uuid"
)

type ImagingRepository struct {
	db *database.DB
}

func NewImagingRepository(db *database.DB) *ImagingRepository {
	return &ImagingRepository{db: db}
}

// CreateAsset inserts a new image asset
func (r *ImagingRepository) CreateAsset(ctx context.Context, asset *imaging.ImageAsset) error {
	query := `
		INSERT INTO image_assets (
			id, content_hash, original_width, original_height, original_format,
			original_size, has_alpha, category, status, version, created_by_user_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		asset.ID, asset.ContentHash, asset.OriginalWidth, asset.OriginalHeight,
		asset.OriginalFormat, asset.OriginalSize, asset.HasAlpha, asset.Category,
		asset.Status, asset.Version, asset.CreatedByUserID, asset.CreatedAt)

	if err != nil {
		return fmt.Errorf("create asset: %w", err)
	}
	return nil
}

// UpdateAssetStatus updates the status of an asset
func (r *ImagingRepository) UpdateAssetStatus(ctx context.Context, id uuid.UUID, status imaging.ProcessingStatus, errorMessage string) error {
	query := `UPDATE image_assets SET status = $1, error_message = $2, processed_at = $3 WHERE id = $4`
	var processedAt *time.Time
	if status == imaging.StatusReady || status == imaging.StatusFailed {
		now := time.Now()
		processedAt = &now
	}

	_, err := r.db.ExecContext(ctx, query, status, errorMessage, processedAt, id)
	if err != nil {
		return fmt.Errorf("update asset status: %w", err)
	}
	return nil
}

// GetAssetByHash retrieves an asset by its content hash
func (r *ImagingRepository) GetAssetByHash(ctx context.Context, hash string) (*imaging.ImageAsset, error) {
	var asset imaging.ImageAsset
	query := `SELECT id, content_hash, original_width, original_height, original_format, original_size, has_alpha, category, status, COALESCE(error_message, '') as error, version, created_by_user_id, created_at, processed_at FROM image_assets WHERE content_hash = $1`

	err := r.db.GetContext(ctx, &asset, query, hash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get asset by hash: %w", err)
	}

	// Load derivatives
	derivatives, err := r.GetDerivatives(ctx, asset.ID)
	if err != nil {
		return nil, fmt.Errorf("get derivatives for asset: %w", err)
	}
	asset.Derivatives = derivatives

	return &asset, nil
}

// GetAssetByID retrieves an asset by its ID
func (r *ImagingRepository) GetAssetByID(ctx context.Context, id uuid.UUID) (*imaging.ImageAsset, error) {
	var asset imaging.ImageAsset
	query := `SELECT id, content_hash, original_width, original_height, original_format, original_size, has_alpha, category, status, COALESCE(error_message, '') as error, version, created_by_user_id, created_at, processed_at FROM image_assets WHERE id = $1`

	err := r.db.GetContext(ctx, &asset, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get asset by id: %w", err)
	}

	// Load derivatives
	derivatives, err := r.GetDerivatives(ctx, asset.ID)
	if err != nil {
		return nil, fmt.Errorf("get derivatives for asset: %w", err)
	}
	asset.Derivatives = derivatives

	return &asset, nil
}

// CreateDerivative inserts a new image derivative
func (r *ImagingRepository) CreateDerivative(ctx context.Context, d imaging.Derivative) error {
	query := `
		INSERT INTO image_derivatives (
			id, asset_id, rendition_name, format, width, height, size_bytes, storage_key
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.AssetID, d.RenditionName, d.Format, d.Width, d.Height, d.SizeBytes, d.StorageKey)

	if err != nil {
		return fmt.Errorf("create derivative: %w", err)
	}
	return nil
}

// GetDerivatives retrieves all derivatives for an asset
func (r *ImagingRepository) GetDerivatives(ctx context.Context, assetID uuid.UUID) ([]imaging.Derivative, error) {
	var derivatives []imaging.Derivative
	query := `SELECT id, asset_id, rendition_name, format, width, height, size_bytes, storage_key FROM image_derivatives WHERE asset_id = $1`

	err := r.db.SelectContext(ctx, &derivatives, query, assetID)
	if err != nil {
		return nil, fmt.Errorf("get derivatives: %w", err)
	}
	return derivatives, nil
}

// CreateJob inserts a new processing job
func (r *ImagingRepository) CreateJob(ctx context.Context, job *imaging.ProcessingJob) error {
	query := `
		INSERT INTO image_processing_jobs (
			id, upload_key, category, user_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		job.ID, job.UploadKey, job.Category, job.UserID, imaging.StatusPending, job.CreatedAt, time.Now())

	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

// UpdateJob updates a job's status and metadata
func (r *ImagingRepository) UpdateJob(ctx context.Context, id uuid.UUID, status imaging.ProcessingStatus, assetID *uuid.UUID, attempts int, lastError string) error {
	query := `UPDATE image_processing_jobs SET status = $1, asset_id = $2, attempts = $3, last_error = $4, updated_at = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, status, assetID, attempts, lastError, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	return nil
}

// GetPendingJobs retrieves all pending jobs
func (r *ImagingRepository) GetPendingJobs(ctx context.Context) ([]imaging.ProcessingJob, error) {
	var jobs []imaging.ProcessingJob
	query := `SELECT id, upload_key, category, user_id, attempts, COALESCE(last_error, '') as last_error, created_at FROM image_processing_jobs WHERE status = 'pending' ORDER BY created_at ASC`

	err := r.db.SelectContext(ctx, &jobs, query)
	if err != nil {
		return nil, fmt.Errorf("get pending jobs: %w", err)
	}
	return jobs, nil
}

// GetJobByID retrieves a specific processing job by its ID
func (r *ImagingRepository) GetJobByID(ctx context.Context, id uuid.UUID) (*imaging.ProcessingJob, error) {
	var job imaging.ProcessingJob
	query := `SELECT id, upload_key, category, user_id, asset_id, status, attempts, COALESCE(last_error, '') as last_error, created_at FROM image_processing_jobs WHERE id = $1`

	err := r.db.GetContext(ctx, &job, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job by id: %w", err)
	}
	return &job, nil
}
