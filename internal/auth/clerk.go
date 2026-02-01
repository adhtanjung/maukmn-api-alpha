package auth

import (
	"context"
	"os"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

// InitClerk initializes the Clerk SDK
func InitClerk() {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		// Log warning or fatal depending on preference, for now just ensure it's set in env
		panic("CLERK_SECRET_KEY not set")
	}
	clerk.SetKey(secretKey)
}

// VerifyToken verifies the session token and returns the claims
func VerifyToken(token string) (*clerk.SessionClaims, error) {
	claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
		Token:  token,
		Leeway: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// GetUser retrieves a user from Clerk by ID
func GetUser(userID string) (*clerk.User, error) {
	return user.Get(context.Background(), userID)
}
