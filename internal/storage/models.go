package storage

import "github.com/shopspring/decimal"

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

// Represents in-mem model for database object
type Account struct {
	ID      string          `json:"id"`
	Balance decimal.Decimal `json:"balance"`
}
