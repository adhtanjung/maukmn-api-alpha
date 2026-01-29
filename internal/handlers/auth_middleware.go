package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"maukemana-backend/internal/auth"
	"maukemana-backend/internal/repositories"
	"maukemana-backend/internal/utils"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByClerkID(ctx context.Context, clerkID string) (*repositories.User, error)
	GetByEmail(ctx context.Context, email string) (*repositories.User, error)
	UpdateClerkID(ctx context.Context, userID uuid.UUID, clerkID string) error
	Create(ctx context.Context, email, name, picture, clerkID, role string) (*repositories.User, error)
}

// AuthHandler handles authentication routes (Clerk integration mostly happens in middleware)
type AuthHandler struct {
	repo UserRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(repo UserRepository) *AuthHandler {
	return &AuthHandler{
		repo: repo,
	}
}

// AuthMiddleware validates Clerk token and syncs user to DB
func AuthMiddleware(repo UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized: missing token", nil)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid header format"})
			return
		}

		tokenStr := parts[1]
		claims, err := auth.VerifyToken(tokenStr)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized: invalid token", err)
			return
		}

		// Lazy Sync
		clerkID := claims.Subject
		var userEmail string
		var userID uuid.UUID
		var displayName sql.NullString
		var dbRole sql.NullString

		// 1. Check if user exists by Clerk ID -- AND fetch role
		user, err := repo.GetByClerkID(c.Request.Context(), clerkID)

		if err == nil {
			// Found user in DB
			userID = user.UserID
			userEmail = user.Email
			displayName = user.Name
			dbRole = user.Role
		} else if err == sql.ErrNoRows {
			// 2. User NOT found by Clerk ID. We need to sync.
			clerkUser, err := auth.GetUser(clerkID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to fetch user info from Clerk"})
				return
			}

			if len(clerkUser.EmailAddresses) == 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User has no email address"})
				return
			}
			primaryEmail := clerkUser.EmailAddresses[0].EmailAddress
			userEmail = primaryEmail

			var name string
			if clerkUser.FirstName != nil {
				name = *clerkUser.FirstName
				if clerkUser.LastName != nil {
					name += " " + *clerkUser.LastName
				}
			}
			displayName = sql.NullString{String: name, Valid: name != ""}

			// 3. Check if user exists by Email (Migrate legacy user)
			legacyUser, err := repo.GetByEmail(c.Request.Context(), primaryEmail)

			if err == nil {
				// Legacy user found, update with clerk_id
				userID = legacyUser.UserID
				err = repo.UpdateClerkID(c.Request.Context(), userID, clerkID)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update legacy user"})
					return
				}
				// Also fetch role/name if needed, but assuming legacy user has them.
				// For now taking from legacyUser if we want, or just proceed.
				// The original code didn't re-fetch legacy user details other than ID.
				// But we need 'dbRole' for context.
				dbRole = legacyUser.Role
				if legacyUser.Name.Valid {
					displayName = legacyUser.Name
				}
			} else if err == sql.ErrNoRows {
				// 4. Create new user
				var picture string
				if clerkUser.ImageURL != nil {
					picture = *clerkUser.ImageURL
				}

				newUser, err := repo.Create(c.Request.Context(), primaryEmail, name, picture, clerkID, "user")
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
					return
				}
				userID = newUser.UserID
				dbRole = newUser.Role
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Set context
		// Extract values from NullString with defaults
		finalDisplayName := ""
		if displayName.Valid {
			finalDisplayName = displayName.String
		}
		finalRole := "user"
		if dbRole.Valid && dbRole.String != "" {
			finalRole = dbRole.String
		}

		c.Set("user_id", userID)
		c.Set("email", userEmail)
		c.Set("display_name", finalDisplayName)
		c.Set("user_role", finalRole)

		c.Next()
	}
}

// GetMe returns the current user's info
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	displayName, _ := c.Get("display_name")
	email, _ := c.Get("email")
	role, _ := c.Get("user_role")
	// userID is uuid.UUID

	utils.SendSuccess(c, "User profile retrieved", gin.H{
		"user_id":      userID,
		"email":        email,
		"display_name": displayName,
		"role":         role,
	})
}
