package domain

import "time"

type AccountType string

const (
	Deposit      AccountType = "deposit"
	Savings      AccountType = "savings"
	Brokerage    AccountType = "brokerage"
	CashCurrency AccountType = "cash_currency"
)

type Account struct {
	ID        int
	Name      string
	Type      AccountType
	Balance   float64
	Currency  string
	Interest  float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID          int
	AccountID   int
	Amount      float64
	Description string
	CreatedAt   time.Time
}

type Profit struct {
	Date      time.Time
	Amount    float64
	AccountID int
}
