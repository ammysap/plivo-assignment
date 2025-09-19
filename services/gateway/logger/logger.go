package logger

import (
	"context"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLoggerFromGinContext(
	c *gin.Context,
) (context.Context, *zap.SugaredLogger, error) {
	ctx := c.Request.Context()
	log := logging.WithContext(ctx)
	return ctx, log, nil
}

func GetLoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	return logging.WithContext(ctx)
}
