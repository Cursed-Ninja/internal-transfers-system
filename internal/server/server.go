package server

import "github.com/cursed-ninja/internal-transfers-system/internal/config"

type Server struct {
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}
