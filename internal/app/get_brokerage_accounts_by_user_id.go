package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)


func (b *Bot) handleBrokerageAccountsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
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
		amount := float64(account.AmountMinorUnits) / defaultMinorUnits

		text += fmt.Sprintf("%d. %s\n", i+1, account.Name)
		text += fmt.Sprintf("   💰 %s\n", formatAmount(amount, account.Currency.String()))
		
		if account.Broker != nil {
			text += fmt.Sprintf("   🏦 %s\n", *account.Broker)
		}
		
		text += fmt.Sprintf("   📊 %s\n", formatAccountType(account.AccountType))
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего счетов: %d", len(accounts))

	b.sendMessage(chatID, text)
}

func formatAccountType(accountType domain.BrokerageType) string {
	switch accountType {
	case domain.Regular:
		return "Обычный"
	case domain.IIS:
		return "ИИС"
	case domain.IIS3:
		return "ИИС-3"
	case domain.IRA:
		return "ИРА"
	case domain.Margin:
		return "Маржинальный"
	default:
		return string(accountType)
	}
}