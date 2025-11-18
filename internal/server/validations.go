package server

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	ErrMissingAccountID       = errors.New("account_id is required")
	ErrMissingBalance         = errors.New("balance is required")
	ErrInvalidBalance         = errors.New("balance must be a valid decimal number")
	ErrNegativeBalance        = errors.New("balance must be non-negative")
	ErrMissingSourceAccountID = errors.New("source_account_id is required")
	ErrMissingDestAccountID   = errors.New("destination_account_id is required")
	ErrMissingAmount          = errors.New("amount is required")
	ErrInvalidAmount          = errors.New("amount must be a valid decimal number")
	ErrNonPositiveAmount      = errors.New("amount must be positive")
)

func ValidateCreateAccount(req *createAccountRequest) (decimal.Decimal, error) {
	req.AccountID = strings.TrimSpace(req.AccountID)
	req.InitialBalance = strings.TrimSpace(req.InitialBalance)

	if req.AccountID == "" {
		return decimal.Zero, ErrMissingAccountID
	}

	if req.InitialBalance == "" {
		return decimal.Zero, ErrMissingBalance
	}

	balance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		return decimal.Zero, ErrInvalidBalance
	}

	if balance.IsNegative() {
		return decimal.Zero, ErrNegativeBalance
	}

	return balance, nil
}

func ValidateProcessTransaction(req *processTransactionRequest) (decimal.Decimal, error) {
	req.SourceAccID = strings.TrimSpace(req.SourceAccID)
	req.DestAccID = strings.TrimSpace(req.DestAccID)
	req.Amount = strings.TrimSpace(req.Amount)

	if req.SourceAccID == "" {
		return decimal.Zero, ErrMissingSourceAccountID
	}

	if req.DestAccID == "" {
		return decimal.Zero, ErrMissingDestAccountID
	}

	if req.Amount == "" {
		return decimal.Zero, ErrMissingAmount
	}

	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return decimal.Zero, ErrInvalidAmount
	}

	if !amt.IsPositive() {
		return decimal.Zero, ErrNonPositiveAmount
	}

	return amt, nil
}
