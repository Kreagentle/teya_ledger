package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter creates a new gin router with all defined routes
func NewRouter(accountsStorage *AccountStorage) http.Handler {
	s := gin.Default()

	// define routes
	s.POST("/account", accountsStorage.CreateLedgerAccount)
	s.GET("/balance/:id", accountsStorage.GetBalance)
	s.GET("/history/:id", accountsStorage.GetHistory)
	s.POST("/accounts/:id/transactions", accountsStorage.SubmitTransaction)

	return s
}
