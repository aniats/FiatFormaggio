package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/utils"
	"log"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (b *Bot) sendDepositsCommand(ctx context.Context, chatID int64, userID domain.UserID) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		params := input.(map[string]interface{})
		chatID := params["chatID"].(int64)
		userID := params["userID"].(domain.UserID)

		return b.processDepositsCommand(ctx, chatID, userID)
	}

	params := map[string]interface{}{
		"chatID": chatID,
		"userID": userID,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendDepositsCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processDepositsCommand(ctx context.Context, chatID int64, UserID domain.UserID) (interface{}, error) {

	deposits, err := b.financeService.GetDepositsByUserID(ctx, UserID)
	if err != nil {
		log.Printf("Ошибка при получении депозитов для пользователя %s: %v", utils.FormatInteger(int64(UserID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении депозитов. Попробуйте позже.")
		return nil, err
	}

	if len(deposits) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных депозитов.\n\nСоздайте первый депозит: /create_deposit")
		b.sendMainMenu(chatID)
		return nil, nil
	}

	text := "🏦 Ваши депозиты:\n\n"
	for i, deposit := range deposits {
		amount := float64(deposit.AmountMinorUnits) / utils.DefaultMinorUnits
		interestRate := float64(deposit.InterestRateBasisPoints) / 100.0

		text += fmt.Sprintf("%d. %s\n", i+1, deposit.Name)
		text += fmt.Sprintf("   💰 %s\n", utils.FormatAmount(amount, deposit.Currency.String()))
		text += fmt.Sprintf("   📈 %s%%/год\n", utils.FormatNumber(interestRate))

		if deposit.ExpirationDate != nil {
			text += fmt.Sprintf("   📅 До: %s\n", deposit.ExpirationDate.Format("02.01.2006"))
		}
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего депозитов: %s", utils.FormatInteger(int64(len(deposits))))

	b.sendMessage(chatID, text)
	b.sendMainMenu(chatID)
	return nil, nil
}
