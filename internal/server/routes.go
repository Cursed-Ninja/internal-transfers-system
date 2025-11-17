package server

import "net/http"

func (s *Server) BindRoutes(mux *http.ServeMux) {
	mux.Handle("GET /health", s.loggingMiddleware(http.HandlerFunc(s.HealthHandler)))
	mux.Handle("POST /accounts", s.loggingMiddleware(http.HandlerFunc(s.CreateAccount)))
	mux.Handle("GET /accounts/{accountID}", s.loggingMiddleware(http.HandlerFunc(s.GetAccountDetails)))
	mux.Handle("POST /transactions", s.loggingMiddleware(http.HandlerFunc(s.ProcessTransaction)))
}
