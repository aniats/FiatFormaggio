package app

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultMinorUnits = 100.0

func FormatNumber(num float64) string {
	str := strconv.FormatFloat(num, 'f', 2, 64)

	parts := strings.Split(str, ".")
	integerPart := parts[0]
	decimalPart := parts[1]
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

	decimalPart = strings.TrimRight(decimalPart, "0")
	if decimalPart == "" {
		return integerPart
	}
	return integerPart + "." + decimalPart
}

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

func FormatRate(rate float64) string {
	str := strconv.FormatFloat(rate, 'f', 4, 64)

	parts := strings.Split(str, ".")
	integerPart := parts[0]
	decimalPart := parts[1]
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

	decimalPart = strings.TrimRight(decimalPart, "0")
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
