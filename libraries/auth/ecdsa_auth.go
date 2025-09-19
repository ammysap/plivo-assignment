package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ECDSAJWTConfig holds ECDSA-specific configuration
type ECDSAJWTConfig struct {
	PrivateKey     *ecdsa.PrivateKey
	PublicKey      *ecdsa.PublicKey
	ExpirationTime time.Duration
}

// ECDSAAuth implements AuthInterface using ECDSA keys
type ECDSAAuth struct {
	config     *ECDSAJWTConfig
	authConfig *Config
}

// NewECDSAAuth creates a new ECDSA auth instance
func NewECDSAAuth(authConfig *Config) (*ECDSAAuth, error) {
	ecdsaAuth := &ECDSAAuth{
		authConfig: authConfig,
	}

	// Import private key
	privateKey, err := ecdsaAuth.importECDSAPrivateKey()
	if err != nil {
		return nil, err
	}

	// Import public key
	publicKey, err := ecdsaAuth.importECDSAPublicKey()
	if err != nil {
		return nil, err
	}

	ecdsaAuth.config = &ECDSAJWTConfig{
		PrivateKey:     privateKey,
		PublicKey:      publicKey,
		ExpirationTime: time.Duration(authConfig.JWTExpirationTime) * time.Minute,
	}

	return ecdsaAuth, nil
}

func (e *ECDSAAuth) importECDSAPublicKey() (*ecdsa.PublicKey, error) {
	if e.authConfig == nil {
		return nil, errors.New("auth config not initialized")
	}

	publicKeyBytes := e.authConfig.PublicKey

	decodedPublicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBytes)
	if err != nil {
		// Try parsing directly if decoding fails
		ecdsaPublicKey, parseErr := jwt.ParseECPublicKeyFromPEM([]byte(publicKeyBytes))
		if parseErr != nil {
			return nil, errors.New("failed to parse public key: " + parseErr.Error())
		}

		return ecdsaPublicKey, nil
	}

	ecdsaPublicKey, err := jwt.ParseECPublicKeyFromPEM(decodedPublicKeyBytes)
	if err != nil {
		return nil, errors.New("provided key is not a valid ecdsa public key: " + err.Error())
	}

	return ecdsaPublicKey, nil
}

func (e *ECDSAAuth) importECDSAPrivateKey() (*ecdsa.PrivateKey, error) {
	if e.authConfig == nil {
		return nil, errors.New("auth config not initialized")
	}

	privateKeyBytes := e.authConfig.PrivateKey

	decodedPrivateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBytes)
	if err != nil {
		// Try parsing directly if decoding fails
		ecdsaPrivateKey, parseErr := jwt.ParseECPrivateKeyFromPEM([]byte(privateKeyBytes))
		if parseErr != nil {
			return nil, errors.New("type casting to ecdsa private key failed" + parseErr.Error())
		}

		return ecdsaPrivateKey, nil
	}

	ecdsaPrivateKey, err := jwt.ParseECPrivateKeyFromPEM(decodedPrivateKeyBytes)
	if err != nil {
		return nil, errors.New("parsing ecdsa private key failed: " + err.Error())
	}

	return ecdsaPrivateKey, nil
}

// GenerateJWT creates a JWT token using ECDSA
func (e *ECDSAAuth) GenerateJWT(sub string) (string, error) {
	return e.GenerateJWTWithExpiry(sub, e.config.ExpirationTime)
}

// GenerateJWTWithExpiry creates a JWT token with custom expiry using ECDSA
func (e *ECDSAAuth) GenerateJWTWithExpiry(sub string, expiryDuration time.Duration) (string, error) {
	log := logging.Default()
	aud := jwt.ClaimStrings{"aud"}
	expirationTime := time.Now().Add(expiryDuration)

	claims := &jwt.RegisteredClaims{
		Audience:  aud,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "quickly.com",
		Subject:   sub,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	if e.config.PrivateKey == nil {
		return "", errors.New("private key is not configured")
	}

	signedToken, err := token.SignedString(e.config.PrivateKey)
	if err != nil {
		log.Errorw("signing private key throws error", "error", err)
	}

	return signedToken, err
}

// Verify verifies a JWT token using ECDSA public key
func (e *ECDSAAuth) Verify(token string) (*jwt.RegisteredClaims, error) {
	return e.VerifyWithPublicKey(token, e.config.PublicKey)
}

// VerifyWithPublicKey verifies a JWT token with a specific public key
func (e *ECDSAAuth) VerifyWithPublicKey(token string, publicKey *ecdsa.PublicKey) (*jwt.RegisteredClaims, error) {
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

// SignMessage signs a message using ECDSA private key
func (e *ECDSAAuth) SignMessage(msg []byte) (string, error) {
	privateKey, err := e.importECDSAPrivateKey()
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(msg)

	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies a message signature using ECDSA public key
func (e *ECDSAAuth) VerifySignature(msg []byte, signature string) bool {
	log := logging.Default()

	publicKey, err := e.importECDSAPublicKey()
	if err != nil {
		log.Errorw("importing public key failed", "error", err)
		return false
	}

	// Decode signature
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	// Calculate hash of the message
	hash := sha256.Sum256(msg)

	// Verify the signature
	return ecdsa.VerifyASN1(publicKey, hash[:], decodedSignature)
}

// HashPassword creates a bcrypt hash of the password with salt
func (e *ECDSAAuth) HashPassword(password, salt string) (string, error) {
	passwordWithSalt := password + salt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash and salt
func (e *ECDSAAuth) VerifyPassword(password, hashedPassword, salt string) error {
	passwordWithSalt := password + salt
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(passwordWithSalt))
}

// VerifyPasswordBool is a convenience function that returns a boolean instead of an error
func (e *ECDSAAuth) VerifyPasswordBool(password, hashedPassword, salt string) bool {
	return e.VerifyPassword(password, hashedPassword, salt) == nil
}

// ClientIDFromJWT extracts client ID from JWT token
func (e *ECDSAAuth) ClientIDFromJWT(token string) (clientID string, err error) {
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

	jsonSub := make(map[string]interface{})

	err = json.Unmarshal([]byte(subStr), &jsonSub)
	if err != nil {
		return "", err
	}

	clientID, ok = jsonSub["clientID"].(string)
	if !ok {
		return "", errors.New("error converting clientId")
	}

	return clientID, nil
}
