package storage

import "github.com/shopspring/decimal"

// Error definitions for storage operations
const (
	ErrAccountExists         = "account already exists"
	ErrCreateAccountMsg      = "internal Server Error: failed to create account"
	ErrAccountNotFound       = "account doesn't exist"
	ErrGetAccountDetailsMsg  = "internal Server Error: failed to get account details"
	ErrProcessTransactionMsg = "internal Server Error: failed to process transaction"
	ErrDestinationAccountMsg = "destination account not found"
	ErrSourceAccountMsg      = "source account not found"
	ErrInsufficientFundsMsg  = "insufficient funds in source account"
)

// Account represents an account in storage, with a unique ID and balance.
type Account struct {
	ID      string          `json:"id"`
	Balance decimal.Decimal `json:"balance"`
}
