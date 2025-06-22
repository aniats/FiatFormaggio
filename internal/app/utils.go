package app

import (
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

const DefaultMinorUnits = 100.0

func FormatAmount(amount float64, currency string) string {
	switch currency {
	case "RUB":
		return fmt.Sprintf("%.2f ₽", amount)
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

func FormatAccountType(accountType domain.BrokerageType) string {
	switch accountType {
	case domain.Regular:
		return "Обычный"
	case domain.IIS:
		return "ИИС"
	case domain.IIS3:
		return "ИИС-3"
	case domain.IRA:
		return "ИРА"
	case domain.Margin:
		return "Маржинальный"
	default:
		return string(accountType)
	}
}