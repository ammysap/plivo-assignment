package websocket

import (
	"context"
	"net/http"

	"github.com/ammysap/plivo-pub-sub/libraries/auth"
	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ctxKey string

const (
	ctxKeyUserID ctxKey = "user_id"
	ctxKeyClaims ctxKey = "claims"
)

// endpoint implements the Endpoint interface
type Endpoint interface {
	HandleWebSocket(c *gin.Context)
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

// HandleWebSocket handles WebSocket connections
func (e *endpoint) HandleWebSocket(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.WithContext(ctx)

	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		log.Warnw("WebSocket connection attempted without token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// Validate JWT token
	claims, err := auth.Verify(token)
	if err != nil {
		log.Warnw("Invalid token provided for WebSocket connection", "error", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Log successful authentication
	log.Infow("WebSocket connection authenticated", "user_id", claims.Subject)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorw("Failed to upgrade WebSocket connection", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	ctx = context.WithValue(ctx, ctxKeyUserID, claims.Subject)
	ctx = context.WithValue(ctx, ctxKeyClaims, claims)

	e.service.HandleWebSocketConnection(conn, ctx)
}
