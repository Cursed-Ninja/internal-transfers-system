package storage

import (
	"context"

	"github.com/shopspring/decimal"
)

// Storage defines the interface for account and transaction operations.
type Storage interface {
	CreateAccount(ctx context.Context, accountID string, balance decimal.Decimal) error
	GetAccountDetails(ctx context.Context, accountID string) (*Account, error)
	ProcessTransaction(ctx context.Context, sourceAccID string, destAccID string, amount decimal.Decimal) error
}
