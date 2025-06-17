package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)

func (b *Bot) sendCashHoldingsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	holdings, err := b.financeService.GetCashHoldingsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении наличных счетов для пользователя %d: %v", userID, err)
		b.sendMessage(chatID, "❌ Ошибка при получении наличных счетов. Попробуйте позже.")
		return
	}

	if len(holdings) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных наличных счетов.\n\nСоздайте первый счет: /create_cash_holding")
		return
	}

	text := "💵 Ваши наличные счета:\n\n"
	for i, holding := range holdings {
		amount := float64(holding.AmountMinorUnits) / defaultMinorUnits

		text += fmt.Sprintf("%d. %s\n", i+1, holding.Name)
		text += fmt.Sprintf("   💰 %s\n", formatAmount(amount, holding.Currency.String()))
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего счетов: %d", len(holdings))

	b.sendMessage(chatID, text)
}