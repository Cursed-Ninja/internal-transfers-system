package server

import (
	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/storage"
)

// Server holds the configuration and storage layer for handling HTTP requests.
type Server struct {
	cfg   *config.Config
	store storage.Storage
}

// NewServer creates a new Server instance with the given configuration and storage.
func NewServer(cfg *config.Config, store storage.Storage) *Server {
	return &Server{
		cfg:   cfg,
		store: store,
	}
}
