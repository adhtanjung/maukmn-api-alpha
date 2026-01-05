package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"maukemana-backend/internal/auth"
	"maukemana-backend/internal/database"
)

// AuthHandler handles authentication routes (Clerk integration mostly happens in middleware)
type AuthHandler struct {
	db *database.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

// AuthMiddleware validates Clerk token and syncs user to DB
func AuthMiddleware(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing token"})
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid token - " + err.Error()})
			return
		}

		// Lazy Sync
		clerkID := claims.Subject
		var userEmail string
		// Basic way to get email from claims if Clerk provides it in session claims,
		// otherwise might need to fetch user.
		// For now, assuming email is not always in initial claims, we might need to fetch user from Clerk if creating.
		// NOTE: Session claims usually don't have email unless customized.
		// We'll rely on checking DB by ClerkID first.

		var userID uuid.UUID
		var displayName string

		// 1. Check if user exists by Clerk ID
		err = db.QueryRowContext(c.Request.Context(),
			"SELECT user_id, email, name FROM users WHERE clerk_id = $1",
			clerkID,
		).Scan(&userID, &userEmail, &displayName)

		if err == nil {
			// User found
			c.Set("user_id", userID)
			c.Set("email", userEmail)
			c.Set("display_name", displayName)
			c.Next()
			return
		} else if err != sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// 2. User NOT found by Clerk ID. We need to sync.
		// Fetch full user details from Clerk to get email
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

		var name string
		if clerkUser.FirstName != nil {
			name = *clerkUser.FirstName
			if clerkUser.LastName != nil {
				name += " " + *clerkUser.LastName
			}
		}

		// 3. Check if user exists by Email (Migrate legacy user)
		err = db.QueryRowContext(c.Request.Context(),
			"SELECT user_id FROM users WHERE email = $1",
			primaryEmail,
		).Scan(&userID)

		if err == nil {
			// Legacy user found, update with clerk_id
			_, err = db.ExecContext(c.Request.Context(),
				"UPDATE users SET clerk_id = $1 WHERE user_id = $2",
				clerkID, userID,
			)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update legacy user"})
				return
			}
		} else if err == sql.ErrNoRows {
			// 4. Create new user
			var picture string
			if clerkUser.ImageURL != nil {
				picture = *clerkUser.ImageURL
			}

			err = db.QueryRowContext(c.Request.Context(),
				`INSERT INTO users (email, name, picture_url, clerk_id, role)
				 VALUES ($1, $2, $3, $4, 'user')
				 RETURNING user_id`,
				primaryEmail, name, picture, clerkID,
			).Scan(&userID)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Set context
		c.Set("user_id", userID)
		c.Set("email", primaryEmail)
		c.Set("display_name", name)

		c.Next()
	}
}

// GetMe returns the current user's info
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	dialplayName, _ := c.Get("display_name")
	email, _ := c.Get("email")
	// userID is uuid.UUID

	c.JSON(http.StatusOK, gin.H{
		"user_id":      userID,
		"email":        email,
		"display_name": dialplayName,
	})
}
