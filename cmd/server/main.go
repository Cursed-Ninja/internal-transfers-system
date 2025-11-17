package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/server"
)

func main() {
	appEnv := config.GetEnv()

	err := config.InitViper(appEnv)
	if err != nil {
		log.Fatalf("failed to initialize viper: %v", err)
	}

	cfg := config.NewConfig(appEnv)

	srv := startServer(cfg)

	// Listen for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop
	log.Println("Shutdown signal received")

	stopServer(srv)
}

func startServer(cfg *config.Config) *http.Server {
	mux := http.NewServeMux()
	server.BindRoutes(mux)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP server listening on %s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v\n", err)
		}
	}()
	return server
}

func stopServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}
	log.Println("Server stopped")
}
