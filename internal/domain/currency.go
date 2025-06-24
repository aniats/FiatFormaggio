package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/errors"
)

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
	ID             int64        `db:"id"`
	Currency       CurrencyName `db:"currency"`
	RateMinorUnits int64        `db:"rate_minor_units"`
	BaseCurrency   CurrencyName `db:"base_currency"`
	Source         string       `db:"source"`
	UpdatedAt      time.Time    `db:"updated_at"`
}

var currencyToHumanMap = map[CurrencyName]string{
	USD: "US Dollar",
	EUR: "Euro",
	GBP: "British Pound",
	JPY: "Japanese Yen",
	RUB: "Russian Ruble",
	CNY: "Chinese Yuan",
	RSD: "Serbian Dinar",
	XBT: "Bitcoin",
	KZT: "Kazakhstani Tenge",
}

var currencyToHumanRussianMap = map[CurrencyName]string{
	USD: "Доллар США",
	EUR: "Евро",
	GBP: "Британский фунт",
	JPY: "Японская иена",
	RUB: "Российский рубль",
	CNY: "Китайский юань",
	RSD: "Сербский динар",
	XBT: "Биткоин",
	KZT: "Казахстанский тенге",
}

var humanToCurrencyMap = map[string]CurrencyName{
	"us dollar":         USD,
	"usd":               USD,
	"dollar":            USD,
	"euro":              EUR,
	"eur":               EUR,
	"british pound":     GBP,
	"gbp":               GBP,
	"pound":             GBP,
	"japanese yen":      JPY,
	"jpy":               JPY,
	"yen":               JPY,
	"russian ruble":     RUB,
	"rub":               RUB,
	"ruble":             RUB,
	"chinese yuan":      CNY,
	"cny":               CNY,
	"yuan":              CNY,
	"serbian dinar":     RSD,
	"rsd":               RSD,
	"dinar":             RSD,
	"bitcoin":           XBT,
	"xbt":               XBT,
	"btc":               XBT,
	"kazakhstani tenge": KZT,
	"kzt":               KZT,
	"tenge":             KZT,
}

var humanToCurrencyRussianMap = map[string]CurrencyName{
	"доллар сша":          USD,
	"доллар":              USD,
	"американский доллар": USD,
	"евро":                EUR,
	"британский фунт":     GBP,
	"фунт":                GBP,
	"японская иена":       JPY,
	"иена":                JPY,
	"российский рубль":    RUB,
	"рубль":               RUB,
	"китайский юань":      CNY,
	"юань":                CNY,
	"сербский динар":      RSD,
	"динар":               RSD,
	"биткоин":             XBT,
	"биткойн":             XBT,
	"казахстанский тенге": KZT,
	"тенге":               KZT,
}

func (c CurrencyName) ToHuman() string {
	if human, exists := currencyToHumanMap[c]; exists {
		return human
	}
	return string(c)
}

func (c CurrencyName) ToHumanRussian() string {
	if human, exists := currencyToHumanRussianMap[c]; exists {
		return human
	}
	return string(c)
}

func (c CurrencyName) String() string {
	return string(c)
}

func (c CurrencyName) IsValid() bool {
	_, exists := currencyToHumanMap[c]
	return exists
}

func (c CurrencyName) Symbol() string {
	switch c {
	case USD:
		return "$"
	case EUR:
		return "€"
	case GBP:
		return "£"
	case JPY:
		return "¥"
	case RUB:
		return "₽"
	case CNY:
		return "¥"
	case XBT:
		return "₿"
	default:
		return string(c)
	}
}

func CurrencyFromHuman(human string) (CurrencyName, error) {
	normalized := strings.ToLower(strings.TrimSpace(human))

	if currency, exists := humanToCurrencyMap[normalized]; exists {
		return currency, nil
	}

	if currency, exists := humanToCurrencyRussianMap[normalized]; exists {
		return currency, nil
	}

	upperHuman := strings.ToUpper(normalized)
	if currency := CurrencyName(upperHuman); currency.IsValid() {
		return currency, nil
	}

	return "", errors.NewCurrencyError(human)
}

func formatNumberWithCommas(num float64) string {
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

func (c CurrencyName) FormatAmount(amount float64) string {
	formattedAmount := formatNumberWithCommas(amount)
	return fmt.Sprintf("%s %s (%s)", formattedAmount, c.Symbol(), c.ToHuman())
}

func (c CurrencyName) FormatAmountRussian(amount float64) string {
	formattedAmount := formatNumberWithCommas(amount)
	return fmt.Sprintf("%s %s (%s)", formattedAmount, c.Symbol(), c.ToHumanRussian())
}
