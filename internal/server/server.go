package server

import (
	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/storage"
)

type Server struct {
	cfg   *config.Config
	store storage.Storage
}

func NewServer(cfg *config.Config, store storage.Storage) *Server {
	return &Server{
		cfg:   cfg,
		store: store,
	}
}
