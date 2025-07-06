package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)

func (b *Bot) sendBrokerageAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return b.processBrokerageAccountsCommand(ctx, chatID, userID)
	}

	params := map[string]interface{}{
		"user_id": int64(userID),
		"chat_id": chatID,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendBrokerageAccountsCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processBrokerageAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) (interface{}, error) {
	accounts, err := b.financeService.GetBrokerageAccountsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении брокерских счетов для пользователя %s: %v", FormatInteger(int64(userID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении брокерских счетов. Попробуйте позже.")
		return nil, err
	}

	if len(accounts) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных брокерских счетов.\n\nСоздайте первый счет: /create_brokerage_account")
		b.sendMainMenu(chatID)
		return nil, nil
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

	text += fmt.Sprintf("📊 Всего счетов: %s", FormatInteger(int64(len(accounts))))

	b.sendMessage(chatID, text)
	b.sendMainMenu(chatID)
	return nil, nil
}
