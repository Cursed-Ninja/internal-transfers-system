package server

import "net/http"

func BindRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/accounts", CreateAccount)
	mux.HandleFunc("/accounts/", GetAccountDetails)
	mux.HandleFunc("/transasctions", ProcessTransaction)
}
