package auth

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ECDSAConfig struct {
	PrivateKey string `env:"PRIVATE_KEY" env-default:""`
	PublicKey  string `env:"PUBLIC_KEY" env-default:""`
}

type HMACConfig struct {
	SecretKey string `env:"JWT_SECRET_KEY" env-default:""`
}

// Config holds the configuration for the auth module (used by factory)
type Config struct {
	PrivateKey        string `env:"PRIVATE_KEY" env-default:""`
	PublicKey         string `env:"PUBLIC_KEY" env-default:""`
	SecretKey         string `env:"JWT_SECRET_KEY" env-default:""`
	JWTExpirationTime int    `env:"JWT_EXPIRATION_TIME" env-default:"1440"` // in minutes
}

// LoadECDSAConfig loads the configuration from environment variables
func LoadECDSAConfig() *ECDSAConfig {
	var cfg ECDSAConfig
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("error reading auth config: %v", err))
	}

	// Validate required fields - at least one auth method must be configured
	if cfg.PrivateKey == "" || cfg.PublicKey == "" {
		panic("PRIVATE_KEY and PUBLIC_KEY environment variables are required")
	}

	return &cfg
}

func LoadHMACConfig() *HMACConfig {
	var cfg HMACConfig
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("error reading auth config: %v", err))
	}

	// Validate required fields - at least one auth method must be configured
	if cfg.SecretKey == "" {
		panic("SECRET_KEY environment variable is required")
	}

	return &cfg
}

// GetExpirationTime returns the JWT expiration time as a Duration
func GetExpirationTime() time.Duration {
	return time.Duration(1440) * time.Minute
}
