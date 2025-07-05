package domain

type BrokerageType string

const (
	Regular BrokerageType = "regular"
	IIS     BrokerageType = "iis"
	IIS3    BrokerageType = "iis3"
	IRA     BrokerageType = "ira"
	Margin  BrokerageType = "margin"
)
