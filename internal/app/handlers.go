package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

func (b *Bot) handleDepositsCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	deposits, err := b.financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error getting deposits for user %d: %v", userID, err)
		b.sendMessage(chatID, "Error retrieving deposits. Please try again later.")
		return
	}

	if len(deposits) == 0 {
		b.sendMessage(chatID, "You have no deposits.")
		return
	}

	text := "Your deposits:\n"
	for _, deposit := range deposits {
		amount := float64(deposit.AmountMinorUnits) / 100.0
		text += fmt.Sprintf("• %s: %.2f %s\n", deposit.Name, amount, deposit.Currency)
	}

	b.sendMessage(chatID, text)
}

func (b *Bot) handleCreateDeposit(ctx context.Context, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())

	if len(args) == 0 {
		b.sendCreateDepositHelp(msg.Chat.ID)
		return
	}

	depositReq, err := b.parseCreateDepositArgs(args, domain.UserId(msg.From.ID))
	if err != nil {
		b.sendMessage(msg.Chat.ID, fmt.Sprintf("❌ Ошибка в параметрах: %s\n\nИспользуйте: /createdeposit <название> <сумма> [процент] [дата_окончания] [валюта]", err.Error()))
		b.sendCreateDepositHelp(msg.Chat.ID)
		return
	}

	deposit, err := b.financeService.CreateDeposit(ctx, depositReq)
	if err != nil {
		log.Printf("Error creating deposit for user %d: %v", msg.From.ID, err)
		b.sendMessage(msg.Chat.ID, "❌ Ошибка при создании депозита. Попробуйте позже.")
		return
	}

	b.sendDepositCreatedConfirmation(msg.Chat.ID, deposit)
}

// parseCreateDepositArgs парсит аргументы команды создания депозита
func (b *Bot) parseCreateDepositArgs(args []string, userID domain.UserId) (*models.CreateDepositRequest, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("недостаточно аргументов")
	}

	req := &models.CreateDepositRequest{
		UserID:   userID,
		Currency: "RUB",
	}

	var nameEnd int
	var amountStr string

	for i, arg := range args {
		if _, err := strconv.ParseFloat(arg, 64); err == nil {
			nameEnd = i
			amountStr = arg
			break
		}
	}

	if nameEnd == 0 {
		return nil, fmt.Errorf("не найдена сумма депозита")
	}

	req.Name = strings.Join(args[:nameEnd], " ")
	if req.Name == "" {
		return nil, fmt.Errorf("название не может быть пустым")
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("некорректная сумма: %s", amountStr)
	}
	req.AmountRUB = amount

	remainingArgs := args[nameEnd+1:]

	for i, arg := range remainingArgs {
		if val, err := strconv.ParseFloat(arg, 64); err == nil && val <= 100 && strings.Contains(arg, ".") {
			req.InterestRatePercent = &val
			continue
		}

		// Если это дата в формате DD.MM.YYYY или YYYY-MM-DD
		if date, err := b.parseDate(arg); err == nil {
			req.ExpirationDate = &date
			continue
		}

		if len(arg) == 3 && strings.ToUpper(arg) == arg {
			req.Currency = strings.ToUpper(arg)
			continue
		}

		if i == 0 {
			req.Name += " " + arg
		}
	}

	return req, nil
}

func (b *Bot) parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"02.01.2006", // DD.MM.YYYY
		"2006-01-02", // YYYY-MM-DD
		"02/01/2006", // DD/MM/YYYY
		"01/02/2006", // MM/DD/YYYY
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("неверный формат даты")
}

func (b *Bot) sendCreateDepositHelp(chatID int64) {
	helpText := `📝 Создание депозита:

		Формат: /createdeposit <название> <сумма> [процент] [дата_окончания] [валюта]
		
		Примеры:
		• /createdeposit "Сбербанк Депозит" 100000
		• /createdeposit "ВТБ Вклад" 50000 5.5
		• /createdeposit "Альфа Депозит" 25000 4.2 31.12.2025
		• /createdeposit "USD Deposit" 1000 3.0 2025-12-31 USD
		
		Параметры:
		• название - название депозита (в кавычках если содержит пробелы)
		• сумма - сумма в указанной валюте
		• процент - годовая процентная ставка (опционально)
		• дата_окончания - дата в формате DD.MM.YYYY или YYYY-MM-DD (опционально)
		• валюта - RUB, USD, EUR, CNY (по умолчанию RUB)
		
		❗ Кавычки в названии необязательны если оно из одного слова.
	`

	b.sendMessage(chatID, helpText)
}

func (b *Bot) sendDepositCreatedConfirmation(chatID int64, deposit *domain.Deposit) {
	amount := float64(deposit.AmountMinorUnits) / 100.0

	text := fmt.Sprintf("✅ Депозит успешно создан!\n\n")
	text += fmt.Sprintf("📋 Название: %s\n", deposit.Name)
	text += fmt.Sprintf("💰 Сумма: %.2f %s\n", amount, deposit.Currency)

	rate := float64(deposit.InterestRateBasisPoints) / 100.0
	text += fmt.Sprintf("📈 Процентная ставка: %.2f%%\n", rate)

	if deposit.ExpirationDate != nil {
		text += fmt.Sprintf("📅 Дата окончания: %s\n", deposit.ExpirationDate.Format("02.01.2006"))
	}

	text += fmt.Sprintf("🆔 ID: %d\n", deposit.Id)
	text += fmt.Sprintf("📅 Создан: %s", time.Now().Format("02.01.2006 15:04"))

	b.sendMessage(chatID, text)
}
