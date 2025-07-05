package domain

type BrokerageAccount struct {
	Id               int64
	UserId           UserId
	Name             string
	AmountMinorUnits int64
	Broker           *string
	AccountType      BrokerageType
	Currency         CurrencyName
}
