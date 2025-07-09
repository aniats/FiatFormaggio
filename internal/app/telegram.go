package app

import (
	"log"

	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	UpdateOffsetFromBeginning = 0
)

type TelegramBotAPI struct {
	api     *tgBotAPI.BotAPI
	updates tgBotAPI.UpdatesChannel
}

func NewTelegramBotAPI(token string) (BotAPI, error) {
	api, err := tgBotAPI.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	u := tgBotAPI.NewUpdate(UpdateOffsetFromBeginning)

	u.Timeout = UpdateTimeoutSeconds

	updates := api.GetUpdatesChan(u)

	return &TelegramBotAPI{
		api:     api,
		updates: updates,
	}, nil
}

func (t *TelegramBotAPI) GetLastEvents() <-chan tgBotAPI.Update {
	return t.updates
}

func (t *TelegramBotAPI) SendMessage(chatID int64, text string) error {
	msg := tgBotAPI.NewMessage(chatID, text)
	_, err := t.api.Send(msg)
	return err
}

func (t *TelegramBotAPI) SendMessageWithKeyboard(chatID int64, text string, keyboard tgBotAPI.InlineKeyboardMarkup) error {
	msg := tgBotAPI.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := t.api.Send(msg)
	return err
}

func (t *TelegramBotAPI) SetMyCommands(commands []tgBotAPI.BotCommand) error {
	commandsConfig := tgBotAPI.NewSetMyCommands(commands...)
	_, err := t.api.Request(commandsConfig)
	return err
}

func (t *TelegramBotAPI) Close() {
	t.api.StopReceivingUpdates()
	log.Println("Telegram Bot API connection closed")
}

type MockBotAPI struct {
	updates chan tgBotAPI.Update
	sent    []MockMessage
}

type MockMessage struct {
	ChatID int64
	Text   string
}

func (m *MockBotAPI) GetLastEvents() <-chan tgBotAPI.Update {
	return m.updates
}

func (m *MockBotAPI) SendMessage(chatID int64, text string) error {
	m.sent = append(m.sent, MockMessage{
		ChatID: chatID,
		Text:   text,
	})
	return nil
}

func (m *MockBotAPI) SendMessageWithKeyboard(chatID int64, text string, keyboard tgBotAPI.InlineKeyboardMarkup) error {
	m.sent = append(m.sent, MockMessage{
		ChatID: chatID,
		Text:   text,
	})
	return nil
}

func (m *MockBotAPI) SetMyCommands(commands []tgBotAPI.BotCommand) error {
	return nil
}

func (m *MockBotAPI) Close() {
	close(m.updates)
}
