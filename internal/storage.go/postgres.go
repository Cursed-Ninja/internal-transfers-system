package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgressStorage struct {
	db *sql.DB
}

func NewPostgressManager(ctx context.Context, cfg *config.PostgresConfig) (*PostgressStorage, error) {
	db, err := sql.Open("postgres", cfg.ConnStr)
	if err != nil {
		return nil, err
	}

	return &PostgressStorage{
		db: db,
	}, nil
}

func (p *PostgressStorage) DB() *sql.DB {
	return p.db
}

func (p *PostgressStorage) CreateAccount(ctx context.Context, accountID string, balance float64) error {
	const query = `
		INSERT INTO accounts (id, balance)
		VALUES ($1, $2)
	`

	logger := utils.ContextLogger(ctx)

	_, err := p.db.ExecContext(ctx, query, accountID, balance)
	if err != nil {
		logger.Error("Failed to create account", zap.Error(err))
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return fmt.Errorf("Account already exists")
			}
		}
		return fmt.Errorf("internal Server Error: failed to create account")
	}
	return nil
}

func (p *PostgressStorage) GetAccountDetails(ctx context.Context, accountID string) (*Account, error) {
	const query = `
		SELECT id, balance
		FROM accounts
		WHERE id = $1
	`

	logger := utils.ContextLogger(ctx)

	var acc Account
	err := p.db.QueryRowContext(ctx, query, accountID).Scan(&acc.ID, &acc.Balance)
	if err != nil {
		logger.Error("Failed to get account details", zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Account doesn't exist")
		}
		return nil, fmt.Errorf("internal Server Error: failed to get account details")
	}

	return &acc, nil
}

func (p *PostgressStorage) ProcessTransaction(ctx context.Context, sourceAccID, destAccID string, amount float64) (err error) {
	const (
		// Query to check destination Acc exists
		destExistsQuery = `
			SELECT 1
			FROM accounts
			WHERE id = $1
		`
		// Query to check Source Acc exists
		sourceBalanceQuery = `
			SELECT balance
			FROM accounts
			WHERE id = $1
		`
		// Query to update source Acc balance
		withdrawQuery = `
			UPDATE accounts
			SET balance = balance - $1
			WHERE id = $2
		`
		// Query to update destination Acc balance
		depositQuery = `
			UPDATE accounts
			SET balance = balance + $1
			WHERE id = $2
		`
	)
	var tx *sql.Tx

	logger := utils.ContextLogger(ctx)

	tx, err = p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to create transaction", zap.Error(err))
		return fmt.Errorf("internal Server Error: failed to process transaction")
	}

	defer func() {
		if tx == nil {
			return
		}
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	var tmp int
	if err = tx.QueryRowContext(ctx, destExistsQuery, destAccID).Scan(&tmp); err != nil {
		logger.Error("failed to get source account details", zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("destination account not found: %w", err)
		}
		return fmt.Errorf("internal Server Error: failed to process transaction")
	}

	var sourceBalance float64
	if err = tx.QueryRowContext(ctx, sourceBalanceQuery, sourceAccID).Scan(&sourceBalance); err != nil {
		logger.Error("failed to get destination account details", zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("source account not found: %w", err)
		}
		return fmt.Errorf("internal Server Error: failed to process transaction")
	}

	if sourceBalance < amount {
		logger.Error("insufficient funds in source account")
		return fmt.Errorf("insufficient funds in source account %s", sourceAccID)
	}

	if _, err = tx.ExecContext(ctx, withdrawQuery, amount, sourceAccID); err != nil {
		logger.Error("failed to update source account details", zap.Error(err))
		return fmt.Errorf("internal Server Error: failed to process transaction")
	}

	if _, err = tx.ExecContext(ctx, depositQuery, amount, destAccID); err != nil {
		logger.Error("failed to update destination account details", zap.Error(err))
		return fmt.Errorf("internal Server Error: failed to process transaction")
	}

	return nil
}
