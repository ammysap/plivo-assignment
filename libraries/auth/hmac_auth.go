package auth

import (
	"errors"
	"time"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HMACAuth implements AuthInterface using HMAC (symmetric key)
type HMACAuth struct {
	secretKey      string
	expirationTime time.Duration
}

// NewHMACAuth creates a new HMAC auth instance
func NewHMACAuth(secretKey string, expirationMinutes int) AuthInterface {
	return &HMACAuth{
		secretKey:      secretKey,
		expirationTime: time.Duration(expirationMinutes) * time.Minute,
	}
}

// GenerateJWT creates a JWT token using HMAC
func (h *HMACAuth) GenerateJWT(sub string) (string, error) {
	return h.GenerateJWTWithExpiry(sub, h.expirationTime)
}

// GenerateJWTWithExpiry creates a JWT token with custom expiry using HMAC
func (h *HMACAuth) GenerateJWTWithExpiry(sub string, expiryDuration time.Duration) (string, error) {
	log := logging.Default()

	claims := &jwt.RegisteredClaims{
		Audience:  jwt.ClaimStrings{"aud"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "shopping-gateway",
		Subject:   sub,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if h.secretKey == "" {
		return "", errors.New("secret key is not configured")
	}

	signedToken, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		log.Errorw("signing token failed", "error", err)
		return "", err
	}

	return signedToken, nil
}

// Verify verifies a JWT token using HMAC
func (h *HMACAuth) Verify(tokenString string) (*jwt.RegisteredClaims, error) {
	log := logging.Default()
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(h.secretKey), nil
	})

	if err != nil {
		log.Errorf("token parsing failed: %s", err)
		return nil, err
	}

	if !token.Valid {
		log.Errorf("token is not valid")
		return nil, errors.New("unauthorized")
	}

	return claims, nil
}

// SignMessage is not supported for HMAC auth (returns error)
func (h *HMACAuth) SignMessage(msg []byte) (string, error) {
	return "", errors.New("message signing not supported for HMAC auth")
}

// VerifySignature is not supported for HMAC auth (returns false)
func (h *HMACAuth) VerifySignature(msg []byte, signature string) bool {
	return false
}

// HashPassword creates a bcrypt hash of the password with salt
func (h *HMACAuth) HashPassword(password, salt string) (string, error) {
	passwordWithSalt := password + salt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash and salt
func (h *HMACAuth) VerifyPassword(password, hashedPassword, salt string) error {
	passwordWithSalt := password + salt
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(passwordWithSalt))
}

// VerifyPasswordBool is a convenience function that returns a boolean instead of an error
func (h *HMACAuth) VerifyPasswordBool(password, hashedPassword, salt string) bool {
	return h.VerifyPassword(password, hashedPassword, salt) == nil
}

// ClientIDFromJWT extracts client ID from JWT token
func (h *HMACAuth) ClientIDFromJWT(token string) (clientID string, err error) {
	jwtToken, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	mapClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("error converting claims")
	}

	sub := mapClaims["sub"]

	subStr, ok := sub.(string)
	if !ok {
		return "", errors.New("error converting sub")
	}

	// For HMAC auth, we assume sub is directly the client ID
	// If you need JSON parsing like ECDSA, uncomment the following:
	/*
		jsonSub := make(map[string]interface{})
		err = json.Unmarshal([]byte(subStr), &jsonSub)
		if err != nil {
			return "", err
		}

		clientID, ok = jsonSub["clientID"].(string)
		if !ok {
			return "", errors.New("error converting clientId")
		}
	*/

	// For simplicity, return sub as clientID
	return subStr, nil
}
