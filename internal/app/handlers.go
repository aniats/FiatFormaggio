package app

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	command := strings.ToLower(msg.Command())

	switch command {
	case "start", "help":
		b.sendHelp(msg.Chat.ID)
	case "total":
		b.handleTotalBalance(ctx, msg)
	default:
		b.sendMessage(msg.Chat.ID, "Неизвестная команда. Введите /help для списка команд.")
	}
}

func (b *Bot) sendHelp(chatID int64) {
	helpText := `Доступные команды:
				/total - Общий баланс
				/nobroker - Баланс без брокерских счетов
				/nocurrency - Баланс без наличной валюты
				/accounts [type] - Список счетов по типу (deposit, savings, brokerage, cash_currency)
				/sum [type] - Сумма по типу счетов
				/profit [day|month] - Прибыль за день/месяц
				/currency - Текущие курсы валют`

	b.sendMessage(chatID, helpText)
}

func (b *Bot) handleTotalBalance(ctx context.Context, msg *tgbotapi.Message) {
	total, err := b.financeService.GetTotalBalance(ctx)
	if err != nil {
		b.sendMessage(msg.Chat.ID, "error while getting account balance")
		return
	}

	b.sendMessage(msg.Chat.ID, fmt.Sprintf("Total: %.2f RUB", total))
}
