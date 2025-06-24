package app

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (b *Bot) sendCurrencyRatesCommand(ctx context.Context, chatID int64) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, b.processCurrencyRatesCommand(ctx, chatID)
	}

	params := map[string]interface{}{
		"chat_id": chatID,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendCurrencyRatesCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processCurrencyRatesCommand(ctx context.Context, chatID int64) error {
	rates, err := b.currencyService.GetCurrencyRates(ctx)
	if err != nil {
		apiErr := errors.WrapError(err, errors.CodeExternalAPIError,
			"Failed to fetch currency rates from external service",
			"❌ Ошибка получения курсов валют. Попробуйте позже.")
		b.sendErrorMessage(ctx, chatID, apiErr, "get_currency_rates")
		return apiErr
	}

	if len(rates) == 0 {
		noRatesErr := errors.ErrCurrencyRatesUnavailable
		b.sendMessage(chatID, errors.GetUserMessage(noRatesErr))
		return noRatesErr
	}

	lastUpdateText := "сегодня"
	if len(rates) > 0 {
		lastUpdateText = rates[0].UpdatedAt.Format("02.01.2006 15:04")
	}

	text := fmt.Sprintf("💱 Курсы валют ЦБ РФ (обновлено: %s):\n\n", lastUpdateText)

	majorCurrencies := []domain.CurrencyName{domain.USD, domain.EUR, domain.GBP, domain.JPY, domain.CNY}
	majorRatesShown := make(map[domain.CurrencyName]bool)

	for _, currencyCode := range majorCurrencies {
		for _, rate := range rates {
			if rate.Currency == currencyCode {
				unitRate := float64(rate.RateMinorUnits) / 100.0
				text += fmt.Sprintf("%s: %s ₽\n", rate.Currency, FormatRate(unitRate))
				majorRatesShown[currencyCode] = true
				break
			}
		}
	}

	text += "\n📈 Другие валюты:\n"

	for _, rate := range rates {
		if !majorRatesShown[rate.Currency] {
			unitRate := float64(rate.RateMinorUnits) / 100.0
			text += fmt.Sprintf("%s: %s ₽\n", rate.Currency, FormatRate(unitRate))
		}
	}

	text += fmt.Sprintf("\n📊 Всего валют: %s", FormatInteger(int64(len(rates))))

	b.sendMessage(chatID, text)
	return nil
}
