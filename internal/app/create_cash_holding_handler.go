package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/utils"
	"strconv"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"github.com/aniats/FiatFormaggio/internal/tracing"
)

type CashHoldingCreationHandler struct{}

func (h *CashHoldingCreationHandler) GetSessionType() SessionType {
	return SessionCreateCashHolding
}

func (h *CashHoldingCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *domain.Message) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.processStep(ctx, bot, session, msg)
	}

	params := map[string]interface{}{
		tracing.ParamUserID:      int64(session.UserID),
		tracing.ParamSessionStep: string(session.CurrentStep),
		tracing.ParamSessionType: "create_cash_holding",
	}

	wrappedHandler := bot.interceptor.Chain(handler, "CashHoldingCreationHandler.HandleStep")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *CashHoldingCreationHandler) processStep(ctx context.Context, bot *Bot, session *UserSession, msg *domain.Message) error {
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

	if err := h.ValidateName(name); err != nil {
		bot.sendMessage(session.ChatID, errors.GetUserMessage(err))
		return nil
	}

	session.SetData("name", name)
	session.CurrentStep = StepAmount

	text := fmt.Sprintf(`✅ Название: "%s"

	Шаг 2/4: Введите сумму наличных
	Например: 5,000; 1,500.50; 100
	
	Валюта будет указана на следующем шаге.`, name)

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *CashHoldingCreationHandler) handleAmount(bot *Bot, session *UserSession, input string) error {
	amountStr := strings.TrimSpace(strings.Replace(input, ",", ".", -1))

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		bot.sendMessage(session.ChatID, errors.GetUserMessage(errors.ErrInValidInput))
		return nil
	}

	if err = h.ValidateAmount(amount); err != nil {
		bot.sendMessage(session.ChatID, errors.GetUserMessage(err))
		return nil
	}

	session.SetData("amount", amount)
	session.CurrentStep = StepCurrency

	text := fmt.Sprintf(`
	✅ Сумма: %s

	Шаг 3/4: Выберите валюту
	
	💰 Выберите валюту из списка ниже или введите код валюты:`, utils.FormatNumber(amount))

	keyboard := CreateCurrencySelectionKeyboard()
	bot.sendMessageWithKeyboard(session.ChatID, text, keyboard)
	return nil
}

func (h *CashHoldingCreationHandler) handleCurrency(bot *Bot, session *UserSession, input string) error {
	currencyStr := strings.TrimSpace(input)

	if strings.HasPrefix(input, domain.CallbackCurrencyPrefix) {
		currencyCode := strings.TrimPrefix(input, domain.CallbackCurrencyPrefix)

		if currencyCode == "skip" {
			session.SetData("currency", domain.RUB)
			session.CurrentStep = StepConfirmation
			h.sendConfirmation(bot, session)
			return nil
		}

		currency := domain.CurrencyName(currencyCode)
		if !currency.IsValid() {
			bot.sendMessage(session.ChatID, "❌ Неизвестная валюта. Используйте кнопки выше для выбора:")
			return nil
		}

		session.SetData("currency", currency)
		session.CurrentStep = StepConfirmation
		h.sendConfirmation(bot, session)
		return nil
	}

	if utils.IsSkipResponse(currencyStr) {
		session.SetData("currency", domain.RUB)
		session.CurrentStep = StepConfirmation
		h.sendConfirmation(bot, session)
		return nil
	}

	currency, err := domain.CurrencyFromHuman(currencyStr)
	if err != nil {
		text := fmt.Sprintf(`❌ Неизвестная валюта "%s"

	Доступные варианты:
	• RUB - Российский рубль ₽
	• USD - Доллар США $
	• EUR - Евро €
	• GBP - Британский фунт £
	• JPY - Японская иена ¥
	• CNY - Китайский юань ¥
	• RSD - Сербский динар
	• XBT - Биткоин ₿
	• KZT - Казахстанский тенге
	
	Используйте кнопки выше или введите код валюты:`, currencyStr)
		bot.sendMessage(session.ChatID, text)
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

🔍 Подтвердите создание наличного счета:`,
		name,
		currency.FormatAmountRussian(amount),
		currency.ToHumanRussian(),
		currency.Symbol())

	keyboard := CreateConfirmationKeyboard(domain.CallbackConfirmCashYes, domain.CallbackConfirmCashNo)
	bot.sendMessageWithKeyboard(session.ChatID, text, keyboard)
}

func (h *CashHoldingCreationHandler) handleConfirmation(ctx context.Context, bot *Bot, session *UserSession, input string) error {
	if input == domain.CallbackConfirmCashYes {
		bot.sessionManager.ClearSession(session.UserID)
		return h.CompleteSession(ctx, bot, session)
	}

	if input == domain.CallbackConfirmCashNo {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание наличного счета отменено.")
		return nil
	}

	if utils.IsNegativeResponse(input) {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание наличного счета отменено.")
		return nil
	}

	if !utils.IsPositiveResponse(input) {
		if utils.IsValidResponse(input) {
			bot.sendMessage(session.ChatID, "❓ Используйте кнопки выше или введите 'да' для создания или 'нет' для отмены:")
		} else {
			bot.sendMessage(session.ChatID, utils.GetSuggestionMessage())
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
		tracing.ParamUserID: int64(session.UserID),
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
		utils.FormatInteger(cashHolding.ID))

	bot.sendMessage(session.ChatID, text)
	bot.sendMainMenu(session.ChatID)
	return nil
}

func (h *CashHoldingCreationHandler) FormatConfirmation(session *UserSession) string {
	return "Confirmation"
}

func (h *CashHoldingCreationHandler) ValidateName(name string) error {
	if name == "" {
		return errors.NewBusinessError(errors.CodeInValidName, "❌ Название не может быть пустым. Попробуйте еще раз:")
	}
	if len(name) > 255 {
		return errors.NewBusinessError(errors.CodeInValidName, "❌ Название слишком длинное (максимум 255 символов). Попробуйте еще раз:")
	}
	return nil
}

func (h *CashHoldingCreationHandler) ValidateAmount(amount float64) error {
	if amount < 0 {
		return errors.NewBusinessError(errors.CodeInValidAmount, "❌ Сумма не может быть отрицательной. Попробуйте еще раз:")
	}
	if amount > 1000000000 {
		return errors.NewBusinessError(errors.CodeInValidAmount, fmt.Sprintf("❌ Слишком большая сумма (максимум %s). Попробуйте еще раз:", utils.FormatInteger(1000000000)))
	}
	return nil
}
