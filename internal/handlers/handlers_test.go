package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/Kreagentle/teya_ledger/internal/models"
)

func TestCreateLedgerAccount(t *testing.T) {
	accountsStorage := make(map[uuid.UUID]models.LedgerAccount)
	r := NewRouter(&AccountStorage{LedgerAccountsStorage: accountsStorage})

	tests := []struct {
		name         string
		requestURL   *http.Request
		expectedCode int
	}{
		{
			name:         "Success",
			requestURL:   httptest.NewRequest(http.MethodPost, "/account", nil),
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.requestURL)

			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			if tt.expectedCode == http.StatusCreated {
				assert.NotEmpty(t, response["id"])
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	accountsStorage := make(map[uuid.UUID]models.LedgerAccount)
	accountID := uuid.New()
	accountsStorage[accountID] = models.LedgerAccount{
		Balance: decimal.NewFromFloat(100.0),
	}

	r := NewRouter(&AccountStorage{LedgerAccountsStorage: accountsStorage})

	tests := []struct {
		name            string
		requestURL      string
		expectedCode    int
		expectedBalance string
		expectedError   string
	}{
		{
			name:            "Success",
			requestURL:      "/balance/" + accountID.String(),
			expectedCode:    http.StatusOK,
			expectedBalance: "100",
		},
		{
			name:          "Failure: invalid id",
			requestURL:    "/balance/invalid-id",
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid account ID format",
		},
		{
			name:          "Failure: account not found",
			requestURL:    "/balance/" + uuid.New().String(),
			expectedCode:  http.StatusNotFound,
			expectedError: "Account not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.requestURL, nil)
			assert.NoError(t, err, "Error creating HTTP request")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, tt.expectedBalance, response["balance"])
			} else {
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestGetHistory(t *testing.T) {
	accountsStorage := make(map[uuid.UUID]models.LedgerAccount)
	accountID := uuid.New()

	// define the transactions for the account
	depositTransaction := models.Transaction{
		Id:        uuid.New(),
		Type:      "deposit",
		Amount:    decimal.NewFromFloat(50.0),
		Timestamp: time.Now().Add(-time.Hour),
	}
	withdrawalTransaction := models.Transaction{
		Id:        uuid.New(),
		Type:      "withdrawal",
		Amount:    decimal.NewFromFloat(20.0),
		Timestamp: time.Now().Add(-30 * time.Minute),
	}

	account := models.LedgerAccount{
		Balance:      decimal.NewFromFloat(100.0),
		Transactions: []models.Transaction{depositTransaction, withdrawalTransaction},
	}
	accountsStorage[accountID] = account

	r := NewRouter(&AccountStorage{LedgerAccountsStorage: accountsStorage})

	tests := []struct {
		name                 string
		requestURL           string
		expectedCode         int
		expectedLength       int
		expectedError        string
		expectedTransactions []models.Transaction
	}{
		{
			name:           "Success: get all transactions",
			requestURL:     "/history/" + accountID.String(),
			expectedCode:   http.StatusOK,
			expectedLength: 2,
			expectedTransactions: []models.Transaction{
				depositTransaction,
				withdrawalTransaction,
			},
		},
		{
			name:           "Success: get transactions within time period",
			requestURL:     "/history/" + accountID.String() + "?period=45m",
			expectedCode:   http.StatusOK,
			expectedLength: 1,
			expectedTransactions: []models.Transaction{
				withdrawalTransaction,
			},
		},
		{
			name:                 "Success: no transactions within time period",
			requestURL:           "/history/" + accountID.String() + "?period=10m",
			expectedCode:         http.StatusOK,
			expectedLength:       0,
			expectedTransactions: []models.Transaction{},
		},
		{
			name:          "Failure: invalid account id format",
			requestURL:    "/history/invalid-uuid",
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid account id format",
		},
		{
			name:          "Failure: invalid duration format",
			requestURL:    "/history/" + accountID.String() + "?period=invalid",
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid duration format",
		},
		{
			name:          "Failure: account not found",
			requestURL:    "/history/" + uuid.New().String(),
			expectedCode:  http.StatusNotFound,
			expectedError: "Account not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.requestURL, nil)
			assert.NoError(t, err, "Error creating HTTP request")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				// for success cases, check the length of transactions
				var response []models.Transaction
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedLength)

				// compare transactions
				for i, expectedTx := range tt.expectedTransactions {
					assert.Equal(t, expectedTx.Id, response[i].Id, "Transaction ID mismatch")
					assert.Equal(t, expectedTx.Type, response[i].Type, "Transaction Type mismatch")
					assert.True(t, expectedTx.Amount.Equal(response[i].Amount), "Transaction Amount mismatch")
					assert.WithinDuration(t, expectedTx.Timestamp, response[i].Timestamp, time.Second, "Transaction Timestamp mismatch")
				}
			} else {
				// for error cases, check the error message
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestSubmitTransaction(t *testing.T) {
	accountID := uuid.New()
	// function to initialize the account storage to the initial state
	initAccountsStorage := func() *AccountStorage {
		// initialize the account storage
		account := models.LedgerAccount{
			Balance:      decimal.NewFromFloat(100.0),
			Transactions: []models.Transaction{},
		}
		accountsStorage := make(map[uuid.UUID]models.LedgerAccount)
		accountsStorage[accountID] = account

		return &AccountStorage{LedgerAccountsStorage: accountsStorage}
	}

	tests := []struct {
		name            string
		requestURL      string
		body            models.TransactionRequest
		expectedCode    int
		expectedBalance decimal.Decimal
		expectedError   string
	}{
		{
			name:            "Success: deposit",
			requestURL:      "/accounts/" + accountID.String() + "/transactions",
			body:            models.TransactionRequest{Type: "deposit", Amount: decimal.NewFromFloat(50.0)},
			expectedCode:    http.StatusOK,
			expectedBalance: decimal.NewFromFloat(150.0),
		},
		{
			name:            "Success: withdrawal",
			requestURL:      "/accounts/" + accountID.String() + "/transactions",
			body:            models.TransactionRequest{Type: "withdrawal", Amount: decimal.NewFromFloat(30.0)},
			expectedCode:    http.StatusOK,
			expectedBalance: decimal.NewFromFloat(70.0),
		},
		{
			name:          "Failure: insufficient balance",
			requestURL:    "/accounts/" + accountID.String() + "/transactions",
			body:          models.TransactionRequest{Type: "withdrawal", Amount: decimal.NewFromFloat(200.0)},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Insufficient balance",
		},
		{
			name:          "Failure: wrong type",
			requestURL:    "/accounts/" + accountID.String() + "/transactions",
			body:          models.TransactionRequest{Type: "deposits", Amount: decimal.NewFromFloat(50.0)},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name:          "Failure: negative number",
			requestURL:    "/accounts/" + accountID.String() + "/transactions",
			body:          models.TransactionRequest{Type: "deposit", Amount: decimal.NewFromFloat(-50.0)},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Amount must be greater than 0",
		},
		{
			name:          "Failure: invalid account id format",
			requestURL:    "/accounts/invalid-uuid/transactions",
			body:          models.TransactionRequest{Type: "deposit", Amount: decimal.NewFromFloat(50.0)},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid account id format",
		},
		{
			name:          "Failure: account not found",
			requestURL:    "/accounts/" + uuid.New().String() + "/transactions",
			body:          models.TransactionRequest{Type: "deposit", Amount: decimal.NewFromFloat(50.0)},
			expectedCode:  http.StatusNotFound,
			expectedError: "Account not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialize a fresh account storage before each test
			accountStorage := initAccountsStorage()

			r := NewRouter(accountStorage)

			// prepare request body
			body, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// prepare request
			req, err := http.NewRequest(http.MethodPost, tt.requestURL, bytes.NewReader(body))
			assert.NoError(t, err, "Error creating HTTP request")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// check status code
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				// for success cases, check the balance
				accountAfterTransaction := accountStorage.LedgerAccountsStorage[accountID]
				assert.Equal(t, tt.expectedBalance.String(), accountAfterTransaction.Balance.String())
			} else {
				// for error cases, check the error message
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}
