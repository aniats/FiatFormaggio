package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"strconv"
	"strings"
)

type CashHoldingCreationHandler struct{}

func (h *CashHoldingCreationHandler) GetSessionType() SessionType {
	return SessionCreateCashHolding
}

func (h *CashHoldingCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.processStep(ctx, bot, session, msg)
	}

	params := map[string]interface{}{
		"user_id":      int64(session.UserID),
		"session_step": string(session.CurrentStep),
		"session_type": "create_cash_holding",
	}

	wrappedHandler := bot.interceptor.Chain(handler, "CashHoldingCreationHandler.HandleStep")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *CashHoldingCreationHandler) processStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	switch session.CurrentStep {
	case StepStart:
		return h.handleStart(bot, session)
	case StepName:
		return h.handleName(bot, session, msg.Text)
	case StepAmount:
		return h.handleAmount(bot, session, msg.Text)
	case StepCurrency:
		return h.handleCurrency(bot, session, msg.Text)
	case StepConfirmation:
		return h.handleConfirmation(ctx, bot, session, msg.Text)
	default:
		return errors.ErrUnknownStep.WithContext("step", session.CurrentStep)
	}
}

func (h *CashHoldingCreationHandler) handleStart(bot *Bot, session *UserSession) error {
	text := `💵 Создание нового наличного счета

	Шаг 1/4: Введите название счета
	Например: "Наличные в кошельке", "Доллары дома", "Евро в сейфе"
	
	Для отмены введите /cancel`

	bot.sendMessage(session.ChatID, text)
	session.CurrentStep = StepName
	return nil
}

func (h *CashHoldingCreationHandler) handleName(bot *Bot, session *UserSession, input string) error {
	name := strings.TrimSpace(input)

	if err := h.validateName(name); err != nil {
		bot.sendMessage(session.ChatID, errors.GetUserMessage(err))
		return nil
	}

	session.SetData("name", name)
	session.CurrentStep = StepAmount

	text := fmt.Sprintf(`✅ Название: "%s"

	Шаг 2/4: Введите сумму наличных
	Например: 5000, 1500.50, 100
	
	Валюта будет указана на следующем шаге.`, name)

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *CashHoldingCreationHandler) handleAmount(bot *Bot, session *UserSession, input string) error {
	amountStr := strings.TrimSpace(strings.Replace(input, ",", ".", -1))

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		bot.sendMessage(session.ChatID, errors.GetUserMessage(errors.ErrInvalidInput))
		return nil
	}

	if err := h.validateAmount(amount); err != nil {
		bot.sendMessage(session.ChatID, errors.GetUserMessage(err))
		return nil
	}

	session.SetData("amount", amount)
	session.CurrentStep = StepCurrency

	text := fmt.Sprintf(`✅ Сумма: %s

	Шаг 3/4: Выберите валюту
	Доступные варианты:
	• RUB, рубль - Российский рубль ₽
	• USD, доллар - Американский доллар $
	• EUR, евро - Евро €
	• CNY, юань - Китайский юань ¥
	• GBP, фунт - Британский фунт £
	
	По умолчанию: RUB (введите "пропустить" для RUB)`, FormatNumber(amount))

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *CashHoldingCreationHandler) handleCurrency(bot *Bot, session *UserSession, input string) error {
	currencyStr := strings.TrimSpace(input)

	if IsSkipResponse(currencyStr) || currencyStr == "" {
		session.SetData("currency", domain.RUB)
		session.CurrentStep = StepConfirmation
		h.sendConfirmation(bot, session)
		return nil
	}

	currency, err := domain.CurrencyFromHuman(currencyStr)
	if err != nil {
		currencyErr := errors.NewCurrencyError(currencyStr)
		bot.sendMessage(session.ChatID, errors.GetUserMessage(currencyErr))
		return nil
	}

	session.SetData("currency", currency)
	session.CurrentStep = StepConfirmation

	h.sendConfirmation(bot, session)
	return nil
}

func (h *CashHoldingCreationHandler) sendConfirmation(bot *Bot, session *UserSession) {
	name := session.GetData("name").(string)
	amount := session.GetData("amount").(float64)
	currency := session.GetData("currency").(domain.CurrencyName)

	text := fmt.Sprintf(`💵 Подтверждение создания наличного счета

	📝 Название: %s
	💰 Сумма: %s
	💱 Валюта: %s (%s)

	Все верно? Отправьте "да" для создания счета или "нет" для отмены.`,
		name,
		currency.FormatAmountRussian(amount),
		currency.ToHumanRussian(),
		currency.Symbol())

	bot.sendMessage(session.ChatID, text)
}

func (h *CashHoldingCreationHandler) handleConfirmation(ctx context.Context, bot *Bot, session *UserSession, input string) error {
	if IsNegativeResponse(input) {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание наличного счета отменено.")
		return nil
	}

	if !IsPositiveResponse(input) {
		if IsValidResponse(input) {
			bot.sendMessage(session.ChatID, "Пожалуйста, ответьте 'да' для подтверждения или 'нет' для отмены:")
		} else {
			bot.sendMessage(session.ChatID, GetSuggestionMessage())
		}
		return nil
	}

	bot.sessionManager.ClearSession(session.UserID)
	return h.CompleteSession(ctx, bot, session)
}

func (h *CashHoldingCreationHandler) CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.executeCompletion(ctx, bot, session)
	}

	name := session.GetData("name").(string)
	params := map[string]interface{}{
		"user_id":           int64(session.UserID),
		"cash_holding_name": name,
	}

	wrappedHandler := bot.interceptor.Chain(handler, "CashHoldingCreationHandler.CompleteSession")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *CashHoldingCreationHandler) executeCompletion(ctx context.Context, bot *Bot, session *UserSession) error {
	name := session.GetData("name").(string)
	amount := session.GetData("amount").(float64)
	currency := session.GetData("currency").(domain.CurrencyName)

	req := &models.CreateCashHoldingRequest{
		UserID:    session.UserID,
		Name:      name,
		AmountRUB: amount,
		Currency:  string(currency),
	}

	cashHolding, err := bot.financeService.CreateCashHolding(ctx, req)
	if err != nil {
		serviceErr := errors.WrapError(err, errors.CodeInternalError, 
			"Failed to create cash holding", 
			"❌ Ошибка при создании наличного счета. Попробуйте позже.")
		bot.sendMessage(session.ChatID, errors.GetUserMessage(serviceErr))
		return serviceErr
	}

	text := fmt.Sprintf(`✅ Наличный счет успешно создан!

📝 Название: %s
💰 Сумма: %s
🆔 ID: %s

Используйте /cash_holdings чтобы посмотреть все ваши наличные счета.`,
		cashHolding.Name,
		currency.FormatAmountRussian(amount),
		FormatInteger(int64(cashHolding.Id)))

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *CashHoldingCreationHandler) GetNextStep(currentStep SessionStep, input string) (SessionStep, error) {
	return StepComplete, nil
}

func (h *CashHoldingCreationHandler) ValidateInput(step SessionStep, input string) error {
	return nil
}

func (h *CashHoldingCreationHandler) FormatConfirmation(session *UserSession) string {
	return "Confirmation"
}

func (h *CashHoldingCreationHandler) validateName(name string) error {
	if name == "" {
		return errors.NewBusinessError(errors.CodeInvalidName, "❌ Название не может быть пустым. Попробуйте еще раз:")
	}
	if len(name) > 255 {
		return errors.NewBusinessError(errors.CodeInvalidName, "❌ Название слишком длинное (максимум 255 символов). Попробуйте еще раз:")
	}
	return nil
}

func (h *CashHoldingCreationHandler) validateAmount(amount float64) error {
	if amount < 0 {
		return errors.NewBusinessError(errors.CodeInvalidAmount, "❌ Сумма не может быть отрицательной. Попробуйте еще раз:")
	}
	if amount > 1000000000 {
		return errors.NewBusinessError(errors.CodeInvalidAmount, fmt.Sprintf("❌ Слишком большая сумма (максимум %s). Попробуйте еще раз:", FormatInteger(1000000000)))
	}
	return nil
}
