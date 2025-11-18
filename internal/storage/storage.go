package storage

import "context"

type Storage interface {
	CreateAccount(ctx context.Context, accountID string, balance float64) error
	GetAccountDetails(ctx context.Context, accountID string) (*Account, error)
	ProcessTransaction(ctx context.Context, sourceAccID string, destAccID string, amount float64) error
}
