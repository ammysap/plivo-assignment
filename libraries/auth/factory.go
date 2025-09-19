package auth

import (
	"errors"
	"fmt"
)

// AuthFactory creates auth instances based on configuration
type AuthFactory struct{}

// NewAuthFactory creates a new auth factory
func NewAuthFactory() *AuthFactory {
	return &AuthFactory{}
}

// CreateAuth creates an auth instance based on the provided type and config
func (f *AuthFactory) CreateAuth(authType AuthType, config *Config) (AuthInterface, error) {
	switch authType {
	case AuthTypeECDSA:
		return f.createECDSAAuth(config)
	case AuthTypeHMAC:
		return f.createHMACAuth(config)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}
}

// createECDSAAuth creates an ECDSA auth instance
func (f *AuthFactory) createECDSAAuth(config *Config) (AuthInterface, error) {
	// Validate required fields for ECDSA
	if config.PrivateKey == "" {
		return nil, errors.New("PRIVATE_KEY environment variable is required for ECDSA auth")
	}
	if config.PublicKey == "" {
		return nil, errors.New("PUBLIC_KEY environment variable is required for ECDSA auth")
	}

	return NewECDSAAuth(config)
}

// createHMACAuth creates an HMAC auth instance
func (f *AuthFactory) createHMACAuth(config *Config) (AuthInterface, error) {
	// Validate required fields for HMAC
	if config.SecretKey == "" {
		return nil, errors.New("JWT_SECRET_KEY environment variable is required for HMAC auth")
	}

	return NewHMACAuth(config.SecretKey, config.JWTExpirationTime), nil
}

// CreateAuthFromConfig creates an auth instance based on config detection
func (f *AuthFactory) CreateAuthFromConfig(config *Config) (AuthInterface, error) {
	// Auto-detect auth type based on available configuration
	if config.SecretKey != "" {
		// If secret key is provided, use HMAC (simpler)
		return f.createHMACAuth(config)
	} else if config.PrivateKey != "" && config.PublicKey != "" {
		// If both private and public keys are provided, use ECDSA
		return f.createECDSAAuth(config)
	} else {
		return nil, errors.New("no valid auth configuration found. Provide either JWT_SECRET_KEY for HMAC or both PRIVATE_KEY and PUBLIC_KEY for ECDSA")
	}
}
