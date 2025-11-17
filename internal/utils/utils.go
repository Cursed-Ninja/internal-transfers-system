package utils

import (
	"context"
	"log"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"go.uber.org/zap"
)

type contextKey string

const LoggerContextKey contextKey = "requestLogger"

func ContextLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(LoggerContextKey).(*zap.Logger)
	if !ok || logger == nil {
		return GetLogger(config.AppEnvLocal)
	}
	return logger
}

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
