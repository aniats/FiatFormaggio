package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type CashHoldingCreationHandler struct{}

func (h *CashHoldingCreationHandler) GetSessionType() SessionType {
	return SessionCreateCashHolding
}

func (h *CashHoldingCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "CashHoldingCreationHandler.HandleStep")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(session.UserID)),
		attribute.String("session.step", string(session.CurrentStep)),
		attribute.String("session.type", "create_cash_holding"),
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
	case StepConfirmation:
		return h.handleConfirmation(ctx, bot, session, msg.Text)
	default:
		return fmt.Errorf("неизвестный шаг: %s", session.CurrentStep)
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
		bot.sendMessage(session.ChatID, "❌ Неверный формат суммы. Введите число (например: 5000, 1500.50, 100):")
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

	if strings.ToLower(currencyStr) == "пропустить" || currencyStr == "" {
		session.SetData("currency", domain.RUB)
		session.CurrentStep = StepConfirmation
		h.sendConfirmation(bot, session)
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
	response := strings.ToLower(strings.TrimSpace(input))

	if response == "нет" || response == "отмена" {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание наличного счета отменено.")
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

func (h *CashHoldingCreationHandler) CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "CashHoldingCreationHandler.CompleteSession")
	defer span.End()

	name := session.GetData("name").(string)
	span.SetAttributes(
		attribute.Int64("user.id", int64(session.UserID)),
		attribute.String("cash_holding.name", name),
	)
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
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ Ошибка при создании наличного счета: %v", err))
		return err
	}

	text := fmt.Sprintf(`✅ Наличный счет успешно создан!

	📝 Название: %s
	💰 Сумма: %s

	Используйте /cash_holdings чтобы посмотреть все ваши наличные счета.`,
		cashHolding.Name,
		currency.FormatAmountRussian(amount))

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
