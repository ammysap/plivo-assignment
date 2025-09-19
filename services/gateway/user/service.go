package user

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ammysap/plivo-pub-sub/libraries/auth"
	"golang.org/x/crypto/bcrypt"
)

// Service interface for user operations
type Service interface {
	Register(username, password string) (*User, error)
	Login(username, password string) (*User, error)
	GetUserByID(userID string) (*User, error)
	GetUserByUsername(username string) (*User, error)
}
type service struct {
	users     map[string]*User // username -> user
	usersByID map[string]*User // user_id -> user
	mu        sync.RWMutex
}

// NewService creates a new user service
func NewService() Service {
	return &service{
		users:     make(map[string]*User),
		usersByID: make(map[string]*User),
	}
}

// Register creates a new user
func (s *service) Register(username, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[username]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID, err := generateUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Create user
	user := &User{
		ID:             userID,
		Username:       username,
		HashedPassword: string(hashedPassword),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Store user with hashed password
	s.users[username] = user
	s.usersByID[userID] = user

	return user, nil
}

// Login authenticates a user
func (s *service) Login(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if user exists
	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password against stored hash
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(userID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByID[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *service) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// generateUserID generates a random user ID
func generateUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateJWTToken generates a JWT token for the user
func GenerateJWTToken(user *User) (string, error) {
	// Generate token using the auth library
	// The auth library uses the user ID as the subject
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
