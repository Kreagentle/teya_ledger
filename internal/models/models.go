package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	Id        uuid.UUID       `json:"id"`
	Amount    decimal.Decimal `json:"amount"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
}

type LedgerAccount struct {
	Balance      decimal.Decimal `json:"balance"`
	Transactions []Transaction   `json:"transactions"`
}

type TransactionRequest struct {
	Type   string          `json:"type" binding:"required,oneof=deposit withdrawal"`
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type HistoryRequest struct {
	Period string `form:"period"`
}
