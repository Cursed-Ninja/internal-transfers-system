package server

import "net/http"

func (s *Server) BindRoutes(mux *http.ServeMux) {
	mux.Handle("/health", s.loggingMiddleware(http.HandlerFunc(s.HealthHandler)))
	mux.Handle("/accounts", s.loggingMiddleware(http.HandlerFunc(s.CreateAccount)))
	mux.Handle("/accounts/", s.loggingMiddleware(http.HandlerFunc(s.GetAccountDetails)))
	mux.Handle("/transasctions", s.loggingMiddleware(http.HandlerFunc(s.ProcessTransaction)))
}
