package domain

type CashHolding struct {
	Id               int64
	UserId           UserId
	Name             string
	AmountMinorUnits int64
	Currency         CurrencyName
}
