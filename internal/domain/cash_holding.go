package domain

type CashHolding struct {
	ID               int64
	UserID           UserID
	Name             string
	AmountMinorUnits int64
	Currency         CurrencyName
}
