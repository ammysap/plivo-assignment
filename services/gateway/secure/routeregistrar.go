package secure

import "github.com/gin-gonic/gin"

type RouteRegistrarInterface interface {
	RegisterAuthRoutes(authGroup *gin.RouterGroup)
	RegisterUnAuthRoutes(unAuthGroup *gin.RouterGroup)
}

type Router struct {
	authGroup       *gin.RouterGroup
	unAuthGroup     *gin.RouterGroup
	routeRegistrars []RouteRegistrarInterface
}

func NewRouter(authGroup, unAuthGroup *gin.RouterGroup) *Router {
	return &Router{
		authGroup:   authGroup,
		unAuthGroup: unAuthGroup,
	}
}

func (r *Router) RegisterRegistrars(
	routeRegistrars ...RouteRegistrarInterface,
) {
	r.routeRegistrars = routeRegistrars
}

func (r *Router) RegisterRoutes() {
	for _, routeRegistrar := range r.routeRegistrars {
		routeRegistrar.RegisterAuthRoutes(r.authGroup)
		routeRegistrar.RegisterUnAuthRoutes(r.unAuthGroup)
	}
}
