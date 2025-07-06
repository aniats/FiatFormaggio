package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

type SavingAccountCreationHandler struct{}

func (h *SavingAccountCreationHandler) GetSessionType() SessionType {
	return SessionCreateSavingAccount
}

func (h *SavingAccountCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.processStep(ctx, bot, session, msg)
	}

	params := map[string]interface{}{
		"user_id":      int64(session.UserID),
		"session_step": string(session.CurrentStep),
		"session_type": "create_saving_account",
	}

	wrappedHandler := bot.interceptor.Chain(handler, "SavingAccountCreationHandler.HandleStep")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *SavingAccountCreationHandler) processStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
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

func (h *SavingAccountCreationHandler) handleStart(bot *Bot, session *UserSession) error {
	text := `💰 Создание нового накопительного счета

	Шаг 1/6: Введите название счета
	Например: "Сбербанк Накопительный", "Копилка на отпуск"
	
	Для отмены введите /cancel`

	bot.sendMessage(session.ChatID, text)
	session.CurrentStep = StepName
	return nil
}

func (h *SavingAccountCreationHandler) handleName(bot *Bot, session *UserSession, input string) error {
	name := strings.TrimSpace(input)

	if name == "" {
		bot.sendMessage(session.ChatID, "❌ Название не может быть пустым. Попробуйте еще раз:")
		return nil
	}

	if len(name) > 255 {
		bot.sendMessage(session.ChatID, "❌ Название слишком длинное (максимум 255 символов). Попробуйте еще раз:")
		return nil
	}

	session.SetData("name", name)
	session.CurrentStep = StepAmount

	text := fmt.Sprintf(`✅ Название: "%s"

	Шаг 2/6: Введите текущую сумму на счете
	Например: 25,000; 1,000.50; 0
	
	Валюта будет указана на следующем шаге.`, name)

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *SavingAccountCreationHandler) handleAmount(bot *Bot, session *UserSession, input string) error {
	amountStr := strings.TrimSpace(strings.Replace(input, ",", ".", -1))

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		bot.sendMessage(session.ChatID, "❌ Неверный формат суммы. Введите число (например: 25000, 1000.50, 0):")
		return nil
	}

	if amount < 0 {
		bot.sendMessage(session.ChatID, "❌ Сумма не может быть отрицательной. Попробуйте еще раз:")
		return nil
	}

	if amount > 1000000000 {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ Слишком большая сумма (максимум %s). Попробуйте еще раз:", FormatInteger(1000000000)))
		return nil
	}

	session.SetData("amount", amount)
	session.CurrentStep = StepCurrency

	text := fmt.Sprintf(`✅ Сумма: %s

Шаг 3/6: Выберите валюту счета

💰 Выберите валюту из списка ниже или введите код валюты:`, FormatNumber(amount))

	keyboard := CreateCurrencySelectionKeyboard()
	bot.sendMessageWithKeyboard(session.ChatID, text, keyboard)
	return nil
}

func (h *SavingAccountCreationHandler) handleCurrency(bot *Bot, session *UserSession, input string) error {
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

	session.SetData("currency", currency)
	session.CurrentStep = StepInterestRate

	h.sendInterestRatePrompt(bot, session)
	return nil
}

func (h *SavingAccountCreationHandler) sendInterestRatePrompt(bot *Bot, session *UserSession) {
	currency := session.GetData("currency").(domain.CurrencyName)
	amount := session.GetData("amount").(float64)

	text := fmt.Sprintf(`
		✅ Валюта: %s (%s)
		✅ Сумма: %s
	
		Шаг 4/6: Введите процентную ставку (необязательно)
		Например: 5.5, 7.2, 4
		
		Введите "пропустить" если не хотите указывать ставку
		`,
		currency.ToHumanRussian(),
		currency.Symbol(),
		currency.FormatAmountRussian(amount))

	bot.sendMessage(session.ChatID, text)
}

func (h *SavingAccountCreationHandler) handleInterestRate(bot *Bot, session *UserSession, input string) error {
	rateStr := strings.TrimSpace(input)

	if IsSkipResponse(rateStr) || rateStr == "" {
		session.SetData("interestRate", nil)
		session.CurrentStep = StepDate
		h.sendExpirationDatePrompt(bot, session)
		return nil
	}

	rate, err := strconv.ParseFloat(strings.Replace(rateStr, ",", ".", -1), 64)
	if err != nil {
		bot.sendMessage(session.ChatID, "❌ Неверный формат ставки. Введите число (например: 5.5, 7.2, 4):")
		return nil
	}

	if rate < 0 || rate > 50 {
		bot.sendMessage(session.ChatID, "❌ Ставка должна быть от 0 до 50%. Попробуйте еще раз:")
		return nil
	}

	session.SetData("interestRate", &rate)
	session.CurrentStep = StepDate

	h.sendExpirationDatePrompt(bot, session)
	return nil
}

func (h *SavingAccountCreationHandler) sendExpirationDatePrompt(bot *Bot, session *UserSession) {
	rateData := session.GetData("interestRate")
	rateText := "Не указана"
	if rateData != nil {
		rate := rateData.(*float64)
		if rate != nil {
			rateText = fmt.Sprintf("%s%%", FormatRate(*rate))
		}
	}

	text := fmt.Sprintf(`✅ Процентная ставка: %s

	Шаг 5/6: Введите дату окончания действия счета (необязательно)
	Форматы: 31.12.2025, 2025-12-31, 31/12/2025 
	
	Введите "пропустить" если не хотите указывать дату`, rateText)

	bot.sendMessage(session.ChatID, text)
}

func (h *SavingAccountCreationHandler) handleExpirationDate(bot *Bot, session *UserSession, input string) error {
	handler := ExpirationDateHandler{
		DataKey:        "expirationDate",
		NextStep:       StepConfirmation,
		ConfirmationFn: h.sendConfirmation,
		StoreAsPointer: true,
	}
	return HandleExpirationDate(bot, session, input, handler)
}

func (h *SavingAccountCreationHandler) sendConfirmation(bot *Bot, session *UserSession) {
	name := session.GetData("name").(string)
	amount := session.GetData("amount").(float64)
	currency := session.GetData("currency").(domain.CurrencyName)

	rateText := "Не указана"
	rateData := session.GetData("interestRate")
	if rateData != nil {
		rate := rateData.(*float64)
		if rate != nil {
			rateText = fmt.Sprintf("%s%%", FormatRate(*rate))
		}
	}

	dateText := "Не указана"
	dateData := session.GetData("expirationDate")
	if dateData != nil {
		date := dateData.(*time.Time)
		if date != nil {
			dateText = date.Format("02.01.2006")
		}
	}

	text := fmt.Sprintf(`💰 Подтверждение создания накопительного счета

	📝 Название: %s
	💰 Сумма: %s
	💱 Валюта: %s (%s)
	📈 Процентная ставка: %s
	📅 Дата окончания: %s
	
	🔍 Подтвердите создание накопительного счета:`,
		name,
		currency.FormatAmountRussian(amount),
		currency.ToHumanRussian(),
		currency.Symbol(),
		rateText,
		dateText)

	keyboard := CreateConfirmationKeyboard(CallbackConfirmSavingYes, CallbackConfirmSavingNo)
	bot.sendMessageWithKeyboard(session.ChatID, text, keyboard)
}

func (h *SavingAccountCreationHandler) handleConfirmation(ctx context.Context, bot *Bot, session *UserSession, input string) error {
	if input == CallbackConfirmSavingYes {
		bot.sessionManager.ClearSession(session.UserID)
		return h.CompleteSession(ctx, bot, session)
	}

	if input == CallbackConfirmSavingNo {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание накопительного счета отменено.")
		return nil
	}

	if IsNegativeResponse(input) {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание накопительного счета отменено.")
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

func (h *SavingAccountCreationHandler) CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, h.executeCompletion(ctx, bot, session)
	}

	name := session.GetData("name").(string)
	params := map[string]interface{}{
		"user_id":             int64(session.UserID),
		"saving_account_name": name,
	}

	wrappedHandler := bot.interceptor.Chain(handler, "SavingAccountCreationHandler.CompleteSession")
	_, err := wrappedHandler(ctx, params)
	return err
}

func (h *SavingAccountCreationHandler) executeCompletion(ctx context.Context, bot *Bot, session *UserSession) error {
	name := session.GetData("name").(string)
	amount := session.GetData("amount").(float64)
	currency := session.GetData("currency").(domain.CurrencyName)

	var interestRate *float64
	if rateData := session.GetData("interestRate"); rateData != nil {
		interestRate = rateData.(*float64)
	}

	var expirationDate *time.Time
	if dateData := session.GetData("expirationDate"); dateData != nil {
		expirationDate = dateData.(*time.Time)
	}

	req := &models.CreateSavingAccountRequest{
		UserID:              session.UserID,
		Name:                name,
		AmountRUB:           amount,
		Currency:            string(currency),
		InterestRatePercent: interestRate,
		ExpirationDate:      expirationDate,
	}

	account, err := bot.financeService.CreateSavingAccount(ctx, req)
	if err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ Ошибка при создании накопительного счета: %v", err))
		return err
	}

	rateText := "не указана"
	if account.InterestRateBasisPoints > 0 {
		rate := float64(account.InterestRateBasisPoints) / 100.0
		rateText = fmt.Sprintf("%s%%", FormatNumber(rate))
	}

	dateText := "не указана"
	if account.ExpirationDate != nil {
		dateText = account.ExpirationDate.Format("02.01.2006")
	}

	text := fmt.Sprintf(`✅ Накопительный счет успешно создан!

		📝 Название: %s
		💰 Сумма: %s
		📈 Процентная ставка: %s
		📅 Дата окончания: %s
		🆔 ID: %s
		
		Используйте /saving_accounts чтобы посмотреть все ваши накопительные счета.`,
		account.Name,
		currency.FormatAmountRussian(amount),
		rateText,
		dateText,
		FormatInteger(int64(account.Id)))

	bot.sendMessage(session.ChatID, text)
	bot.sendMainMenu(session.ChatID)
	return nil
}

func (h *SavingAccountCreationHandler) GetNextStep(currentStep SessionStep, input string) (SessionStep, error) {
	return StepComplete, nil
}

func (h *SavingAccountCreationHandler) ValidateInput(step SessionStep, input string) error {
	return nil
}

func (h *SavingAccountCreationHandler) FormatConfirmation(session *UserSession) string {
	return "Confirmation"
}
