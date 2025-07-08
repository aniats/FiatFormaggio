package app

import (
	"context"
	"fmt"
	"log"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/chatgpt"
	"github.com/aniats/FiatFormaggio/internal/tracing"
	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type FinancialAdviceHandler struct{}

func (h *FinancialAdviceHandler) GetSessionType() SessionType { return SessionFinancialAdvice }

func (h *FinancialAdviceHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *domain.Message) error {
	tracer := otel.Tracer("financial-advice-handler")
	ctx, span := tracer.Start(ctx, "FinancialAdviceHandler.HandleStep")
	defer span.End()

	userMessage := msg.Text

	span.SetAttributes(
		attribute.Int64(tracing.UserID, int64(session.UserID)),
		attribute.Int64(tracing.ChatID, session.ChatID),
		attribute.String(tracing.SessionType, string(session.Type)),
		attribute.String(tracing.UserMessage, userMessage),
	)

	if userMessage == domain.CallbackStopFinancialAdvice {
		span.SetAttributes(attribute.Bool(tracing.SessionStopped, true))
		span.SetStatus(codes.Ok, "Financial advice session stopped by user")
		bot.sendMessage(session.ChatID, "Сеанс финансовых советов завершен.")
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMainMenu(session.ChatID)
		return nil
	}

	previousMessages := h.getPreviousMessages(session)
	span.SetAttributes(
		attribute.Int(tracing.PreviousMessagesCount, len(previousMessages)),
	)

	response, err := bot.financeService.GetFinancialAdvice(ctx, session.UserID, userMessage, previousMessages)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get AI financial advice")
		log.Printf("Failed to get financial advice: %v", err)
		bot.sendMessage(session.ChatID, "❌ Ошибка получения совета. Попробуйте позже.")
		return err
	}

	h.storeMessages(session, userMessage, response)

	responseText := fmt.Sprintf("💡 %s\n\n💬 Продолжайте задавать вопросы:", response)

	stopKeyboard := tgBotAPI.NewInlineKeyboardMarkup(
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("🛑 Завершить сеанс", domain.CallbackStopFinancialAdvice),
		),
	)

	bot.sendMessageWithKeyboard(session.ChatID, responseText, stopKeyboard)

	span.SetAttributes(
		attribute.Int(tracing.ResponseLength, len(response)),
		attribute.Bool(tracing.SessionActive, true),
	)
	span.SetStatus(codes.Ok, "Financial advice provided successfully")

	return nil
}

func (h *FinancialAdviceHandler) getPreviousMessages(session *UserSession) []chatgpt.ChatMessage {
	if messages, ok := session.Data["messages"].([]chatgpt.ChatMessage); ok {
		return messages
	}
	return []chatgpt.ChatMessage{}
}

func (h *FinancialAdviceHandler) storeMessages(session *UserSession, userMessage, aiResponse string) {
	messages := h.getPreviousMessages(session)

	messages = append(messages, chatgpt.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	messages = append(messages, chatgpt.ChatMessage{
		Role:    "assistant",
		Content: aiResponse,
	})

	session.Data["messages"] = messages
}

func (h *FinancialAdviceHandler) FormatConfirmation(*UserSession) string {
	return ""
}

func (h *FinancialAdviceHandler) CompleteSession(context.Context, *Bot, *UserSession) error {
	return nil
}

func (b *Bot) sendFinancialAdviceCommand(ctx context.Context, message *domain.Message) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, b.processFinancialAdviceCommand(ctx, message)
	}

	params := map[string]interface{}{
		tracing.ParamChatID: message.ChatID,
		tracing.ParamUserID: message.UserID,
		tracing.Command:     tracing.FinancialAdviceCommand,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendFinancialAdviceCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processFinancialAdviceCommand(ctx context.Context, message *domain.Message) error {
	tracer := otel.Tracer("financial-advice-bot")
	ctx, span := tracer.Start(ctx, "Bot.processFinancialAdviceCommand")
	defer span.End()

	chatID := message.ChatID
	UserID := domain.UserID(message.UserID)

	span.SetAttributes(
		attribute.Int64(tracing.UserID, int64(UserID)),
		attribute.Int64(tracing.ChatID, chatID),
		attribute.String(tracing.Command, tracing.FinancialAdviceCommand),
	)

	initialPrompt := "Проанализируйте мой портфель и дайте общие рекомендации по инвестированию и диверсификации."
	span.SetAttributes(attribute.String(tracing.InitialPrompt, initialPrompt))

	response, err := b.financeService.GetFinancialAdvice(ctx, UserID, initialPrompt, []chatgpt.ChatMessage{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get initial AI advice")
		log.Printf("Failed to get initial financial advice: %v", err)
		b.sendMessage(chatID, "❌ Ошибка получения финансовых советов. Попробуйте позже.")
		return err
	}

	session, err := b.sessionManager.StartSession(UserID, chatID, SessionFinancialAdvice)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to start financial advice session")
		log.Printf("Failed to start financial advice session: %v", err)
		b.sendMessage(chatID, "❌ Ошибка запуска сеанса. Попробуйте позже.")
		return err
	}
	log.Printf("Started financial advice session for user %d, session type: %s", UserID, session.Type)

	welcomeText := fmt.Sprintf("💡 **Финансовые советы для вашего портфеля:**\n\n%s\n\n", response)
	welcomeText += "💬 **Теперь вы можете задавать вопросы по финансам и инвестициям.**\n"
	welcomeText += "Например: \"Как лучше диверсифицировать портфель?\", \"Стоит ли инвестировать в акции?\""

	stopKeyboard := tgBotAPI.NewInlineKeyboardMarkup(
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("🛑 Завершить сеанс", domain.CallbackStopFinancialAdvice),
		),
	)

	b.sendMessageWithKeyboard(chatID, welcomeText, stopKeyboard)

	span.SetAttributes(
		attribute.String(tracing.SessionID, string(session.Type)),
		attribute.Int(tracing.InitialResponseLength, len(response)),
		attribute.Bool(tracing.SessionCreated, true),
	)
	span.SetStatus(codes.Ok, "Financial advice session started successfully")

	return nil
}
