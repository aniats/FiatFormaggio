package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)


func (b *Bot) sendBrokerageAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.sendBrokerageAccountsCommand")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(userID)),
		attribute.Int64("chat.id", chatID),
	)

	accounts, err := b.financeService.GetBrokerageAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении брокерских счетов для пользователя %d: %v", userID, err)
		b.sendMessage(chatID, "❌ Ошибка при получении брокерских счетов. Попробуйте позже.")
		return
	}

	if len(accounts) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных брокерских счетов.\n\nСоздайте первый счет: /create_brokerage_account")
		return
	}

	text := "📈 Ваши брокерские счета:\n\n"
	for i, account := range accounts {
		amount := float64(account.AmountMinorUnits) / DefaultMinorUnits

		text += fmt.Sprintf("%d. %s\n", i+1, account.Name)
		text += fmt.Sprintf("   💰 %s\n", FormatAmount(amount, account.Currency.String()))
		
		if account.Broker != nil {
			text += fmt.Sprintf("   🏦 %s\n", *account.Broker)
		}
		
		text += fmt.Sprintf("   📊 %s\n", FormatAccountType(account.AccountType))
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего счетов: %d", len(accounts))

	b.sendMessage(chatID, text)
}

