package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/errors"
)

func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"02.01.2006",
		"2006-01-02",
		"02/01/2006",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, errors.ErrInvalidDateFormat
}

func ValidateDate(date time.Time) error {
	if date.Before(time.Now()) {
		return errors.ErrDateInPast
	}
	maxDate := time.Now().AddDate(10, 0, 0)
	if date.After(maxDate) {
		return errors.ErrDateTooFar
	}
	return nil
}

type ExpirationDateHandler struct {
	DataKey        string
	NextStep       SessionStep
	ConfirmationFn func(*Bot, *UserSession)
	StoreAsPointer bool
}

func HandleExpirationDate(bot *Bot, session *UserSession, input string, handler ExpirationDateHandler) error {
	dateStr := strings.TrimSpace(input)

	if IsSkipResponse(dateStr) {
		if handler.StoreAsPointer {
			session.SetData(handler.DataKey, nil)
		}
		session.CurrentStep = handler.NextStep
		handler.ConfirmationFn(bot, session)
		return nil
	}

	date, err := ParseDate(dateStr)
	if err != nil {
		bot.sendMessage(session.ChatID, "❌ Некорректная дата. Используйте формат: 31.12.2025, 2025-12-31, 31/12/2025:")
		return nil
	}

	if err := ValidateDate(date); err != nil {
		bot.sendMessage(session.ChatID, fmt.Sprintf("❌ %s Попробуйте еще раз:", err.Error()))
		return nil
	}

	if handler.StoreAsPointer {
		session.SetData(handler.DataKey, &date)
	} else {
		session.SetData(handler.DataKey, date)
	}
	session.CurrentStep = handler.NextStep

	handler.ConfirmationFn(bot, session)
	return nil
}
