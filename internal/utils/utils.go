package utils

import (
	"context"
	"log"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"go.uber.org/zap"
)

type contextKey string

const LoggerContextKey contextKey = "requestLogger"

// ContextLogger returns the logger stored in context, or a default logger if none exists.
func ContextLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(LoggerContextKey).(*zap.Logger)
	if !ok || logger == nil {
		return GetLogger(config.AppEnvLocal)
	}
	return logger
}

// GetLogger initializes and returns a zap.Logger based on the application environment.
func GetLogger(appEnv config.AppEnv) *zap.Logger {
	logger, err := zap.NewProduction()

	if appEnv != config.AppEnvProduction {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("failed to setup logger %v", err)
		return nil
	}
	return logger
}

// LoggerWithKey adds a field to the logger in the context and returns the updated context and logger.
func LoggerWithKey(ctx context.Context, field zap.Field) (context.Context, *zap.Logger) {
	logger := ContextLogger(ctx)
	logger = logger.With(field)
	ctx = context.WithValue(ctx, LoggerContextKey, logger)
	return ctx, logger
}
