package domain

type BrokerageAccount struct {
	ID               int64
	UserID           UserID
	Name             string
	AmountMinorUnits int64
	Broker           *string
	AccountType      BrokerageType
	Currency         CurrencyName
}
