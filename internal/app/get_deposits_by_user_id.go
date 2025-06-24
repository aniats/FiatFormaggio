package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"log"
)


func (b *Bot) sendDepositsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	// Use unified interceptor for tracing and middleware
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		params := input.(map[string]interface{})
		chatID := params["chatID"].(int64)
		userID := params["userID"].(domain.UserId)
		
		return b.processDepositsCommand(ctx, chatID, userID)
	}
	
	params := map[string]interface{}{
		"chatID": chatID,
		"userID": userID,
	}
	
	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendDepositsCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processDepositsCommand(ctx context.Context, chatID int64, userID domain.UserId) (interface{}, error) {

	deposits, err := b.financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении депозитов для пользователя %s: %v", FormatInteger(int64(userID)), err)
		b.sendMessage(chatID, "❌ Ошибка при получении депозитов. Попробуйте позже.")
		return nil, err
	}

	if len(deposits) == 0 {
		b.sendMessage(chatID, "📭 У вас пока нет сохраненных депозитов.\n\nСоздайте первый депозит: /create_deposit")
		return nil, nil
	}

	text := "🏦 Ваши депозиты:\n\n"
	for i, deposit := range deposits {
		amount := float64(deposit.AmountMinorUnits) / DefaultMinorUnits
		interestRate := float64(deposit.InterestRateBasisPoints) / 100.0

		text += fmt.Sprintf("%d. %s\n", i+1, deposit.Name)
		text += fmt.Sprintf("   💰 %s\n", FormatAmount(amount, deposit.Currency.String()))
		text += fmt.Sprintf("   📈 %s%%/год\n", FormatNumber(interestRate))

		if deposit.ExpirationDate != nil {
			text += fmt.Sprintf("   📅 До: %s\n", deposit.ExpirationDate.Format("02.01.2006"))
		}
		text += "\n"
	}

	text += fmt.Sprintf("📊 Всего депозитов: %s", FormatInteger(int64(len(deposits))))

	b.sendMessage(chatID, text)
	return nil, nil
}

