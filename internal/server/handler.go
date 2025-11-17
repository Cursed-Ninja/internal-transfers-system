package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cursed-ninja/internal-transfers-system/internal/storage.go"
	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"go.uber.org/zap"
)

type createAccountRequest struct {
	AccountID      string `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

type accountResponse struct {
	ID      string `json:"id"`
	Balance string `json:"balance"`
}

type processTransactionRequest struct {
	SourceAccID string `json:"source_account_id"`
	DestAccID   string `json:"destination_account_id"`
	Amount      string `json:"amount"`
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.ContextLogger(ctx)

	var req createAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to parse request body", zap.Error(err))
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.AccountID == "" {
		logger.Error("missing account_id in request body")
		http.Error(w, "account_id is required", http.StatusBadRequest)
		return
	}

	if req.InitialBalance == "" {
		logger.Error("missing balance in request body")
		http.Error(w, "balance is required", http.StatusBadRequest)
		return
	}

	balance, err := strconv.ParseFloat(req.InitialBalance, 64)
	if err != nil {
		logger.Error("failed to parse balance to float")
		http.Error(w, "balance must be a valid decimal number", http.StatusBadRequest)
		return
	}

	ctx, _ = utils.LoggerWithKey(ctx, zap.String("account_id", req.AccountID))
	ctx, logger = utils.LoggerWithKey(ctx, zap.Float64("balance", balance))

	err = s.store.CreateAccount(ctx, req.AccountID, balance)
	if err != nil {
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

func (s *Server) GetAccountDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.ContextLogger(ctx)

	w.Header().Set("Content-Type", "application/json")

	accountID := r.PathValue("accountID")
	if accountID == "" || accountID == r.URL.Path {
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
		Balance: fmt.Sprint(acc.Balance),
	}

	logger.Info("account details retrieved successfully")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.ContextLogger(ctx)

	var req processTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to parse request body", zap.Error(err))
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.SourceAccID == "" {
		logger.Error("missing source_account_id in request body")
		http.Error(w, "source_account_id is required", http.StatusBadRequest)
		return
	}

	if req.DestAccID == "" {
		logger.Error("missing destination_account_id in request body")
		http.Error(w, "destination_account_id is required", http.StatusBadRequest)
		return
	}

	if req.Amount == "" {
		logger.Error("missing amount in request body")
		http.Error(w, "amount is required", http.StatusBadRequest)
		return
	}

	amt, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		logger.Error("failed to parse amount to float")
		http.Error(w, "amount must be a valid decimal number", http.StatusBadRequest)
		return
	}

	ctx, _ = utils.LoggerWithKey(ctx, zap.String("source_account_id", req.SourceAccID))
	ctx, _ = utils.LoggerWithKey(ctx, zap.String("destination_account_id", req.DestAccID))
	ctx, logger = utils.LoggerWithKey(ctx, zap.Float64("amount", amt))

	err = s.store.ProcessTransaction(ctx, req.SourceAccID, req.DestAccID, amt)
	if err != nil {
		logger.Error("failed to process transaction", zap.Error(err))
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError
		if errorMsg == storage.ErrSourceAccountMsg ||
			errorMsg == storage.ErrDestinationAccountMsg ||
			errorMsg == storage.ErrInsufficientFundsMsg {
			statusCode = http.StatusConflict
		}
		http.Error(w, errorMsg, statusCode)
		return
	}

	logger.Info("transaction processed successfully")
	w.WriteHeader(http.StatusCreated)
}
