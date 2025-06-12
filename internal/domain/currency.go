package domain

type CurrencyName string

const (
	USD CurrencyName = "USD"
	EUR CurrencyName = "EUR"
	GBP CurrencyName = "GBP"
	JPY CurrencyName = "JPY"
	RUB CurrencyName = "RUB"
	CNY CurrencyName = "CNY"
	RSD CurrencyName = "RSD"
	XBT CurrencyName = "XBT"
	KZT CurrencyName = "KZT"
)

type CurrencyMinorUnits struct {
	Name          CurrencyName
	NameMinorUnit string
	UnitsPerMajor int64
	Symbol        string
	Description   string
}

type CurrencyRateCBR struct {
	CharCode string
	Name     string
	Value    float64
	Nominal  int
}

type CurrencyRate struct {
	Currency       CurrencyName
	RateMinorUnits int64
	BaseCurrency   CurrencyName
	Source         string
}
