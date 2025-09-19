package websocket

import (
	"github.com/ammysap/plivo-pub-sub/services/gateway/secure"
	"github.com/gin-gonic/gin"
)

// RouteRegistrar implements the secure.RouteRegistrarInterface
type RouteRegistrar struct {
	endpoint Endpoint
}

// NewRouteRegistrar creates a new route registrar
func NewRouteRegistrar(service Service) secure.RouteRegistrarInterface {
	return &RouteRegistrar{
		endpoint: NewEndpoint(service),
	}
}

// RegisterAuthRoutes registers authenticated routes
func (r *RouteRegistrar) RegisterAuthRoutes(authGroup *gin.RouterGroup) {
	// no auth routes
}

// RegisterUnAuthRoutes registers unauthenticated routes
func (r *RouteRegistrar) RegisterUnAuthRoutes(unAuthGroup *gin.RouterGroup) {
	// WebSocket endpoint (unauthenticated for now)
	unAuthGroup.GET("/ws", r.endpoint.HandleWebSocket)
}
