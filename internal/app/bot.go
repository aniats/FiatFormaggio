package app

import (
	"context"
	"log"

	"github.com/aniats/FiatFormaggio/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	botAPI         *tgbotapi.BotAPI
	financeService *service.FinanceService
}

func NewBot(token string, financeService *service.FinanceService) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		botAPI:         botAPI,
		financeService: financeService,
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.botAPI.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			go b.handleMessage(ctx, update.Message)
		}
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.botAPI.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
