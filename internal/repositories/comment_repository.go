package repositories

import (
	"context"
	"errors"
	"fmt"
	"maukemana-backend/internal/database"
	"maukemana-backend/internal/models"

	"github.com/google/uuid"
)

type CommentRepository struct {
	db *database.DB
}

func NewCommentRepository(db *database.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (poi_id, user_id, content, parent_id)
		VALUES (:poi_id, :user_id, :content, :parent_id)
		RETURNING comment_id, created_at, updated_at
	`
	rows, err := r.db.NamedQueryContext(ctx, query, comment)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&comment.CommentID, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return fmt.Errorf("scan comment: %w", err)
		}
		return nil
	}
	return nil
}

func (r *CommentRepository) GetByPOI(ctx context.Context, poiID uuid.UUID, limit, offset int) ([]models.Comment, error) {
	query := `
		SELECT
			c.*,
			u.user_id "user.user_id",
			u.name "user.name",
			u.picture_url "user.picture_url"
		FROM comments c
		JOIN users u ON c.user_id = u.user_id
		WHERE c.poi_id = $1 AND c.parent_id IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`
	var comments []models.Comment
	err := r.db.SelectContext(ctx, &comments, query, poiID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get comments by poi: %w", err)
	}
	return comments, nil
}

func (r *CommentRepository) GetReplies(ctx context.Context, parentID uuid.UUID) ([]models.Comment, error) {
	query := `
		SELECT
			c.*,
			u.user_id "user.user_id",
			u.name "user.name",
			u.picture_url "user.picture_url"
		FROM comments c
		JOIN users u ON c.user_id = u.user_id
		WHERE c.parent_id = $1
		ORDER BY c.created_at ASC
	`
	var comments []models.Comment
	err := r.db.SelectContext(ctx, &comments, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("get replies: %w", err)
	}
	return comments, nil
}

func (r *CommentRepository) Delete(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM comments WHERE comment_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, commentID, userID)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete comment rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("not found")
	}
	return nil
}
