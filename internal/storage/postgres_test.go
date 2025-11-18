package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// newTestStorage creates a PostgressStorage instance with sqlmock and returns a cleanup function.
func newTestStorage(t *testing.T) (*PostgressStorage, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return &PostgressStorage{db: db}, mock, func() { db.Close() }
}

// TestCreateAccountSuccess validates account creation scenarios, including success and duplicate account errors.
func TestCreateAccountSuccess(t *testing.T) {
	tests := []struct {
		name        string
		args        []driver.Value
		result      driver.Result
		err         error
		expectedErr string
	}{
		{
			name:   "success",
			args:   []driver.Value{"acc-1", "100"},
			result: sqlmock.NewResult(1, 1),
			err:    nil,
		},
		{
			name:        "duplicate account",
			args:        []driver.Value{"acc-1", "100"},
			err:         &pq.Error{Code: "23505"},
			expectedErr: ErrAccountExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store, mock, cleanup := newTestStorage(t)
			defer cleanup()

			expect := mock.ExpectExec(`INSERT INTO accounts`).WithArgs(tc.args...)
			if tc.err != nil {
				expect.WillReturnError(tc.err)
			} else {
				expect.WillReturnResult(tc.result)
			}

			err := store.CreateAccount(ctx, "acc-1", decimal.RequireFromString("100.0"))

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestGetAccountDetails validates retrieval of account details for existing and missing accounts.
func TestGetAccountDetails(t *testing.T) {
	tests := []struct {
		name        string
		accountID   string
		prepare     func(sqlmock.Sqlmock)
		expectedAcc *Account
		expectedErr string
	}{
		{
			name:      "success",
			accountID: "acc-1",
			prepare: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "balance"}).AddRow("acc-1", decimal.RequireFromString("250.5"))
				m.ExpectQuery(`SELECT id, balance FROM accounts`).WithArgs("acc-1").WillReturnRows(rows)
			},
			expectedAcc: &Account{ID: "acc-1", Balance: decimal.RequireFromString("250.5")},
		},
		{
			name:      "not found",
			accountID: "missing",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, balance FROM accounts`).WithArgs("missing").WillReturnError(sql.ErrNoRows)
			},
			expectedErr: ErrAccountNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store, mock, cleanup := newTestStorage(t)
			defer cleanup()

			tc.prepare(mock)

			acc, err := store.GetAccountDetails(ctx, tc.accountID)
			if tc.expectedErr != "" {
				assert.Nil(t, acc)
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAcc, acc)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestProcessTransaction validates transaction processing scenarios, including successful transfers,
// insufficient funds, missing accounts, and update errors.
func TestProcessTransaction(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(sqlmock.Sqlmock)
		amount      string
		expectedErr string
	}{
		{
			name: "success",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT 1 FROM accounts`).WithArgs("dest").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				m.ExpectQuery(`SELECT balance FROM accounts`).WithArgs("source").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.RequireFromString("500.0")))
				m.ExpectExec(`UPDATE accounts SET balance = balance -`).WithArgs(decimal.RequireFromString("200.0"), "source").WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectExec(`UPDATE accounts SET balance = balance +`).WithArgs(decimal.RequireFromString("200.0"), "dest").WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectCommit()
			},
			amount: "200.0",
		},
		{
			name: "destination missing",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT 1 FROM accounts`).WithArgs("dest").WillReturnError(sql.ErrNoRows)
				m.ExpectRollback()
			},
			amount:      "200.0",
			expectedErr: ErrDestinationAccountMsg,
		},
		{
			name: "source missing",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT 1 FROM accounts`).WithArgs("dest").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				m.ExpectQuery(`SELECT balance FROM accounts`).WithArgs("source").WillReturnError(sql.ErrNoRows)
				m.ExpectRollback()
			},
			amount:      "200.0",
			expectedErr: ErrSourceAccountMsg,
		},
		{
			name: "insufficient funds",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT 1 FROM accounts`).WithArgs("dest").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				m.ExpectQuery(`SELECT balance FROM accounts`).WithArgs("source").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.RequireFromString("50.0")))
				m.ExpectRollback()
			},
			amount:      "100.0",
			expectedErr: ErrInsufficientFundsMsg,
		},
		{
			name: "withdraw update error",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT 1 FROM accounts`).WithArgs("dest").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				m.ExpectQuery(`SELECT balance FROM accounts`).WithArgs("source").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.RequireFromString("500.0")))
				m.ExpectExec(`UPDATE accounts SET balance = balance -`).WithArgs(decimal.RequireFromString("100.0"), "source").WillReturnError(errors.New("update source error"))
				m.ExpectRollback()
			},
			amount:      "100.0",
			expectedErr: ErrProcessTransactionMsg,
		},
		{
			name: "deposit update error",
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT 1 FROM accounts`).WithArgs("dest").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				m.ExpectQuery(`SELECT balance FROM accounts`).WithArgs("source").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(decimal.RequireFromString("500.0")))
				m.ExpectExec(`UPDATE accounts SET balance = balance -`).WithArgs(decimal.RequireFromString("100.0"), "source").WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectExec(`UPDATE accounts SET balance = balance +`).WithArgs(decimal.RequireFromString("100.0"), "dest").WillReturnError(errors.New("update dest error"))
				m.ExpectRollback()
			},
			amount:      "100.0",
			expectedErr: ErrProcessTransactionMsg,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store, mock, cleanup := newTestStorage(t)
			defer cleanup()

			tc.prepare(mock)

			err := store.ProcessTransaction(ctx, "source", "dest", decimal.RequireFromString((tc.amount)))
			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
