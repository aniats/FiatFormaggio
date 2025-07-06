package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)

func (b *Bot) sendCashHoldingsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return b.processCashHoldingsCommand(ctx, chatID, userID)
	}

	params := map[string]interface{}{
		"user_id": int64(userID),
		"chat_id": chatID,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendCashHoldingsCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processCashHoldingsCommand(ctx context.Context, chatID int64, userID domain.UserId) (interface{}, error) {
	holdings, err := b.financeService.GetCashHoldingsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении наличных счетов для пользователя %s: %v", FormatInteger(int64(userID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении наличных счетов. Попробуйте позже.")
		return nil, err
	}

	if len(holdings) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных наличных счетов.\n\nСоздайте первый счет: /create_cash_holding")
		b.sendMainMenu(chatID)
		return nil, nil
	}

	text := "💵 Ваши наличные счета:\n\n"
	for i, holding := range holdings {
		amount := float64(holding.AmountMinorUnits) / DefaultMinorUnits

		text += fmt.Sprintf("%d. %s\n", i+1, holding.Name)
		text += fmt.Sprintf("   💰 %s\n", FormatAmount(amount, holding.Currency.String()))
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего счетов: %s", FormatInteger(int64(len(holdings))))

	b.sendMessage(chatID, text)
	b.sendMainMenu(chatID)
	return nil, nil
}
