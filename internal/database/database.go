package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// DB represents the PostgreSQL database connection
type DB struct {
	*sqlx.DB
}

// New creates a new PostgreSQL database connection
func New(databaseURL string) (*DB, error) {
	db, err := otelsqlx.Connect("postgres", databaseURL,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Health checks the database connection health
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return db.BeginTxx(ctx, nil)
}

// RefreshMaterializedView refreshes the POI materialized view
func (db *DB) RefreshMaterializedView(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_pois_with_hero")
	return err
}
