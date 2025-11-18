package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cursed-ninja/internal-transfers-system/internal/storage"
	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Request and Response Structs

type createAccountRequest struct {
	AccountID      string `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

type accountResponse struct {
	ID      string `json:"account_id"`
	Balance string `json:"balance"`
}

type processTransactionRequest struct {
	SourceAccID string `json:"source_account_id"`
	DestAccID   string `json:"destination_account_id"`
	Amount      string `json:"amount"`
}

// HealthHandler returns a simple status for health checks.
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// CreateAccount handles POST /accounts requests to create a new account.
func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.ContextLogger(ctx)

	var req createAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to parse request body", zap.Error(err))
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	balance, err := ValidateCreateAccount(&req)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, _ = utils.LoggerWithKey(ctx, zap.String("account_id", req.AccountID))
	ctx, logger = utils.LoggerWithKey(ctx, zap.String("balance", balance.String()))

	if err := s.store.CreateAccount(ctx, req.AccountID, balance); err != nil {
		logger.Error("failed to create account", zap.Error(err))
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError
		if errorMsg == storage.ErrAccountExists {
			statusCode = http.StatusConflict
		}
		http.Error(w, errorMsg, statusCode)
		return
	}

	logger.Info("account created successfully")
	w.WriteHeader(http.StatusCreated)
}

// GetAccountDetails handles GET /accounts/{accountID} requests.
func (s *Server) GetAccountDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.ContextLogger(ctx)

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	accountID := strings.TrimSpace(vars["accountID"])
	if accountID == "" {
		logger.Error("missing account_id in URL path")
		http.Error(w, "account_id is required in URL path", http.StatusBadRequest)
		return
	}

	ctx, logger = utils.LoggerWithKey(ctx, zap.String("account_id", accountID))

	acc, err := s.store.GetAccountDetails(ctx, accountID)
	if err != nil {
		logger.Error("failed to get account details", zap.Error(err))
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError
		if errorMsg == storage.ErrAccountNotFound {
			statusCode = http.StatusNotFound
		}
		http.Error(w, errorMsg, statusCode)
		return
	}

	response := accountResponse{
		ID:      acc.ID,
		Balance: acc.Balance.String(),
	}

	logger.Info("account details retrieved successfully")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ProcessTransaction handles POST /transactions requests to transfer funds between accounts.
func (s *Server) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.ContextLogger(ctx)

	var req processTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to parse request body", zap.Error(err))
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	amt, err := ValidateProcessTransaction(&req)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, _ = utils.LoggerWithKey(ctx, zap.String("source_account_id", req.SourceAccID))
	ctx, _ = utils.LoggerWithKey(ctx, zap.String("destination_account_id", req.DestAccID))
	ctx, logger = utils.LoggerWithKey(ctx, zap.String("amount", amt.String()))

	if err := s.store.ProcessTransaction(ctx, req.SourceAccID, req.DestAccID, amt); err != nil {
		logger.Error("failed to process transaction", zap.Error(err))
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError
		switch errorMsg {
		case storage.ErrSourceAccountMsg, storage.ErrDestinationAccountMsg:
			statusCode = http.StatusNotFound
		case storage.ErrInsufficientFundsMsg:
			statusCode = http.StatusBadRequest
		}
		http.Error(w, errorMsg, statusCode)
		return
	}

	logger.Info("transaction processed successfully")
	w.WriteHeader(http.StatusCreated)
}
