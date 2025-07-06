package domain

type BrokerageType string

const (
	Regular BrokerageType = "regular"
	IIS     BrokerageType = "iis"
	IIS3    BrokerageType = "iis3"
	IRA     BrokerageType = "ira"
	Margin  BrokerageType = "margin"
)

func (bt BrokerageType) ToDisplayName() string {
	switch bt {
	case Regular:
		return "Обычный"
	case IIS:
		return "ИИС"
	case IIS3:
		return "ИИС-3"
	case IRA:
		return "ИРА"
	case Margin:
		return "Маржинальный"
	default:
		return string(bt)
	}
}
