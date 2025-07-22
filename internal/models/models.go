package models

type Transaction struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type Wallet struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}
