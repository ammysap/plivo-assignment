package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type LoggerKeyType int

const loggerKey LoggerKeyType = iota

const (
	ten     = 10
	hundred = 100
)

// Custom context type that directly holds the logger
type loggerContext struct {
	context.Context

	ctxLogger *zap.SugaredLogger
}

// Override Value method for faster lookups
func (c *loggerContext) Value(key interface{}) interface{} {
	if key == loggerKey {
		return c.ctxLogger
	}

	return c.Context.Value(key)
}

var logger *zap.SugaredLogger

func NewContext(
	ctx context.Context,
	phone, requestID, serviceName, email string, isAdmin bool,
) context.Context {
	role := "admin"
	if !isAdmin {
		role = "user"
	}

	enrichedLogger := WithContext(ctx).With(
		"phone", phone,
		"requestID", requestID,
		"serviceName", serviceName,
		"role", role,
		"isAdmin", isAdmin,
		"email", email,
	)

	return &loggerContext{
		Context:   ctx,
		ctxLogger: enrichedLogger,
	}
}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return logger
	}

	// Try direct type assertion first (much faster)
	if lc, ok := ctx.(*loggerContext); ok {
		return lc.ctxLogger
	}

	// Fall back to regular lookup for backward compatibility, should remove this eventually
	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return ctxLogger
	}

	return logger
}

func Default() *zap.SugaredLogger {
	return logger
}

func setLogger(l *zap.SugaredLogger) {
	logger = l
}

func LogBuildVersionNumber() {
	if logger == nil {
		return
	}

	buildVersion := os.Getenv("BUILD_VERSION")
	if buildVersion == "" {
		return
	}

	// Log the build version
	logger.Infoln("Build version:", buildVersion)
}

func init() {
	env, ok := os.LookupEnv("LOG_ENV")
	if !ok {
		env = "development"
	}

	var cfg zap.Config

	switch env {
	case "development":
		cfg = zap.NewDevelopmentConfig()
	case "production":
		cfg = zap.NewProductionConfig()

		cfg.Sampling = &zap.SamplingConfig{
			Initial:    ten,
			Thereafter: hundred,
		}
	case "quiet":
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.PanicLevel)
		cfg.Sampling = &zap.SamplingConfig{
			Initial:    ten,
			Thereafter: hundred,
		}
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	baseLogger, _ := cfg.Build()

	logger = baseLogger.Sugar()

	LogBuildVersionNumber()
}
