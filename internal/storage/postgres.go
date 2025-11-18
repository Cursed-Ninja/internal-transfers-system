package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
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

func (p *PostgressStorage) CreateAccount(ctx context.Context, accountID string, balance decimal.Decimal) error {
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
				return errors.New(ErrAccountExists)
			}
		}
		return errors.New(ErrCreateAccountMsg)
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
			return nil, errors.New(ErrAccountNotFound)
		}
		return nil, errors.New(ErrGetAccountDetailsMsg)
	}

	return &acc, nil
}

func (p *PostgressStorage) ProcessTransaction(ctx context.Context, sourceAccID, destAccID string, amount decimal.Decimal) (err error) {
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
		return errors.New(ErrProcessTransactionMsg)
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
		logger.Error("failed to get destination account details", zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(ErrDestinationAccountMsg)
		}
		return errors.New(ErrProcessTransactionMsg)
	}

	var balanceStr string
	if err = tx.QueryRowContext(ctx, sourceBalanceQuery, sourceAccID).Scan(&balanceStr); err != nil {
		logger.Error("failed to get source account details", zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(ErrSourceAccountMsg)
		}
		return errors.New(ErrProcessTransactionMsg)
	}

	sourceBalance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return errors.New(ErrProcessTransactionMsg)
	}

	if sourceBalance.LessThan(amount) {
		logger.Error("insufficient funds in source account")
		return errors.New(ErrInsufficientFundsMsg)
	}

	if _, err = tx.ExecContext(ctx, withdrawQuery, amount, sourceAccID); err != nil {
		logger.Error("failed to update source account details", zap.Error(err))
		return errors.New(ErrProcessTransactionMsg)
	}

	if _, err = tx.ExecContext(ctx, depositQuery, amount, destAccID); err != nil {
		logger.Error("failed to update destination account details", zap.Error(err))
		return errors.New(ErrProcessTransactionMsg)
	}

	return nil
}
