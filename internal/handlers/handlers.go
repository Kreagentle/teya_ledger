package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Kreagentle/teya_ledger/internal/models"
)

type AccountStorage struct {
	LedgerAccountsStorage map[uuid.UUID]models.LedgerAccount
}

// CreateLedgerAccount creates a new ledger account
func (st *AccountStorage) CreateLedgerAccount(c *gin.Context) {
	// generate a unique ID for the new account
	generatedId, err := uuid.NewUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate account id", "details": err.Error()})
		return
	}

	st.LedgerAccountsStorage[generatedId] = models.LedgerAccount{
		Balance:      decimal.NewFromInt(0),
		Transactions: []models.Transaction{},
	}

	c.JSON(http.StatusCreated, gin.H{"id": generatedId})
}

// GetBalance retrieves the balance of a ledger
func (st *AccountStorage) GetBalance(c *gin.Context) {
	parsedId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID format", "details": err.Error()})
		return
	}

	// get account balance details
	account, exists := st.LedgerAccountsStorage[parsedId]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": account.Balance.String()})
}

// GetHistory retrieves the transaction history
func (st *AccountStorage) GetHistory(c *gin.Context) {
	var req models.HistoryRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	parsedId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account id format", "details": err.Error()})
		return
	}

	// get account balance details
	account, exists := st.LedgerAccountsStorage[parsedId]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// if there is no time period specified return all transactions
	if req.Period == "" {
		c.JSON(http.StatusOK, account.Transactions)
		return
	}

	dur, err := time.ParseDuration(req.Period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format", "details": err.Error()})
		return
	}

	// filter transactions within time range
	cutoffTime := time.Now().Add(-dur)
	for i := len(account.Transactions) - 1; i >= 0; i-- {
		if cutoffTime.Before(account.Transactions[i].Timestamp) {
			c.JSON(http.StatusOK, account.Transactions[i:])
			return
		}
	}

	c.JSON(http.StatusOK, []models.Transaction{})
}

// SubmitTransaction handles submitting a new transaction
func (st *AccountStorage) SubmitTransaction(c *gin.Context) {
	var req models.TransactionRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.Amount.LessThan(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than 0"})
		return
	}

	parsedId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account id format", "details": err.Error()})
		return
	}

	account, exists := st.LedgerAccountsStorage[parsedId]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// create transaction
	transaction := models.Transaction{
		Id:        uuid.New(),
		Type:      req.Type,
		Amount:    req.Amount,
		Timestamp: time.Now(),
	}

	if transaction.Type == "deposit" {
		account.Balance = account.Balance.Add(transaction.Amount)
	} else if transaction.Type == "withdrawal" {
		if account.Balance.Compare(transaction.Amount) < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return
		}
		account.Balance = account.Balance.Sub(transaction.Amount)
	}

	account.Transactions = append(account.Transactions, transaction)
	st.LedgerAccountsStorage[parsedId] = account

	c.JSON(http.StatusOK, transaction)
}
