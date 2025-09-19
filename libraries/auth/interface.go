package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthInterface defines the contract for authentication implementations
type AuthInterface interface {
	// JWT Operations
	GenerateJWT(sub string) (string, error)
	GenerateJWTWithExpiry(sub string, expiryDuration time.Duration) (string, error)
	Verify(token string) (*jwt.RegisteredClaims, error)

	// Password Operations (with salt support)
	HashPassword(password, salt string) (string, error)
	VerifyPassword(password, hashedPassword, salt string) error
	VerifyPasswordBool(password, hashedPassword, salt string) bool

	// Message Signing (for ECDSA implementations)
	SignMessage(msg []byte) (string, error)
	VerifySignature(msg []byte, signature string) bool

	// Utility functions
	ClientIDFromJWT(token string) (clientID string, err error)
}

// AuthType represents the type of authentication implementation
type AuthType string

const (
	AuthTypeECDSA AuthType = "ecdsa"
	AuthTypeHMAC  AuthType = "hmac"
)
