package app

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	// UpdateOffsetFromBeginning Update offset constants
	UpdateOffsetFromBeginning = 0  // Get all pending updates from start
	UpdateOffsetOnlyNew       = -1 // Skip pending, get only new updates
	UpdateOffsetResume        = 1  // Base for resuming from specific point
)

type TelegramBotAPI struct {
	api     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

func NewTelegramBotAPI(token string) (BotAPI, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	// FYI: Start from `UpdateOffsetFromBeginning`
	u := tgbotapi.NewUpdate(UpdateOffsetFromBeginning)

	// FYI: Wait up to `UpdateTimeoutSeconds` seconds for new messages
	u.Timeout = UpdateTimeoutSeconds

	updates := api.GetUpdatesChan(u)

	return &TelegramBotAPI{
		api:     api,
		updates: updates,
	}, nil
}

func (t *TelegramBotAPI) GetLastEvents() <-chan tgbotapi.Update {
	return t.updates
}

func (t *TelegramBotAPI) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.api.Send(msg)
	return err
}

func (t *TelegramBotAPI) Close() {
	t.api.StopReceivingUpdates()
	log.Println("Telegram Bot API connection closed")
}

type MockBotAPI struct {
	updates chan tgbotapi.Update
	sent    []MockMessage
}

type MockMessage struct {
	ChatID int64
	Text   string
}

func NewMockBotAPI() *MockBotAPI {
	return &MockBotAPI{
		updates: make(chan tgbotapi.Update, 10),
		sent:    make([]MockMessage, 0),
	}
}

func (m *MockBotAPI) GetLastEvents() <-chan tgbotapi.Update {
	return m.updates
}

func (m *MockBotAPI) SendMessage(chatID int64, text string) error {
	m.sent = append(m.sent, MockMessage{
		ChatID: chatID,
		Text:   text,
	})
	return nil
}

func (m *MockBotAPI) Close() {
	close(m.updates)
}

func (m *MockBotAPI) AddUpdate(update tgbotapi.Update) {
	m.updates <- update
}

func (m *MockBotAPI) GetSentMessages() []MockMessage {
	return m.sent
}
