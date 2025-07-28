package models

import (
	"time"
)

type Transaction struct {
	ID          int       `json:"id"`
	FromAddress string    `json:"from"`
	ToAddress   string    `json:"to"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

type Wallet struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

type SendRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}
