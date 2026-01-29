package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"maukemana-backend/internal/database"

	"github.com/google/uuid"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// User represents a user from the database
type User struct {
	UserID     uuid.UUID      `db:"user_id" json:"user_id"`
	Email      string         `db:"email" json:"email"`
	Name       sql.NullString `db:"name" json:"name"`
	Role       sql.NullString `db:"role" json:"role"`
	ClerkID    *string        `db:"clerk_id" json:"clerk_id,omitempty"`
	PictureURL *string        `db:"picture_url" json:"picture_url,omitempty"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

// GetByClerkID retrieves a user by Clerk ID
func (r *UserRepository) GetByClerkID(ctx context.Context, clerkID string) (*User, error) {
	var user User
	// Note: Fetching minimal fields as per auth middleware requirements, can expand if needed
	query := "SELECT user_id, email, name, role, clerk_id, picture_url FROM users WHERE clerk_id = $1"
	err := r.db.QueryRowContext(ctx, query, clerkID).Scan(
		&user.UserID, &user.Email, &user.Name, &user.Role, &user.ClerkID, &user.PictureURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("get user by clerk id: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by User Email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := "SELECT user_id, email, name, role, clerk_id, picture_url FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID, &user.Email, &user.Name, &user.Role, &user.ClerkID, &user.PictureURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// UpdateClerkID updates the Clerk ID for an existing user
func (r *UserRepository) UpdateClerkID(ctx context.Context, userID uuid.UUID, clerkID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET clerk_id = $1 WHERE user_id = $2",
		clerkID, userID,
	)
	if err != nil {
		return fmt.Errorf("update user clerk id: %w", err)
	}
	return nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, email, name, picture, clerkID, role string) (*User, error) {
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, name, picture_url, clerk_id, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING user_id`,
		email, name, picture, clerkID, role,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &User{
		UserID:     userID,
		Email:      email,
		Name:       sql.NullString{String: name, Valid: name != ""},
		ClerkID:    &clerkID,
		PictureURL: &picture,
		Role:       sql.NullString{String: role, Valid: role != ""},
	}, nil
}
