package app

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (b *Bot) sendCurrencyRatesCommand(ctx context.Context, chatID int64) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.sendCurrencyRatesCommand")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("chat.id", chatID),
	)

	rates, err := b.currencyService.GetCurrencyRates(ctx)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка получения курсов валют. Попробуйте позже.")
		return
	}

	if len(rates) == 0 {
		b.sendMessage(chatID, "📭 Курсы валют временно недоступны.")
		return
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
}
