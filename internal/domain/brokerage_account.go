package domain

type BrokerageAccount struct {
	User             User
	Name             string
	AmountMinorUnits int64
	Broker           string
	AccountType      BrokerageType
	CurrencyName     CurrencyName
}
