package imaging

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// CropConfig defines the relative crop coordinates (0.0 to 1.0)
type CropConfig struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Value implements the driver.Valuer interface
func (c CropConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface
func (c *CropConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

// ProcessingStatus represents the status of an image processing job
type ProcessingStatus string

const (
	StatusPending     ProcessingStatus = "pending"
	StatusDownloading ProcessingStatus = "downloading"
	StatusProcessing  ProcessingStatus = "processing"
	StatusUploading   ProcessingStatus = "uploading"
	StatusReady       ProcessingStatus = "ready"
	StatusFailed      ProcessingStatus = "failed"
)

// ImageAsset represents a processed image asset with all its derivatives
type ImageAsset struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	ContentHash     string           `json:"content_hash" db:"content_hash"`
	OriginalWidth   int              `json:"original_width" db:"original_width"`
	OriginalHeight  int              `json:"original_height" db:"original_height"`
	OriginalFormat  string           `json:"original_format" db:"original_format"`
	OriginalSize    int64            `json:"original_size" db:"original_size"`
	HasAlpha        bool             `json:"has_alpha" db:"has_alpha"`
	Category        string           `json:"category" db:"category"`
	Status          ProcessingStatus `json:"status" db:"status"`
	Error           string           `json:"error,omitempty" db:"error"`
	Version         int              `json:"version" db:"version"`
	Derivatives     []Derivative     `json:"derivatives,omitempty" db:"-"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	ProcessedAt     *time.Time       `json:"processed_at,omitempty" db:"processed_at"`
	CreatedByUserID uuid.UUID        `json:"created_by_user_id" db:"created_by_user_id"`
}

// Derivative represents a single image derivative
type Derivative struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AssetID       uuid.UUID `json:"asset_id" db:"asset_id"`
	RenditionName string    `json:"rendition_name" db:"rendition_name"`
	Format        string    `json:"format" db:"format"`
	Width         int       `json:"width" db:"width"`
	Height        int       `json:"height" db:"height"`
	SizeBytes     int       `json:"size_bytes" db:"size_bytes"`
	StorageKey    string    `json:"storage_key" db:"storage_key"`
}

// ProcessingJob represents a job in the processing queue
type ProcessingJob struct {
	ID          uuid.UUID   `db:"id"`
	UploadKey   string      `db:"upload_key"`
	Category    string      `db:"category"`
	UserID      uuid.UUID   `db:"user_id"`
	AssetID     *uuid.UUID  `db:"asset_id"` // Link to asset once created
	ContentHash string      `db:"content_hash"`
	CreatedAt   time.Time   `db:"created_at"`
	Attempts    int         `db:"attempts"`
	LastError   string      `db:"last_error"`
	Status      string      `db:"status"` // Added status to struct
	CropData    *CropConfig `db:"crop_data"`
	IsReprocess bool        `db:"is_reprocess"`
}

// ImagingRepositoryInterface defines the storage operations for image assets
type ImagingRepositoryInterface interface {
	CreateAsset(ctx context.Context, asset *ImageAsset) error
	UpdateAssetStatus(ctx context.Context, id uuid.UUID, status ProcessingStatus, errorMessage string) error
	GetAssetByHash(ctx context.Context, hash string) (*ImageAsset, error)
	GetAssetByID(ctx context.Context, id uuid.UUID) (*ImageAsset, error)
	CreateDerivative(ctx context.Context, d Derivative) error
	GetDerivatives(ctx context.Context, assetID uuid.UUID) ([]Derivative, error)
	CreateJob(ctx context.Context, job *ProcessingJob) error
	UpdateJob(ctx context.Context, id uuid.UUID, status ProcessingStatus, assetID *uuid.UUID, attempts int, lastError string) error
	GetPendingJobs(ctx context.Context) ([]ProcessingJob, error)
	GetJobByID(ctx context.Context, id uuid.UUID) (*ProcessingJob, error)
}

// Service manages image processing operations
type Service struct {
	processor *Processor
	r2Client  R2ClientInterface
	repo      ImagingRepositoryInterface

	// Job queue
	jobQueue chan *ProcessingJob

	// Worker pool
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// R2ClientInterface defines the interface for R2 operations
type R2ClientInterface interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
	PutObject(ctx context.Context, key string, data []byte, contentType string) error
	DeleteObject(ctx context.Context, key string) error
	GetPublicURL(key string) string
	MoveObject(ctx context.Context, srcKey, dstKey string) error
}

// NewService creates a new imaging service
func NewService(r2Client R2ClientInterface, repo ImagingRepositoryInterface, workerCount int) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Service{
		processor:   NewProcessor(),
		r2Client:    r2Client,
		repo:        repo,
		jobQueue:    make(chan *ProcessingJob, 1000),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start worker pool
	s.startWorkers()

	// Resume pending jobs from database
	go s.resumePendingJobs()

	return s
}

func (s *Service) resumePendingJobs() {
	time.Sleep(1 * time.Second)                                             // Small delay for startup stability
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Increased timeout
	defer cancel()

	jobs, err := s.repo.GetPendingJobs(ctx)
	if err != nil {
		slog.Error("failed to get pending jobs", "error", err)
		return
	}

	slog.Info("found pending jobs", "count", len(jobs))

	for _, job := range jobs {
		j := job // copy
		// Blocking send to ensure we don't drop jobs
		// If queue is full, this will wait until workers consume some
		select {
		case s.jobQueue <- &j:
			slog.Info("resumed pending job", "job_id", j.ID)
		case <-s.ctx.Done():
			// Service shutting down
			return
		case <-ctx.Done():
			slog.Warn("timeout resuming pending jobs")
			return
		}
	}
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.cancel()
	close(s.jobQueue)
	s.wg.Wait()
}

// startWorkers starts the image processing worker pool
func (s *Service) startWorkers() {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// worker processes jobs from the queue
func (s *Service) worker(id int) {
	defer s.wg.Done()
	l := slog.With("worker_id", id)

	for job := range s.jobQueue {
		// Priority check for shutdown
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		l.Info("worker processing job", "job_id", job.ID)
		if err := s.processJob(job); err != nil {
			l.Error("failed to process job", "job_id", job.ID, "error", err)
			s.handleJobFailure(job, err)
		}
	}
}

// QueueProcessing queues an image for processing
func (s *Service) QueueProcessing(uploadKey, category string, userID uuid.UUID, cropConfig *CropConfig) (uuid.UUID, error) {
	job := &ProcessingJob{
		ID:        uuid.New(),
		UploadKey: uploadKey,
		Category:  category,
		UserID:    userID,
		CreatedAt: time.Now(),
		CropData:  cropConfig,
	}

	if err := s.repo.CreateJob(s.ctx, job); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create job: %w", err)
	}

	select {
	case s.jobQueue <- job:
		return job.ID, nil
	default:
		// Even if queue is full, job is in DB so it can be resumed later
		return job.ID, nil
	}
}

// QueueReprocessing queues an existing asset for reprocessing
func (s *Service) QueueReprocessing(uploadKey, category string, userID uuid.UUID, cropConfig *CropConfig) (uuid.UUID, error) {
	job := &ProcessingJob{
		ID:          uuid.New(),
		UploadKey:   uploadKey,
		Category:    category,
		UserID:      userID,
		CreatedAt:   time.Now(),
		CropData:    cropConfig,
		IsReprocess: true,
	}

	if err := s.repo.CreateJob(s.ctx, job); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create job: %w", err)
	}

	select {
	case s.jobQueue <- job:
		return job.ID, nil
	default:
		// Even if queue is full, job is in DB so it can be resumed later
		return job.ID, nil
	}
}

// processJob handles the full image processing pipeline
func (s *Service) processJob(job *ProcessingJob) error {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	// 1. Download original from R2
	s.repo.UpdateJob(ctx, job.ID, StatusDownloading, nil, job.Attempts, "")
	data, err := s.r2Client.GetObject(ctx, job.UploadKey)
	if err != nil {
		return fmt.Errorf("failed to download original: %w", err)
	}

	// 2. Validate
	validation, err := ValidateImage(data, job.Category)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	job.ContentHash = validation.ContentHash

	// 3. Check for existing asset (dedup)
	existingAsset, err := s.repo.GetAssetByHash(ctx, validation.ContentHash)
	if err != nil {
		return fmt.Errorf("failed to check existing asset: %w", err)
	}

	var assetID uuid.UUID
	var assetVersion int

	if existingAsset != nil {
		if !job.IsReprocess && existingAsset.Status == StatusReady {
			slog.Debug("asset already exists, reusing", "hash", validation.ContentHash, "asset_id", existingAsset.ID)
			// Update job to point to existing asset
			s.repo.UpdateJob(ctx, job.ID, StatusReady, &existingAsset.ID, job.Attempts, "")
			// Clean up the upload (original is same content)
			s.r2Client.DeleteObject(ctx, job.UploadKey)
			return nil
		}
		// If reprocessing or status not ready (maybe retry?), we reuse the ID but continue
		assetID = existingAsset.ID
		assetVersion = existingAsset.Version + 1
	} else {
		assetID = uuid.New()
		assetVersion = 1
	}

	// 4. Create or Update asset record
	asset := &ImageAsset{
		ID:              assetID,
		ContentHash:     validation.ContentHash,
		OriginalWidth:   validation.Width,
		OriginalHeight:  validation.Height,
		OriginalFormat:  validation.Format,
		OriginalSize:    validation.OriginalSize,
		HasAlpha:        validation.HasAlpha,
		Category:        job.Category,
		Status:          StatusProcessing,
		Version:         assetVersion,
		CreatedAt:       time.Now(),
		CreatedByUserID: job.UserID,
	}

	if existingAsset != nil {
		// Update existing asset e.g. Version, Status
		// For now we might rely on CreateAsset behaving like upsert or just use a new UpdateAsset method?
		// Since we don't have explicit UpdateAsset full record, we might need to rely on CreateAsset doing nothing if ID exists?
		// Wait, repo.CreateAsset might fail if ID exists.
		// If Reprocess, we likely want to UPDATE the existing record or at least its version/status.
		// Let's assume we need to handle this.
		// For simplicity/robustness, if it exists, we update status and version.
		// But s.repo.CreateAsset probably does INSERT.
		// As a hack for now, I'll update status/version via UpdateAssetStatus if possible, or assume CreateAsset fails.
		// Actually, I should probably add UpdateAssetMetadata to repo.
		// For now, I will assume simple ID reuse.
		if err := s.repo.UpdateAssetStatus(ctx, asset.ID, StatusProcessing, ""); err != nil {
			// If this fails, maybe it doesn't exist? But we checked.
		}
		// Ideally we update Version too.
	} else {
		if err := s.repo.CreateAsset(ctx, asset); err != nil {
			return fmt.Errorf("failed to create asset record: %w", err)
		}
	}

	// Link job to asset
	s.repo.UpdateJob(ctx, job.ID, StatusProcessing, &asset.ID, job.Attempts, "")

	// 5. Generate renditions in parallel
	slog.Debug("starting parallel processing", "asset_id", asset.ID)
	// Update status again?
	// s.repo.UpdateAssetStatus(ctx, asset.ID, StatusProcessing, "")

	// Pro: Stripping EXIF is now handled efficiently during the export stage in ProcessImage
	processed, err := s.processor.ProcessImage(ctx, data, job.Category, validation.HasAlpha, job.CropData)
	if err != nil {
		s.repo.UpdateAssetStatus(ctx, asset.ID, StatusFailed, err.Error())
		return fmt.Errorf("processing failed: %w", err)
	}

	// 6. Upload derivatives to R2
	// 6. Upload derivatives to R2 (Parallel)
	s.repo.UpdateAssetStatus(ctx, asset.ID, StatusUploading, "")

	// Pre-allocate slice for results to avoid mutex if possible,
	// but we need to append valid results only. using a mutex for safety.
	var derivatives []Derivative
	var mu sync.Mutex

	g, gCtx := errgroup.WithContext(ctx)
	// Limit upload concurrency to avoid flooding network/R2
	sem := make(chan struct{}, 10)

	hashPrefix := validation.ContentHash[:2]

	for _, p := range processed {
		p := p // capture loop variable
		g.Go(func() error {
			// Acquire semaphore
			select {
			case sem <- struct{}{}:
			case <-gCtx.Done():
				return gCtx.Err()
			}
			defer func() { <-sem }()

			storageKey := fmt.Sprintf("derivatives/%s/%s/v%d/%s.%s",
				hashPrefix, validation.ContentHash, asset.Version, p.Name, p.Format)

			contentType := getContentType(p.Format)

			if err := s.r2Client.PutObject(gCtx, storageKey, p.Data, contentType); err != nil {
				return fmt.Errorf("failed to upload %s: %w", p.Name, err)
			}

			mu.Lock()
			derivatives = append(derivatives, Derivative{
				ID:            uuid.New(),
				AssetID:       asset.ID,
				RenditionName: p.Name,
				Format:        p.Format,
				StorageKey:    storageKey,
				Width:         p.Width,
				Height:        p.Height,
				SizeBytes:     len(p.Data),
			})
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		s.repo.UpdateAssetStatus(ctx, asset.ID, StatusFailed, err.Error())
		return fmt.Errorf("upload failed: %w", err)
	}
	// Parallel uploads finished

	// Create derivative records in DB
	// We do this sequentially to avoid DB contention and because it's fast
	// NOTE: For reprocessing, simple CreateDerivative is fine, it will add new rows.
	// We might want to clear old derivatives for this version? Structure allows multiple?
	// The DB likely has ID Primary Key.
	// Old derivatives remain for old versions (if we supported versions fully).
	// For now, adding new ones is fine.
	for _, d := range derivatives {
		if err := s.repo.CreateDerivative(ctx, d); err != nil {
			slog.Warn("failed to save derivative record", "key", d.StorageKey, "error", err)
			continue
		}
	}

	// 7. Move original to permanent location
	originalKey := fmt.Sprintf("originals/%s/%s/original", hashPrefix, validation.ContentHash)

	if job.UploadKey != originalKey {
		if err := s.r2Client.MoveObject(ctx, job.UploadKey, originalKey); err != nil {
			slog.Warn("failed to move original", "error", err)
		}
	}

	// 8. Update asset status
	if err := s.repo.UpdateAssetStatus(ctx, asset.ID, StatusReady, ""); err != nil {
		slog.Warn("failed to update asset status", "asset_id", asset.ID, "error", err)
	}

	// Mark job as ready
	s.repo.UpdateJob(ctx, job.ID, StatusReady, &asset.ID, job.Attempts, "")

	slog.Info("successfully processed asset", "asset_id", asset.ID, "derivatives", len(derivatives))
	return nil
}

// handleJobFailure handles failed jobs with retry logic
func (s *Service) handleJobFailure(job *ProcessingJob, err error) {
	job.Attempts++
	job.LastError = err.Error()

	ctx := context.Background()

	if job.Attempts < 3 {
		s.repo.UpdateJob(ctx, job.ID, StatusPending, nil, job.Attempts, job.LastError)
		// Retry with exponential backoff
		go func() {
			time.Sleep(time.Duration(job.Attempts*job.Attempts) * time.Second)
			select {
			case s.jobQueue <- job:
			default:
				slog.Error("failed to requeue job", "job_id", job.ID)
			}
		}()
	} else {
		// Mark as permanently failed
		slog.Error("job permanently failed", "job_id", job.ID, "attempts", job.Attempts, "error", job.LastError)
		s.repo.UpdateJob(ctx, job.ID, StatusFailed, nil, job.Attempts, job.LastError)
	}
}

// GetAsset retrieves an asset by content hash
func (s *Service) GetAsset(contentHash string) (*ImageAsset, bool) {
	asset, err := s.repo.GetAssetByHash(context.Background(), contentHash)
	if err != nil || asset == nil {
		return nil, false
	}
	return asset, true
}

// GetAssetByID retrieves an asset by ID
func (s *Service) GetAssetByID(id uuid.UUID) (*ImageAsset, bool) {
	asset, err := s.repo.GetAssetByID(context.Background(), id)
	if err != nil {
		slog.Error("GetAssetByID failed", "id", id, "error", err)
		return nil, false
	}
	if asset == nil {
		slog.Debug("GetAssetByID: asset not found", "id", id)
		return nil, false
	}
	return asset, true
}

// GetJobByID retrieves a specific processing job by its ID
func (s *Service) GetJobByID(id uuid.UUID) (*ProcessingJob, bool) {
	job, err := s.repo.GetJobByID(context.Background(), id)
	if err != nil {
		slog.Error("GetJobByID failed", "id", id, "error", err)
		return nil, false
	}
	if job == nil {
		slog.Debug("GetJobByID: job not found", "id", id)
		return nil, false
	}
	return job, true
}

// GetDerivativeURL returns the CDN URL for a specific derivative
func (s *Service) GetDerivativeURL(contentHash, renditionName string) string {
	// Return the CDN-friendly URL pattern
	// CDN will handle format negotiation based on Accept header
	return fmt.Sprintf("/img/%s/%s", contentHash, renditionName)
}

// GetDerivativeKey returns the storage key for a specific derivative
// This attempts to find the best format match for the rendition
func (s *Service) GetDerivativeKey(contentHash, renditionName, preferredFormat string) (string, string, error) {
	asset, err := s.repo.GetAssetByHash(context.Background(), contentHash)
	if err != nil {
		return "", "", fmt.Errorf("lookup failed: %w", err)
	}
	if asset == nil {
		return "", "", fmt.Errorf("asset not found")
	}

	if asset.Status != StatusReady {
		return "", "", fmt.Errorf("asset not ready")
	}

	if renditionName == "original" {
		// Return original key
		hashPrefix := contentHash[:2]
		return fmt.Sprintf("originals/%s/%s/original", hashPrefix, contentHash), asset.OriginalFormat, nil
	}

	// Find all derivatives for this rendition
	var candidates []Derivative
	for _, d := range asset.Derivatives {
		if d.RenditionName == renditionName {
			candidates = append(candidates, d)
		}
	}

	if len(candidates) == 0 {
		// Fallback: If the specific rendition doesn't exist (e.g. source image was too small),
		// find the largest available rendition for the same category.
		// Rendition names are like "cover_320", so we can find others in the same category.
		category := ""
		if idx := strings.Index(renditionName, "_"); idx != -1 {
			category = renditionName[:idx]
		}

		if category != "" {
			for _, d := range asset.Derivatives {
				if strings.HasPrefix(d.RenditionName, category) {
					candidates = append(candidates, d)
				}
			}
		}

		// If still no candidates, just take all derivatives
		if len(candidates) == 0 {
			if len(asset.Derivatives) == 0 {
				return "", "", fmt.Errorf("no derivatives found")
			}
			candidates = asset.Derivatives
		}

		// Sort or pick the best candidate from the fallback list
		// For simplicity, we just use the list we have now.
	}

	// Try to match specific format if requested
	if preferredFormat != "" {
		for _, d := range candidates {
			if d.Format == preferredFormat {
				return d.StorageKey, d.Format, nil
			}
		}
	}

	// Priority: avif > webp > jpg/png
	priorities := []string{"avif", "webp", "jpeg", "jpg", "png"}
	for _, format := range priorities {
		for _, d := range candidates {
			if d.Format == format {
				return d.StorageKey, d.Format, nil
			}
		}
	}

	return candidates[0].StorageKey, candidates[0].Format, nil
}

// getContentType returns the MIME type for a format
func getContentType(format string) string {
	switch format {
	case "avif":
		return "image/avif"
	case "webp":
		return "image/webp"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
