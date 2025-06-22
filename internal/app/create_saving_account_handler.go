package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type SavingAccountCreationHandler struct{}

func (h *SavingAccountCreationHandler) GetSessionType() SessionType {
	return SessionCreateSavingAccount
}

func (h *SavingAccountCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "SavingAccountCreationHandler.HandleStep")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(session.UserID)),
		attribute.String("session.step", string(session.CurrentStep)),
		attribute.String("session.type", "create_saving_account"),
	)

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
		return fmt.Errorf("неизвестный шаг: %s", session.CurrentStep)
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
	Например: 25000, 1000.50, 0
	
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
		bot.sendMessage(session.ChatID, "❌ Слишком большая сумма (максимум 1,000,000,000). Попробуйте еще раз:")
		return nil
	}

	session.SetData("amount", amount)
	session.CurrentStep = StepCurrency

	text := fmt.Sprintf(`✅ Сумма: %.2f

	Шаг 3/6: Выберите валюту счета
	Доступные варианты:
	• RUB, рубль - Российский рубль ₽
	• USD, доллар - Американский доллар $
	• EUR, евро - Евро €
	• CNY, юань - Китайский юань ¥
	• GBP, фунт - Британский фунт £
	
	По умолчанию: RUB (введите "пропустить" для RUB)`, amount)

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *SavingAccountCreationHandler) handleCurrency(bot *Bot, session *UserSession, input string) error {
	currencyStr := strings.TrimSpace(input)

	if strings.ToLower(currencyStr) == "пропустить" || currencyStr == "" {
		session.SetData("currency", domain.RUB)
		session.CurrentStep = StepInterestRate
		h.sendInterestRatePrompt(bot, session)
		return nil
	}

	currency, err := domain.CurrencyFromHuman(currencyStr)
	if err != nil {
		text := fmt.Sprintf(`❌ Неизвестная валюта "%s"

		Доступные варианты:
		• RUB, рубль, российский рубль
		• USD, доллар, американский доллар  
		• EUR, евро
		• CNY, юань, китайский юань
		• GBP, фунт, британский фунт
		
		Попробуйте еще раз:`, currencyStr)
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

	text := fmt.Sprintf(`✅ Валюта: %s (%s)
	✅ Сумма: %s

	Шаг 4/6: Введите процентную ставку (необязательно)
	Например: 5.5, 7.2, 4
	
	Введите "пропустить" если не хотите указывать ставку`,
		currency.ToHumanRussian(),
		currency.Symbol(),
		currency.FormatAmountRussian(amount))

	bot.sendMessage(session.ChatID, text)
}

func (h *SavingAccountCreationHandler) handleInterestRate(bot *Bot, session *UserSession, input string) error {
	rateStr := strings.TrimSpace(input)

	if strings.ToLower(rateStr) == "пропустить" || rateStr == "" {
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
			rateText = fmt.Sprintf("%.2f%%", *rate)
		}
	}

	text := fmt.Sprintf(`✅ Процентная ставка: %s

	Шаг 5/6: Введите дату окончания действия счета (необязательно)
	Форматы: 31.12.2025, 2025-12-31
	
	Введите "пропустить" если не хотите указывать дату`, rateText)

	bot.sendMessage(session.ChatID, text)
}

func (h *SavingAccountCreationHandler) handleExpirationDate(bot *Bot, session *UserSession, input string) error {
	dateStr := strings.TrimSpace(input)

	if strings.ToLower(dateStr) == "пропустить" || dateStr == "" {
		session.SetData("expirationDate", nil)
		session.CurrentStep = StepConfirmation
		h.sendConfirmation(bot, session)
		return nil
	}

	date, err := h.parseDate(dateStr)
	if err != nil {
		bot.sendMessage(session.ChatID, "❌ Некорректная дата. Используйте формат: 31.12.2025 или 2025-12-31:")
		return nil
	}

	if err := h.validateDate(date); err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ %s Попробуйте еще раз:", err.Error()))
		return nil
	}

	session.SetData("expirationDate", date)
	session.CurrentStep = StepConfirmation

	h.sendConfirmation(bot, session)
	return nil
}

func (h *SavingAccountCreationHandler) parseDate(input string) (*time.Time, error) {
	formats := []string{
		"02.01.2006",
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, input); err == nil {
			return &date, nil
		}
	}

	return nil, fmt.Errorf("неподдерживаемый формат даты")
}

func (h *SavingAccountCreationHandler) validateDate(date *time.Time) error {
	if date.Before(time.Now()) {
		return fmt.Errorf("дата не может быть в прошлом")
	}

	maxDate := time.Now().AddDate(10, 0, 0)
	if date.After(maxDate) {
		return fmt.Errorf("дата не может быть более чем через 10 лет")
	}

	return nil
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
			rateText = fmt.Sprintf("%.2f%%", *rate)
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

	Все верно? Отправьте "да" для создания счета или "нет" для отмены.`,
		name,
		currency.FormatAmountRussian(amount),
		currency.ToHumanRussian(),
		currency.Symbol(),
		rateText,
		dateText)

	bot.sendMessage(session.ChatID, text)
}

func (h *SavingAccountCreationHandler) handleConfirmation(ctx context.Context, bot *Bot, session *UserSession, input string) error {
	response := strings.ToLower(strings.TrimSpace(input))

	if response == "нет" || response == "отмена" {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание накопительного счета отменено.")
		return nil
	}

	if response != "да" && response != "yes" && response != "подтверждаю" {
		bot.sendMessage(session.ChatID, "Пожалуйста, ответьте 'да' для подтверждения или 'нет' для отмены:")
		return nil
	}

	// Prevent double execution by clearing session first
	bot.sessionManager.ClearSession(session.UserID)
	return h.CompleteSession(ctx, bot, session)
}

func (h *SavingAccountCreationHandler) CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "SavingAccountCreationHandler.CompleteSession")
	defer span.End()

	name := session.GetData("name").(string)
	span.SetAttributes(
		attribute.Int64("user.id", int64(session.UserID)),
		attribute.String("saving_account.name", name),
	)
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

	rateText := "Не указана"
	if interestRate != nil {
		rateText = fmt.Sprintf("%.2f%%", *interestRate)
	}

	dateText := "Не указана"
	if expirationDate != nil {
		dateText = expirationDate.Format("02.01.2006")
	}

	text := fmt.Sprintf(`✅ Накопительный счет успешно создан!

	📝 Название: %s
	💰 Сумма: %s
	📈 Процентная ставка: %s
	📅 Дата окончания: %s

	Используйте /saving_accounts чтобы посмотреть все ваши накопительные счета.`,
		account.Name,
		currency.FormatAmountRussian(amount),
		rateText,
		dateText)

	bot.sendMessage(session.ChatID, text)
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
