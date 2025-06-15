package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)

func (b *Bot) handleDepositsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	deposits, err := b.financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении депозитов для пользователя %d: %v", userID, err)
		b.sendMessage(chatID, "Ошибка при получении депозитов. Попробуйте позжеы.")
		return
	}

	if len(deposits) == 0 {
		b.sendMessage(chatID, "У вас нет депозитов, сохраненных в базе.")
		return
	}

	text := "Ваши депозиты:\n"
	for _, deposit := range deposits {
		amount := float64(deposit.AmountMinorUnits) / 100.0
		text += fmt.Sprintf("• %s: %.2f %s\n", deposit.Name, amount, deposit.Currency)
	}

	b.sendMessage(chatID, text)
}
