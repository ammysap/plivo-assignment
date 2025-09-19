package topic

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
	authGroup.POST("/topics", r.endpoint.CreateTopic)
	authGroup.DELETE("/topics/:name", r.endpoint.DeleteTopic)
	authGroup.GET("/topics", r.endpoint.ListTopics)
}

// RegisterUnAuthRoutes registers unauthenticated routes
func (r *RouteRegistrar) RegisterUnAuthRoutes(unAuthGroup *gin.RouterGroup) {
	unAuthGroup.GET("/health", r.endpoint.GetHealth)
	unAuthGroup.GET("/stats", r.endpoint.GetStats)
}
