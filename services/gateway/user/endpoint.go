package user

import (
	"net/http"

	"github.com/ammysap/plivo-pub-sub/services/gateway/logger"
	"github.com/gin-gonic/gin"
)

// Endpoint interface for user endpoints
type Endpoint interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	GetProfile(c *gin.Context)
}
type endpoint struct {
	service Service
}

// NewEndpoint creates a new endpoint
func NewEndpoint(service Service) Endpoint {
	return &endpoint{
		service: service,
	}
}

// Register handles POST /users/register
func (e *endpoint) Register(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorw("Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Register user
	user, err := e.service.Register(req.Username, req.Password)
	if err != nil {
		if err.Error() == "username already exists" {
			log.Warnw("Username already exists", "username", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		log.Errorw("Error registering user", "error", err.Error(), "username", req.Username)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Generate JWT token
	token, err := GenerateJWTToken(user)
	if err != nil {
		log.Errorw("Error generating token", "error", err.Error(), "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := RegisterResponse{
		Status: "registered",
		User:   user,
		Token:  token,
	}

	log.Infow("User registered successfully", "user_id", user.ID, "username", user.Username)
	c.JSON(http.StatusCreated, response)
}

// Login handles POST /users/login
func (e *endpoint) Login(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorw("Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Login user
	user, err := e.service.Login(req.Username, req.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			log.Warnw("Invalid login attempt", "username", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		log.Errorw("Error logging in user", "error", err.Error(), "username", req.Username)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login user"})
		return
	}

	// Generate JWT token
	token, err := GenerateJWTToken(user)
	if err != nil {
		log.Errorw("Error generating token", "error", err.Error(), "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Status: "logged_in",
		User:   user,
		Token:  token,
	}

	log.Infow("User logged in successfully", "user_id", user.ID, "username", user.Username)
	c.JSON(http.StatusOK, response)
}

// GetProfile handles GET /users/profile
func (e *endpoint) GetProfile(c *gin.Context) {
	_, log, err := logger.GetLoggerFromGinContext(c)
	if err != nil {
		log.Errorw("Error getting logger from gin context", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Errorw("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Errorw("Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	// Get user
	user, err := e.service.GetUserByID(userIDStr)
	if err != nil {
		if err.Error() == "user not found" {
			log.Warnw("User not found", "user_id", userIDStr)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Errorw("Error getting user", "error", err.Error(), "user_id", userIDStr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	response := ProfileResponse{
		User: user,
	}

	log.Infow("User profile retrieved successfully", "user_id", user.ID, "username", user.Username)
	c.JSON(http.StatusOK, response)
}
