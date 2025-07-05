package app

import (
	"log"

	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	// UpdateOffset Update offset constants
	UpdateOffsetFromBeginning = 0  // Get all pending updates from start
	UpdateOffsetOnlyNew       = -1 // Skip pending, get only new updates
	UpdateOffsetResume        = 1  // Base for resuming from specific point
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

	// FYI: Start from `UpdateOffsetFromBeginning`
	u := tgBotAPI.NewUpdate(UpdateOffsetFromBeginning)

	// FYI: Wait up to `UpdateTimeoutSeconds` seconds for new messages
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

func NewMockBotAPI() *MockBotAPI {
	return &MockBotAPI{
		updates: make(chan tgBotAPI.Update, 10),
		sent:    make([]MockMessage, 0),
	}
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

func (m *MockBotAPI) Close() {
	close(m.updates)
}

func (m *MockBotAPI) AddUpdate(update tgBotAPI.Update) {
	m.updates <- update
}

func (m *MockBotAPI) AddTestMessage(chatID, userID int64, username, text string) {
	update := tgBotAPI.Update{
		Message: &tgBotAPI.Message{
			Chat: &tgBotAPI.Chat{ID: chatID},
			From: &tgBotAPI.User{ID: userID, UserName: username},
			Text: text,
		},
	}
	m.updates <- update
}

func (m *MockBotAPI) GetSentMessages() []MockMessage {
	return m.sent
}
