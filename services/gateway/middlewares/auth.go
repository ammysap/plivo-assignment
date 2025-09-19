package middlewares

import (
	"net/http"
	"strings"

	"github.com/ammysap/plivo-pub-sub/libraries/auth"
	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logging.WithContext(ctx)

		authHeader := c.Request.Header["Authorization"]
		if authHeader == nil {
			// no token present
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authValue := authHeader[0]
		if !strings.HasPrefix(authValue, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authValue, "Bearer ")

		claims, err := auth.Verify(token)
		if err != nil {
			log.Errorw("Token verification failed", "error", err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Store the claims in context for later use
		c.Set("claims", claims)
		c.Set("user_id", claims.Subject)

		c.Next()
	}
}
