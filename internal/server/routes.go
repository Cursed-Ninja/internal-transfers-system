package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) BindRoutes(r *mux.Router) {
	r.Handle("/health", s.loggingMiddleware(http.HandlerFunc(s.HealthHandler))).Methods(http.MethodGet)
	r.Handle("/accounts", s.loggingMiddleware(http.HandlerFunc(s.CreateAccount))).Methods(http.MethodPost)
	r.Handle("/accounts/{accountID}", s.loggingMiddleware(http.HandlerFunc(s.GetAccountDetails))).Methods(http.MethodGet)
	r.Handle("/transactions", s.loggingMiddleware(http.HandlerFunc(s.ProcessTransaction))).Methods(http.MethodPost)
}
