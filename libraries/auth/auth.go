// This package deals with jwts
package auth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/golang-jwt/jwt/v5"
)

var (
	instance AuthInterface
	once     sync.Once
	mu       sync.RWMutex
)

// InitAuth initializes the auth module with configuration using singleton pattern
func InitAuth(authType AuthType) {
	log := logging.Default()

	once.Do(func() {
		var err error
		factory := NewAuthFactory()

		switch authType {
		case AuthTypeECDSA:
			config := LoadECDSAConfig()
			instance, err = factory.CreateAuth(AuthTypeECDSA, &Config{
				PrivateKey:        config.PrivateKey,
				PublicKey:         config.PublicKey,
				JWTExpirationTime: 1440, // Default 24 hours
			})
			if err != nil {
				log.Errorw("failed to create ECDSA auth instance", "error", err)
				panic(fmt.Sprintf("failed to initialize ECDSA auth: %v", err))
			}
			log.Infow("ECDSA auth initialized successfully")

		case AuthTypeHMAC:
			config := LoadHMACConfig()
			instance, err = factory.CreateAuth(AuthTypeHMAC, &Config{
				SecretKey:         config.SecretKey,
				JWTExpirationTime: 1440, // Default 24 hours
			})
			if err != nil {
				log.Errorw("failed to create HMAC auth instance", "error", err)
				panic(fmt.Sprintf("failed to initialize HMAC auth: %v", err))
			}
			log.Infow("HMAC auth initialized successfully")

		default:
			panic(fmt.Sprintf("unsupported auth type: %s", authType))
		}
	})
}

// getAuthType returns the type of auth instance for logging
func getAuthType(auth AuthInterface) string {
	switch auth.(type) {
	case *ECDSAAuth:
		return "ECDSA"
	case *HMACAuth:
		return "HMAC"
	default:
		return "Unknown"
	}
}

func GenerateJWT(sub string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return "", errors.New("auth not initialized")
	}
	return instance.GenerateJWT(sub)
}

func GenerateJWTWithExpiry(
	sub string,
	expiryDuration time.Duration,
) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return "", errors.New("auth not initialized")
	}
	return instance.GenerateJWTWithExpiry(sub, expiryDuration)
}

func Verify(token string) (*jwt.RegisteredClaims, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return nil, errors.New("auth not initialized")
	}
	return instance.Verify(token)
}

func VerifyWithPublicKey(
	token string, publicKey *ecdsa.PublicKey,
) (*jwt.RegisteredClaims, error) {
	log := logging.Default()
	claims := &jwt.RegisteredClaims{}

	tkn, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		},
	)
	if err != nil {
		log.Errorf("token: %s Parsing failed with %s\n", token, err)
		return nil, err
	}

	if !tkn.Valid {
		log.Errorf("token: %s not valid\n", token)
		return nil, errors.New("unauthorized")
	}

	return claims, nil
}

// HashPassword creates a bcrypt hash of the password with salt
func HashPassword(password, salt string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return "", errors.New("auth not initialized")
	}
	return instance.HashPassword(password, salt)
}

// VerifyPassword verifies a password against its hash and salt
func VerifyPassword(password, hashedPassword, salt string) error {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return errors.New("auth not initialized")
	}
	return instance.VerifyPassword(password, hashedPassword, salt)
}

// VerifyPasswordBool is a convenience function that returns a boolean instead of an error
func VerifyPasswordBool(password, hashedPassword, salt string) bool {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return false
	}
	return instance.VerifyPasswordBool(password, hashedPassword, salt)
}

// SignMessage signs a message (only supported by ECDSA auth)
func SignMessage(msg []byte) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return "", errors.New("auth not initialized")
	}
	return instance.SignMessage(msg)
}

// VerifySignature verifies a message signature (only supported by ECDSA auth)
func VerifySignature(msg []byte, signature string) bool {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return false
	}
	return instance.VerifySignature(msg, signature)
}

// ClientIDFromJWT extracts client ID from JWT token
func ClientIDFromJWT(token string) (clientID string, err error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return "", errors.New("auth not initialized")
	}
	return instance.ClientIDFromJWT(token)
}

// GetInstance returns the current auth instance (for testing or advanced usage)
func GetInstance() AuthInterface {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// IsInitialized checks if the auth instance is initialized
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return instance != nil
}

// GetAuthType returns the type of the current auth instance
func GetAuthType() string {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return "Not Initialized"
	}
	return getAuthType(instance)
}
