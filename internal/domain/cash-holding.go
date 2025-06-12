package domain

type CashHolding struct {
	User             User
	Name             string
	AmountMinorUnits int64
	CurrencyName     CurrencyName
}
