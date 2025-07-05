package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)

const defaultMinorUnits = 100.0

func (b *Bot) sendDepositsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	deposits, err := b.financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении депозитов для пользователя %d: %v", userID, err)
		b.sendMessage(chatID, "❌ Ошибка при получении депозитов. Попробуйте позже.")
		return
	}

	if len(deposits) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных депозитов.\n\nСоздайте первый депозит: /create_deposit")
		return
	}

	text := "🏦 Ваши депозиты:\n\n"
	for i, deposit := range deposits {
		amount := float64(deposit.AmountMinorUnits) / defaultMinorUnits
		interestRate := float64(deposit.InterestRateBasisPoints) / 100.0

		text += fmt.Sprintf("%d. %s\n", i+1, deposit.Name)
		text += fmt.Sprintf("   💰 %s\n", formatAmount(amount, deposit.Currency.String()))
		text += fmt.Sprintf("   📈 %.2f%%/год\n", interestRate)

		if deposit.ExpirationDate != nil {
			text += fmt.Sprintf("   📅 До: %s\n", deposit.ExpirationDate.Format("02.01.2006"))
		}
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего депозитов: %d", len(deposits))

	b.sendMessage(chatID, text)
}

func formatAmount(amount float64, currency string) string {
	switch currency {
	case "RUB":
		return fmt.Sprintf("%.2f ₽", amount)
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}
