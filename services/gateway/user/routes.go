package user

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
	// User profile endpoint (requires authentication)
	authGroup.GET("/users/profile", r.endpoint.GetProfile)
}

// RegisterUnAuthRoutes registers unauthenticated routes
func (r *RouteRegistrar) RegisterUnAuthRoutes(unAuthGroup *gin.RouterGroup) {
	// User registration and login endpoints (no authentication required)
	unAuthGroup.POST("/users/register", r.endpoint.Register)
	unAuthGroup.POST("/users/login", r.endpoint.Login)
}
