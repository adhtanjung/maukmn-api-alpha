package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"maukemana-backend/internal/database"

	"github.com/google/uuid"
)

type PhotoRepository struct {
	db *database.DB
}

func NewPhotoRepository(db *database.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

// VoteWithToggle handles upvote/downvote with Reddit-style toggle logic.
// Returns: newScore, userVote (1, -1, or 0), error
//
// Logic:
// - If user hasn't voted: insert vote
// - If user voted the same way: remove vote (toggle off)
// - If user voted opposite: switch vote
func (r *PhotoRepository) VoteWithToggle(ctx context.Context, photoID, userID uuid.UUID, voteType int) (int, int, error) {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if user already voted
	var existingVote sql.NullInt64
	err = tx.QueryRowContext(ctx, `
		SELECT vote_type FROM photo_votes
		WHERE photo_id = $1 AND user_id = $2
	`, photoID, userID).Scan(&existingVote)

	if err != nil && err != sql.ErrNoRows {
		return 0, 0, fmt.Errorf("check existing vote: %w", err)
	}

	var userVote int

	if !existingVote.Valid {
		// No existing vote - INSERT new vote
		_, err = tx.ExecContext(ctx, `
			INSERT INTO photo_votes (photo_id, user_id, vote_type)
			VALUES ($1, $2, $3)
		`, photoID, userID, voteType)
		if err != nil {
			return 0, 0, fmt.Errorf("insert vote: %w", err)
		}
		userVote = voteType
	} else if int(existingVote.Int64) == voteType {
		// Same vote - DELETE (toggle off)
		_, err = tx.ExecContext(ctx, `
			DELETE FROM photo_votes
			WHERE photo_id = $1 AND user_id = $2
		`, photoID, userID)
		if err != nil {
			return 0, 0, fmt.Errorf("delete vote: %w", err)
		}
		userVote = 0
	} else {
		// Opposite vote - UPDATE
		_, err = tx.ExecContext(ctx, `
			UPDATE photo_votes
			SET vote_type = $3, created_at = NOW()
			WHERE photo_id = $1 AND user_id = $2
		`, photoID, userID, voteType)
		if err != nil {
			return 0, 0, fmt.Errorf("update vote: %w", err)
		}
		userVote = voteType
	}

	// Recalculate score from votes table
	var newScore int
	err = tx.QueryRowContext(ctx, `
		UPDATE photos
		SET upvotes = COALESCE((
				SELECT COUNT(*) FROM photo_votes
				WHERE photo_id = $1 AND vote_type = 1
			), 0),
			downvotes = COALESCE((
				SELECT COUNT(*) FROM photo_votes
				WHERE photo_id = $1 AND vote_type = -1
			), 0),
			score = COALESCE((
				SELECT SUM(vote_type) FROM photo_votes
				WHERE photo_id = $1
			), 0)
		WHERE photo_id = $1
		RETURNING score
	`, photoID).Scan(&newScore)
	if err != nil {
		return 0, 0, fmt.Errorf("update photo score: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit transaction: %w", err)
	}

	return newScore, userVote, nil
}

// GetUserVote returns the current vote status for a user on a photo
// Returns: voteType (1, -1, or 0 if no vote)
func (r *PhotoRepository) GetUserVote(ctx context.Context, photoID, userID uuid.UUID) (int, error) {
	var voteType sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT vote_type FROM photo_votes
		WHERE photo_id = $1 AND user_id = $2
	`, photoID, userID).Scan(&voteType)

	if err == sql.ErrNoRows || !voteType.Valid {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get user vote: %w", err)
	}

	return int(voteType.Int64), nil
}
