package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cursed-ninja/internal-transfers-system/internal/config"
	"github.com/cursed-ninja/internal-transfers-system/internal/storage"
	"github.com/cursed-ninja/internal-transfers-system/internal/storage/mocks"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHealthHandler(t *testing.T) {
	s := Server{
		cfg:   &config.Config{},
		store: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.HealthHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *mocks.MockStorage)
		expectedStatus int
	}{
		{
			name: "success",
			body: `{"account_id":"acc-1","initial_balance":"100.00"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().CreateAccount(gomock.Any(), "acc-1", decimal.RequireFromString("100.00")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json",
			body:           `{not json`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing account id",
			body:           `{"initial_balance":"100"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing balance",
			body:           `{"account_id":"acc-1"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed balance",
			body:           `{"account_id":"acc-1","initial_balance":"xyz"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate account",
			body: `{"account_id":"acc-1","initial_balance":"100"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().CreateAccount(gomock.Any(), "acc-1", decimal.RequireFromString("100")).Return(errors.New(storage.ErrAccountExists))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal server error",
			body: `{"account_id":"acc-1","initial_balance":"100"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().CreateAccount(gomock.Any(), "acc-1", decimal.RequireFromString("100")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockStorage := mocks.NewMockStorage(mockCtrl)

			s := Server{
				cfg:   &config.Config{},
				store: mockStorage,
			}

			if tc.mockSetup != nil {
				tc.mockSetup(mockStorage)
			}

			req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()

			s.CreateAccount(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetAccountDetails(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		mockSetup      func(m *mocks.MockStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "success",
			accountID: "acc-1",
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().GetAccountDetails(gomock.Any(), "acc-1").Return(&storage.Account{
					ID:      "acc-1",
					Balance: decimal.RequireFromString("150.50"),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"account_id":"acc-1","balance":"150.5"}`,
		},
		{
			name:      "account not found",
			accountID: "acc-1",
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().GetAccountDetails(gomock.Any(), "acc-1").Return(nil, errors.New(storage.ErrAccountNotFound))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "internal error",
			accountID: "acc-1",
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().GetAccountDetails(gomock.Any(), "acc-1").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockStorage := mocks.NewMockStorage(mockCtrl)

			if tc.mockSetup != nil {
				tc.mockSetup(mockStorage)
			}

			s := Server{
				cfg: &config.Config{
					Env: config.AppEnvLocal,
				},
				store: mockStorage,
			}
			r := mux.NewRouter()
			s.BindRoutes(r)
			path := "/accounts/" + tc.accountID

			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestProcessTransaction(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *mocks.MockStorage)
		expectedStatus int
	}{
		{
			name: "success",
			body: `{"source_account_id":"acc-1","destination_account_id":"acc-2","amount":"50.00"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().ProcessTransaction(gomock.Any(), "acc-1", "acc-2", decimal.RequireFromString("50.00")).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid json",
			body:           `{bad json`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing source_account_id",
			body:           `{"destination_account_id":"acc-2","amount":"10"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing destination_account_id",
			body:           `{"source_account_id":"acc-1","amount":"10"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing amount",
			body:           `{"source_account_id":"acc-1","destination_account_id":"acc-2"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid amount",
			body:           `{"source_account_id":"acc-1","destination_account_id":"acc-2","amount":"xyz"}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "insufficient funds",
			body: `{"source_account_id":"acc-1","destination_account_id":"acc-2","amount":"50"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().ProcessTransaction(gomock.Any(), "acc-1", "acc-2", decimal.RequireFromString("50")).
					Return(errors.New(storage.ErrInsufficientFundsMsg))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "source_account_id not found",
			body: `{"source_account_id":"acc-1","destination_account_id":"acc-2","amount":"10"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().ProcessTransaction(gomock.Any(), "acc-1", "acc-2", decimal.RequireFromString("10")).
					Return(errors.New(storage.ErrSourceAccountMsg))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "destination_account_id not found",
			body: `{"source_account_id":"acc-1","destination_account_id":"acc-2","amount":"10"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().ProcessTransaction(gomock.Any(), "acc-1", "acc-2", decimal.RequireFromString("10")).
					Return(errors.New(storage.ErrDestinationAccountMsg))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "internal server error",
			body: `{"source_account_id":"acc-1","destination_account_id":"acc-2","amount":"20"}`,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().ProcessTransaction(gomock.Any(), "acc-1", "acc-2", decimal.RequireFromString("20")).
					Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockStorage := mocks.NewMockStorage(mockCtrl)

			s := Server{store: mockStorage}

			if tc.mockSetup != nil {
				tc.mockSetup(mockStorage)
			}

			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()

			s.ProcessTransaction(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
