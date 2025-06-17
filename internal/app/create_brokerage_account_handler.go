package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

type BrokerageAccountCreationHandler struct{}

func (h *BrokerageAccountCreationHandler) GetSessionType() SessionType {
	return SessionCreateBrokerageAccountSession
}

func (h *BrokerageAccountCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *tgBotAPI.Message) error {
	switch session.CurrentStep {
	case StepStart:
		return h.handleStart(bot, session)
	case StepName:
		return h.handleName(bot, session, msg.Text)
	case StepAmount:
		return h.handleAmount(bot, session, msg.Text)
	case StepCurrency:
		return h.handleCurrency(bot, session, msg.Text)
	case StepAccount:
		return h.handleBroker(bot, session, msg.Text)
	case StepCategory:
		return h.handleAccountType(bot, session, msg.Text)
	case StepConfirmation:
		return h.handleConfirmation(ctx, bot, session, msg.Text)
	default:
		return fmt.Errorf("неизвестный шаг: %s", session.CurrentStep)
	}
}

func (h *BrokerageAccountCreationHandler) handleStart(bot *Bot, session *UserSession) error {
	text := `📈 Создание нового брокерского счета

	Шаг 1/6: Введите название счета
	Например: "Тинькофф Инвестиции", "Портфель роста"
	
	Для отмены введите /cancel`

	bot.sendMessage(session.ChatID, text)
	session.CurrentStep = StepName
	return nil
}

func (h *BrokerageAccountCreationHandler) handleName(bot *Bot, session *UserSession, input string) error {
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
	Например: 50000, 1500.50, 0
	
	Валюта будет указана на следующем шаге.`, name)

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *BrokerageAccountCreationHandler) handleAmount(bot *Bot, session *UserSession, input string) error {
	amountStr := strings.TrimSpace(strings.Replace(input, ",", ".", -1))
	
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		bot.sendMessage(session.ChatID, "❌ Неверный формат суммы. Введите число (например: 50000, 1500.50, 0):")
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

func (h *BrokerageAccountCreationHandler) handleCurrency(bot *Bot, session *UserSession, input string) error {
	currencyStr := strings.TrimSpace(input)

	if strings.ToLower(currencyStr) == "пропустить" || currencyStr == "" {
		session.SetData("currency", domain.RUB)
		session.CurrentStep = StepAccount
		h.sendBrokerPrompt(bot, session)
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
	session.CurrentStep = StepAccount

	h.sendBrokerPrompt(bot, session)
	return nil
}

func (h *BrokerageAccountCreationHandler) sendBrokerPrompt(bot *Bot, session *UserSession) {
	currency := session.GetData("currency").(domain.CurrencyName)
	amount := session.GetData("amount").(float64)
	
	text := fmt.Sprintf(`✅ Валюта: %s (%s)
✅ Сумма: %s

	Шаг 4/6: Введите название брокера (необязательно)
	Например: "Тинькофф", "Сбербанк", "ВТБ", "Альфа-Банк"
	
	Введите "пропустить" если не хотите указывать брокера`, 
		currency.ToHumanRussian(), 
		currency.Symbol(),
		currency.FormatAmountRussian(amount))

	bot.sendMessage(session.ChatID, text)
}

func (h *BrokerageAccountCreationHandler) handleBroker(bot *Bot, session *UserSession, input string) error {
	brokerInput := strings.TrimSpace(input)
	
	var broker *string
	if strings.ToLower(brokerInput) == "пропустить" || brokerInput == "" {
		broker = nil
	} else {
		if len(brokerInput) > 255 {
			bot.sendMessage(session.ChatID, "❌ Название брокера слишком длинное (максимум 255 символов). Попробуйте еще раз:")
			return nil
		}
		broker = &brokerInput
	}

	session.SetData("broker", broker)
	session.CurrentStep = StepCategory

	text := `Шаг 5/6: Выберите тип брокерского счета
	
	Доступные типы:
	• regular, обычный - Обычный брокерский счет
	• iis, иис - Индивидуальный инвестиционный счет
	• iis3, иис3 - ИИС третьего типа
	• ira, ира - Индивидуальный пенсионный план
	• margin, маржинальный - Маржинальный счет
	
	По умолчанию: regular (введите "пропустить" для обычного счета)`

	bot.sendMessage(session.ChatID, text)
	return nil
}

func (h *BrokerageAccountCreationHandler) handleAccountType(bot *Bot, session *UserSession, input string) error {
	accountTypeStr := strings.TrimSpace(input)

	var accountType domain.BrokerageType
	if strings.ToLower(accountTypeStr) == "пропустить" || accountTypeStr == "" {
		accountType = domain.Regular
	} else {
		var err error
		accountType, err = h.parseAccountType(accountTypeStr)
		if err != nil {
			text := fmt.Sprintf(`❌ Неизвестный тип счета "%s"

			Доступные варианты:
			• regular, обычный - Обычный брокерский счет
			• iis, иис - Индивидуальный инвестиционный счет  
			• iis3, иис3 - ИИС третьего типа
			• ira, ира - Индивидуальный пенсионный план
			• margin, маржинальный - Маржинальный счет
			
			Попробуйте еще раз:`, accountTypeStr)
			bot.sendMessage(session.ChatID, text)
			return nil
		}
	}

	session.SetData("accountType", accountType)
	session.CurrentStep = StepConfirmation

	h.sendConfirmation(bot, session)
	return nil
}

func (h *BrokerageAccountCreationHandler) parseAccountType(input string) (domain.BrokerageType, error) {
	input = strings.ToLower(strings.TrimSpace(input))
	
	switch input {
	case "regular", "обычный":
		return domain.Regular, nil
	case "iis", "иис":
		return domain.IIS, nil
	case "iis3", "иис3":
		return domain.IIS3, nil
	case "ira", "ира":
		return domain.IRA, nil
	case "margin", "маржинальный":
		return domain.Margin, nil
	default:
		return "", fmt.Errorf("unsupported account type: %s", input)
	}
}

func (h *BrokerageAccountCreationHandler) sendConfirmation(bot *Bot, session *UserSession) {
	name := session.GetData("name").(string)
	amount := session.GetData("amount").(float64)
	currency := session.GetData("currency").(domain.CurrencyName)
	broker := session.GetData("broker").(*string)
	accountType := session.GetData("accountType").(domain.BrokerageType)

	brokerText := "Не указан"
	if broker != nil {
		brokerText = *broker
	}

	accountTypeText := h.formatAccountType(accountType)

	text := fmt.Sprintf(`📈 Подтверждение создания брокерского счета

	📝 Название: %s
	💰 Сумма: %s
	💱 Валюта: %s (%s)
	🏦 Брокер: %s
	📊 Тип счета: %s

	Все верно? Отправьте "да" для создания счета или "нет" для отмены.`,
		name,
		currency.FormatAmountRussian(amount),
		currency.ToHumanRussian(),
		currency.Symbol(),
		brokerText,
		accountTypeText)

	bot.sendMessage(session.ChatID, text)
}

func (h *BrokerageAccountCreationHandler) formatAccountType(accountType domain.BrokerageType) string {
	switch accountType {
	case domain.Regular:
		return "Обычный"
	case domain.IIS:
		return "ИИС"
	case domain.IIS3:
		return "ИИС-3"
	case domain.IRA:
		return "ИРА"
	case domain.Margin:
		return "Маржинальный"
	default:
		return string(accountType)
	}
}

func (h *BrokerageAccountCreationHandler) handleConfirmation(ctx context.Context, bot *Bot, session *UserSession, input string) error {
	response := strings.ToLower(strings.TrimSpace(input))
	
	if response == "нет" || response == "отмена" {
		bot.sessionManager.ClearSession(session.UserID)
		bot.sendMessage(session.ChatID, "❌ Создание брокерского счета отменено.")
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

func (h *BrokerageAccountCreationHandler) CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error {
	name := session.GetData("name").(string)
	amount := session.GetData("amount").(float64)
	currency := session.GetData("currency").(domain.CurrencyName)
	broker := session.GetData("broker").(*string)
	accountType := session.GetData("accountType").(domain.BrokerageType)

	req := &models.CreateBrokerageAccountRequest{
		UserID:      session.UserID,
		Name:        name,
		AmountRUB:   amount,
		Currency:    string(currency),
		Broker:      broker,
		AccountType: string(accountType),
	}

	account, err := bot.financeService.CreateBrokerageAccount(ctx, req)
	if err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ Ошибка при создании брокерского счета: %v", err))
		return err
	}

	brokerText := "Не указан"
	if broker != nil {
		brokerText = *broker
	}

	text := fmt.Sprintf(`✅ Брокерский счет успешно создан!

	📝 Название: %s
	💰 Сумма: %s
	🏦 Брокер: %s
	📊 Тип: %s

	Используйте /brokerage_accounts чтобы посмотреть все ваши брокерские счета.`,
		account.Name,
		currency.FormatAmountRussian(amount),
		brokerText,
		h.formatAccountType(account.AccountType))

	bot.sendMessage(session.ChatID, text)
	return nil
}

// Implement remaining SessionHandler interface methods (required by interface)
func (h *BrokerageAccountCreationHandler) GetNextStep(currentStep SessionStep, input string) (SessionStep, error) {
	return StepComplete, nil
}

func (h *BrokerageAccountCreationHandler) ValidateInput(step SessionStep, input string) error {
	return nil
}

func (h *BrokerageAccountCreationHandler) FormatConfirmation(session *UserSession) string {
	return "Confirmation"
}