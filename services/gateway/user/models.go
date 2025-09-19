package user

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email,omitempty"`
	HashedPassword string    `json:"-"` // Don't include in JSON responses
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	Status string `json:"status"`
	User   *User  `json:"user"`
	Token  string `json:"token"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	Status string `json:"status"`
	User   *User  `json:"user"`
	Token  string `json:"token"`
}

// ProfileResponse represents a user profile response
type ProfileResponse struct {
	User *User `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
