package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)


func (b *Bot) sendCashHoldingsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.sendCashHoldingsCommand")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(userID)),
		attribute.Int64("chat.id", chatID),
	)

	holdings, err := b.financeService.GetCashHoldingsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении наличных счетов для пользователя %s: %v", FormatInteger(int64(userID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении наличных счетов. Попробуйте позже.")
		return
	}

	if len(holdings) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных наличных счетов.\n\nСоздайте первый счет: /create_cash_holding")
		return
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
}

