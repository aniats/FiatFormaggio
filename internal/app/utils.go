package app

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

const DefaultMinorUnits = 100.0

// FormatNumber adds comma separators to numbers (e.g., 50000 -> 50,000)
func FormatNumber(num float64) string {
	// Convert to string with appropriate decimal places
	str := strconv.FormatFloat(num, 'f', 2, 64)
	
	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	integerPart := parts[0]
	decimalPart := parts[1]
	
	// Add commas to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}
	
	// Remove trailing zeros from decimal part
	decimalPart = strings.TrimRight(decimalPart, "0")
	
	// Combine integer and decimal parts
	if decimalPart == "" {
		return integerPart
	}
	return integerPart + "." + decimalPart
}

// FormatInteger adds comma separators to integers (e.g., 50000 -> 50,000)
func FormatInteger(num int64) string {
	str := strconv.FormatInt(num, 10)
	
	if len(str) > 3 {
		var result strings.Builder
		for i, digit := range str {
			if i > 0 && (len(str)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		return result.String()
	}
	return str
}

// FormatRate formats percentage rates with comma separators (e.g., 1200.50 -> 1,200.50)
func FormatRate(rate float64) string {
	str := strconv.FormatFloat(rate, 'f', 4, 64)
	
	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	integerPart := parts[0]
	decimalPart := parts[1]
	
	// Add commas to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}
	
	// Remove trailing zeros from decimal part
	decimalPart = strings.TrimRight(decimalPart, "0")
	
	// Combine integer and decimal parts
	if decimalPart == "" {
		return integerPart
	}
	return integerPart + "." + decimalPart
}

func FormatAmount(amount float64, currency string) string {
	formattedAmount := FormatNumber(amount)
	switch currency {
	case "RUB":
		return fmt.Sprintf("%s ₽", formattedAmount)
	case "USD":
		return fmt.Sprintf("$%s", formattedAmount)
	case "EUR":
		return fmt.Sprintf("€%s", formattedAmount)
	default:
		return fmt.Sprintf("%s %s", formattedAmount, currency)
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
