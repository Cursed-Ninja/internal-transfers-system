package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/migrations"
	"github.com/cursed-ninja/internal-transfers-system/internal/server"
	"github.com/cursed-ninja/internal-transfers-system/internal/storage"
	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	appEnv := config.GetEnv()
	logger := utils.GetLogger(appEnv)

	err := config.InitViper(appEnv)
	if err != nil {
		logger.Fatal("failed to initialize viper", zap.Error(err))
	}

	cfg := config.NewConfig(appEnv)

	pgClient, err := storage.NewPostgressManager(ctx, cfg.PostgresConfig)
	if err != nil {
		logger.Fatal("failed to initialze postgres", zap.Error(err))
	}

	if err := migrations.Run(ctx, pgClient.DB(), logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	server := server.NewServer(cfg, pgClient)

	httpSrv := startServer(cfg, server, logger)

	// Listen for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop
	logger.Info("Shutdown signal received")

	stopServer(ctx, httpSrv, logger)
}

func startServer(cfg *config.Config, server *server.Server, log *zap.Logger) *http.Server {
	r := mux.NewRouter()
	server.BindRoutes(r)

	httpSrv := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info("HTTP server listening", zap.String("port", cfg.Port))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to listen and serve", zap.Error(err))
		}
	}()
	return httpSrv
}

func stopServer(ctx context.Context, httpSrv *http.Server, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error("Failed to shutdown HTTP server", zap.Error(err))
	}
	log.Info("Server stopped")
}
