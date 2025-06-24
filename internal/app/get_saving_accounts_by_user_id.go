package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)


func (b *Bot) sendSavingAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.sendSavingAccountsCommand")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(userID)),
		attribute.Int64("chat.id", chatID),
	)

	accounts, err := b.financeService.GetSavingAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении накопительных счетов для пользователя %s: %v", FormatInteger(int64(userID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении накопительных счетов. Попробуйте позже.")
		return
	}

	if len(accounts) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных накопительных счетов.\n\nСоздайте первый счет: /create_saving_account")
		return
	}

	text := "💰 Ваши накопительные счета:\n\n"
	for i, account := range accounts {
		amount := float64(account.AmountMinorUnits) / DefaultMinorUnits
		interestRate := float64(account.InterestRateBasisPoints) / 100.0

		text += fmt.Sprintf("%d. %s\n", i+1, account.Name)
		text += fmt.Sprintf("   💰 %s\n", FormatAmount(amount, account.Currency.String()))
		
		if interestRate > 0 {
			text += fmt.Sprintf("   📈 %s%%/год\n", FormatNumber(interestRate))
		}

		if account.ExpirationDate != nil {
			text += fmt.Sprintf("   📅 До: %s\n", account.ExpirationDate.Format("02.01.2006"))
		}
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего счетов: %s", FormatInteger(int64(len(accounts))))

	b.sendMessage(chatID, text)
}

