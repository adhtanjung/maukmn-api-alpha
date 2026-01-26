-- +goose Up
-- +goose StatementBegin

CREATE TYPE processing_status AS ENUM ('pending', 'processing', 'ready', 'failed');

CREATE TABLE image_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash VARCHAR(64) UNIQUE NOT NULL,
    original_width INTEGER NOT NULL,
    original_height INTEGER NOT NULL,
    original_format VARCHAR(10) NOT NULL,
    original_size BIGINT NOT NULL,
    has_alpha BOOLEAN NOT NULL DEFAULT FALSE,
    category VARCHAR(50) NOT NULL,
    status processing_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    created_by_user_id UUID NOT NULL REFERENCES users(user_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE image_derivatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES image_assets(id) ON DELETE CASCADE,
    rendition_name VARCHAR(50) NOT NULL,
    format VARCHAR(10) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    size_bytes INTEGER NOT NULL,
    storage_key TEXT NOT NULL,
    UNIQUE(asset_id, rendition_name, format)
);

-- Simple job queue table for persistent processing
CREATE TABLE image_processing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    upload_key TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(user_id),
    status processing_status NOT NULL DEFAULT 'pending',
    asset_id UUID REFERENCES image_assets(id),
    attempts INTEGER DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_image_assets_hash ON image_assets(content_hash);
CREATE INDEX idx_image_derivatives_asset_id ON image_derivatives(asset_id);
CREATE INDEX idx_image_processing_jobs_status ON image_processing_jobs(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS image_processing_jobs;
DROP TABLE IF EXISTS image_derivatives;
DROP TABLE IF EXISTS image_assets;
DROP TYPE IF EXISTS processing_status;
-- +goose StatementEnd
