package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"strconv"
	"strings"
	"time"
)

type DepositCreationHandler struct{}

func (h *DepositCreationHandler) GetSessionType() SessionType {
	return SessionCreateDeposit
}

func (h *DepositCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.processStep(ctx, bot, session, msg)
	}

	params := map[string]interface{}{
		"user_id":      int64(session.UserID),
		"session_step": string(session.CurrentStep),
		"session_type": "create_deposit",
	}

	wrappedHandler := bot.interceptor.Chain(handler, "DepositCreationHandler.HandleStep")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *DepositCreationHandler) processStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	switch session.CurrentStep {
	case StepStart:
		return h.handleStart(bot, session)
	case StepName:
		return h.handleName(bot, session, msg.Text)
	case StepAmount:
		return h.handleAmount(bot, session, msg.Text)
	case StepCurrency:
		return h.handleCurrency(bot, session, msg.Text)
	case StepInterestRate:
		return h.handleInterestRate(bot, session, msg.Text)
	case StepDate:
		return h.handleExpirationDate(bot, session, msg.Text)
	case StepConfirmation:
		return h.handleConfirmation(ctx, bot, session, msg.Text)
	default:
		return errors.ErrUnknownStep.WithContext("step", session.CurrentStep)
	}
}

func (h *DepositCreationHandler) handleStart(bot *Bot, session *UserSession) error {
	text := `🏦 Создание нового депозита

		Шаг 1/5: Введите название депозита
		Например: "Сбербанк Депозит", "Накопления на отпуск"
		
		Для отмены введите /cancel`

	bot.sendMessage(session.ChatID, text)
	session.CurrentStep = StepName
	return nil
}

func (h *DepositCreationHandler) handleName(bot *Bot, session *UserSession, input string) error {
	name := strings.TrimSpace(input)

	if err := h.validateName(name); err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ %s Попробуйте еще раз:", err.Error()))
		return nil
	}

	session.SetData("name", name)
	session.CurrentStep = StepAmount

	text := fmt.Sprintf(`✅ Название: %s
		
		Шаг 2/5: Введите сумму депозита
		Например: %s, %s
		
		Минимум: %s 
		Максимум: %s
		`, name, FormatInteger(100000), FormatNumber(50000.50), FormatInteger(1), FormatInteger(1000000000))

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *DepositCreationHandler) handleAmount(bot *Bot, session *UserSession, input string) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ Некорректная сумма. Введите число (например: %s или %s):", FormatInteger(100000), FormatNumber(50000.50)))
		return nil
	}

	if err = h.validateAmount(amount); err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ %s Попробуйте еще раз:", err.Error()))
		return nil
	}

	session.SetData("amount", amount)
	session.CurrentStep = StepCurrency

	text := fmt.Sprintf(`✅ Сумма: %s

Шаг 3/5: Выберите валюту

💰 Выберите валюту из списка ниже или введите код валюты:`, FormatNumber(amount))

	keyboard := CreateCurrencySelectionKeyboard()
	bot.sendMessageWithKeyboard(session.ChatID, text, keyboard)
	return nil
}

func (h *DepositCreationHandler) handleCurrency(bot *Bot, session *UserSession, input string) error {
	currencyStr := strings.TrimSpace(input)

	if strings.HasPrefix(input, CallbackCurrencyPrefix) {
		currencyCode := strings.TrimPrefix(input, CallbackCurrencyPrefix)

		if currencyCode == "skip" {
			session.SetData("currency", domain.RUB)
			session.CurrentStep = StepInterestRate
			h.sendInterestRatePrompt(bot, session)
			return nil
		}

		currency := domain.CurrencyName(currencyCode)
		if !currency.IsValid() {
			bot.sendMessage(session.ChatID, "❌ Неизвестная валюта. Используйте кнопки выше для выбора:")
			return nil
		}

		session.SetData("currency", currency)
		session.CurrentStep = StepInterestRate
		h.sendInterestRatePrompt(bot, session)
		return nil
	}

	if IsSkipResponse(currencyStr) {
		session.SetData("currency", domain.RUB)
		session.CurrentStep = StepInterestRate
		h.sendInterestRatePrompt(bot, session)
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

	if !h.isCurrencyAllowed(currency) {
		text := fmt.Sprintf(`❌ Валюта "%s" (%s) не поддерживается

Используйте кнопки выше для выбора поддерживаемой валюты:`, currency, currency.ToHumanRussian())
		bot.sendMessage(session.ChatID, text)
		return nil
	}

	session.SetData("currency", currency)
	session.CurrentStep = StepInterestRate

	h.sendInterestRatePrompt(bot, session)
	return nil
}

func (h *DepositCreationHandler) sendInterestRatePrompt(bot *Bot, session *UserSession) {
	currency := session.GetData("currency").(domain.CurrencyName)

	text := fmt.Sprintf(`✅ Валюта: %s (%s)

			Шаг 4/5: Введите процентную ставку (необязательно)
			Например: 5.5; 3.25; 7.0
			
			Диапазон: 0-50%%
			Введите "пропустить" если не хотите указывать`,
		currency,
		currency.ToHumanRussian())

	bot.sendMessage(session.ChatID, text)
}

func (h *DepositCreationHandler) handleInterestRate(bot *Bot, session *UserSession, input string) error {
	rateStr := strings.TrimSpace(input)

	if IsSkipResponse(rateStr) {
		session.CurrentStep = StepDate
		h.sendExpirationDatePrompt(bot, session)
		return nil
	}

	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		bot.sendMessage(session.ChatID, "❌ Некорректная процентная ставка. Введите число (например: 5.5):")
		return nil
	}

	if err = h.validateInterestRate(rate); err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ %s Попробуйте еще раз:", err.Error()))
		return nil
	}

	session.SetData("interest_rate", rate)
	session.CurrentStep = StepDate

	h.sendExpirationDatePrompt(bot, session)
	return nil
}

func (h *DepositCreationHandler) sendExpirationDatePrompt(bot *Bot, session *UserSession) {
	rateText := "не указана"
	if rate := session.GetFloat("interest_rate"); rate > 0 {
		rateText = fmt.Sprintf("%s%%", FormatNumber(rate))
	}

	text := fmt.Sprintf(`✅ Процентная ставка: %s

		Шаг 5/5: Введите дату окончания депозита (необязательно)
		Форматы: 31.12.2025, 2025-12-31, 31/12/2025
		
		Введите "пропустить" если не хотите указывать дату`, rateText)

	bot.sendMessage(session.ChatID, text)
}

func (h *DepositCreationHandler) handleExpirationDate(bot *Bot, session *UserSession, input string) error {
	handler := ExpirationDateHandler{
		DataKey:        "expiration_date",
		NextStep:       StepConfirmation,
		ConfirmationFn: h.sendConfirmationPrompt,
		StoreAsPointer: true,
	}
	return HandleExpirationDate(bot, session, input, handler)
}

func (h *DepositCreationHandler) sendConfirmationPrompt(bot *Bot, session *UserSession) {
	text := h.FormatConfirmation(session)
	text += "\n\n🔍 Подтвердите создание депозита:"

	keyboard := CreateConfirmationKeyboard(CallbackConfirmDepositYes, CallbackConfirmDepositNo)
	bot.sendMessageWithKeyboard(session.ChatID, text, keyboard)
}

func (h *DepositCreationHandler) handleConfirmation(ctx context.Context, bot *Bot, session *UserSession, input string) error {
	if input == CallbackConfirmDepositYes {
		bot.sessionManager.ClearSession(session.UserID)
		return h.CompleteSession(ctx, bot, session)
	}

	if input == CallbackConfirmDepositNo {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание депозита отменено.")
		return nil
	}

	if IsNegativeResponse(input) {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание депозита отменено.")
		return nil
	}

	if !IsPositiveResponse(input) {
		if IsValidResponse(input) {
			bot.sendMessage(session.ChatID, "❓ Используйте кнопки выше или введите 'да' для создания или 'нет' для отмены:")
		} else {
			bot.sendMessage(session.ChatID, GetSuggestionMessage())
		}
		return nil
	}

	bot.sessionManager.ClearSession(session.UserID)
	return h.CompleteSession(ctx, bot, session)
}

func (h *DepositCreationHandler) GetNextStep(currentStep SessionStep, input string) (SessionStep, error) {
	return currentStep, nil
}

func (h *DepositCreationHandler) ValidateInput(step SessionStep, input string) error {
	return nil
}

func (h *DepositCreationHandler) FormatConfirmation(session *UserSession) string {
	name := session.GetString("name")
	amount := session.GetFloat("amount")
	currency := session.GetData("currency").(domain.CurrencyName)

	text := "📋 Подтверждение создания депозита:\n\n"
	text += fmt.Sprintf("📝 Название: %s\n", name)
	text += fmt.Sprintf("💰 Сумма: %s %s\n", FormatNumber(amount), currency.Symbol())
	text += fmt.Sprintf("💱 Валюта: %s (%s)\n", currency, currency.ToHumanRussian())

	if rate := session.GetFloat("interest_rate"); rate > 0 {
		text += fmt.Sprintf("📈 Процентная ставка: %s%%\n", FormatNumber(rate))
	} else {
		text += "📈 Процентная ставка: не указана\n"
	}

	if date := session.GetTime("expiration_date"); date != nil {
		text += fmt.Sprintf("📅 Дата окончания: %s\n", date.Format("02.01.2006"))
	} else {
		text += "📅 Дата окончания: не указана\n"
	}

	return text
}

func (h *DepositCreationHandler) CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.executeCompletion(ctx, bot, session)
	}

	params := map[string]interface{}{
		"user_id":        int64(session.UserID),
		"deposit_name":   session.GetString("name"),
		"deposit_amount": session.GetFloat("amount"),
	}

	wrappedHandler := bot.interceptor.Chain(handler, "DepositCreationHandler.CompleteSession")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *DepositCreationHandler) executeCompletion(ctx context.Context, bot *Bot, session *UserSession) error {

	req := &models.CreateDepositRequest{
		UserID:    session.UserID,
		Name:      session.GetString("name"),
		AmountRUB: session.GetFloat("amount"),
		Currency:  string(session.GetData("currency").(domain.CurrencyName)),
	}

	if rate := session.GetFloat("interest_rate"); rate > 0 {
		req.InterestRatePercent = &rate
	}

	if date := session.GetTime("expiration_date"); date != nil {
		req.ExpirationDate = date
	}

	deposit, err := bot.financeService.CreateDeposit(ctx, req)
	if err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ Ошибка при создании депозита: %s", err.Error()))
		bot.sessionManager.ClearSession(session.UserID)
		return err
	}

	h.sendDepositCreatedConfirmation(bot, session.ChatID, deposit)

	return nil
}

func (h *DepositCreationHandler) validateName(name string) error {
	if name == "" {
		return errors.ErrNameEmpty
	}
	if len(name) > 255 {
		return errors.ErrNameTooLong
	}
	return nil
}

func (h *DepositCreationHandler) validateAmount(amount float64) error {
	if amount <= 0 {
		return errors.ErrAmountNotPositive
	}
	if amount > 1000000000 {
		return errors.ErrAmountTooLarge
	}
	return nil
}

func (h *DepositCreationHandler) validateInterestRate(rate float64) error {
	if rate < 0 || rate > 50 {
		return errors.ErrRateOutOfRange
	}
	return nil
}

func (h *DepositCreationHandler) isCurrencyAllowed(currency domain.CurrencyName) bool {
	return currency.IsValid()
}

func (h *DepositCreationHandler) sendDepositCreatedConfirmation(bot *Bot, chatID int64, deposit *domain.Deposit) {
	amount := float64(deposit.AmountMinorUnits) / 100.0

	text := fmt.Sprintf("✅ Депозит успешно создан!\n\n")
	text += fmt.Sprintf("📋 Название: %s\n", deposit.Name)
	text += fmt.Sprintf("💰 Сумма: %s %s\n", FormatNumber(amount), deposit.Currency)

	if deposit.InterestRateBasisPoints > 0 {
		rate := float64(deposit.InterestRateBasisPoints) / 100.0
		text += fmt.Sprintf("📈 Процентная ставка: %s%%\n", FormatNumber(rate))
	} else {
		text += "📈 Процентная ставка: не указана\n"
	}

	if deposit.ExpirationDate != nil {
		text += fmt.Sprintf("📅 Дата окончания: %s\n", deposit.ExpirationDate.Format("02.01.2006"))
	} else {
		text += "📅 Дата окончания: не указана\n"
	}

	text += fmt.Sprintf("🆔 ID: %s\n", FormatInteger(int64(deposit.Id)))
	text += fmt.Sprintf("📅 Создан: %s", time.Now().Format("02.01.2006 15:04"))

	bot.sendMessage(chatID, text)
	bot.sendMainMenu(chatID)
}
