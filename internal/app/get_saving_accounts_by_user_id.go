package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)

func (b *Bot) sendSavingAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return b.processSavingAccountsCommand(ctx, chatID, userID)
	}

	params := map[string]interface{}{
		"user_id": int64(userID),
		"chat_id": chatID,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendSavingAccountsCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processSavingAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) (interface{}, error) {
	accounts, err := b.financeService.GetSavingAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении накопительных счетов для пользователя %s: %v", FormatInteger(int64(userID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении накопительных счетов. Попробуйте позже.")
		return nil, err
	}

	if len(accounts) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных накопительных счетов.\n\nСоздайте первый счет: /create_saving_account")
		b.sendMainMenu(chatID)
		return nil, nil
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
	b.sendMainMenu(chatID)
	return nil, nil
}
