package app

import (
	"context"
	"fmt"
	"time"
)

func (b *Bot) handleCurrencyRatesCommand(ctx context.Context, chatID int64) {
	rates, err := b.cbrService.GetCurrencyRates(ctx, time.Now())
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка получения курсов валют. Попробуйте позже.")
		return
	}

	if len(rates) == 0 {
		b.sendMessage(chatID, "📭 Курсы валют временно недоступны.")
		return
	}

	text := fmt.Sprintf("💱 Курсы валют ЦБ РФ на %s:\n\n", time.Now().Format("02.01.2006"))

	majorCurrencies := []string{"USD", "EUR", "GBP", "JPY", "CNY"}
	majorRatesShown := make(map[string]bool)

	for _, currencyCode := range majorCurrencies {
		for _, rate := range rates {
			if rate.CharCode == currencyCode {
				unitRate := rate.Value / float64(rate.Nominal)
				text += fmt.Sprintf("%s: %.4f ₽\n", rate.CharCode, unitRate)
				majorRatesShown[currencyCode] = true
				break
			}
		}
	}

	text += "\n📈 Другие валюты:\n"

	for _, rate := range rates {
		if !majorRatesShown[rate.CharCode] {
			unitRate := rate.Value / float64(rate.Nominal)
			text += fmt.Sprintf("%s: %.4f ₽\n", rate.CharCode, unitRate)
		}
	}

	text += fmt.Sprintf("\n📊 Всего валют: %d", len(rates))

	b.sendMessage(chatID, text)
}
