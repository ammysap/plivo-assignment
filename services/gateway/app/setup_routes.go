package app

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/ammysap/plivo-pub-sub/services/gateway/middlewares"
	"github.com/ammysap/plivo-pub-sub/services/gateway/secure"
	"github.com/ammysap/plivo-pub-sub/services/gateway/topic"
	"github.com/ammysap/plivo-pub-sub/services/gateway/user"
	"github.com/ammysap/plivo-pub-sub/services/gateway/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter() (router *gin.Engine, authGroup, unAuthGroup *gin.RouterGroup) {
	router = gin.Default()
	numHours := 12
	allowedOriginsStr, isOrigin := os.LookupEnv("ALLOWED_CORS_ORIGIN")
	allowedMethodsStr, isMethod := os.LookupEnv("ALLOWED_CORS_METHOD")

	if allowedOriginsStr == "" || !isOrigin {
		allowedOriginsStr = "*"
	}

	if allowedMethodsStr == "" || !isMethod {
		allowedMethodsStr = "*"
	}

	allowedOrigins := strings.Split(allowedOriginsStr, ",")
	allowedMethods := strings.Split(allowedMethodsStr, ",")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     allowedMethods,
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           time.Duration(numHours) * time.Hour,
	}))

	authGroup = router.Group(
		"/",
		middlewares.AuthMiddleware(),
	)

	unAuthGroup = router.Group("/")

	return router, authGroup, unAuthGroup
}

func RegisterRoutes(ctx context.Context,
	resolver interface{}, // Can be nil for in-memory pub/sub
) error {
	log := logging.WithContext(ctx)

	log.Info("Registering routes...")

	router, authGroup, unAuthGroup := setupRouter()

	secureRouter := secure.NewRouter(authGroup, unAuthGroup)

	// User service
	log.Info("Creating User service...")
	userService := user.NewService()
	userRouteRegistrar := user.NewRouteRegistrar(userService)

	// Topic management service
	log.Info("Creating Topic service...")
	topicService := topic.NewService()
	topicRouteRegistrar := topic.NewRouteRegistrar(topicService)

	// WebSocket service
	log.Info("Creating WebSocket service...")
	websocketService := websocket.NewService()
	websocketRouteRegistrar := websocket.NewRouteRegistrar(websocketService)

	log.Info("Registering routes...")
	secureRouter.RegisterRegistrars(
		userRouteRegistrar,
		topicRouteRegistrar,
		websocketRouteRegistrar,
	)

	log.Info("Registering all routes...")
	secureRouter.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Info(ctx, "Starting server on port", "port", port)
	return router.Run(":" + port)
}
