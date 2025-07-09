package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/utils"
	"sort"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (b *Bot) sendTotalBalanceCommand(ctx context.Context, chatID int64, UserID domain.UserID) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		params := input.(map[string]interface{})
		chatID := params["chatID"].(int64)
		UserID := params["UserID"].(domain.UserID)

		return b.processTotalBalanceCommand(ctx, chatID, UserID)
	}

	params := map[string]interface{}{
		"chatID": chatID,
		"UserID": UserID,
	}

	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendTotalBalanceCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processTotalBalanceCommand(ctx context.Context, chatID int64, UserID domain.UserID) (interface{}, error) {
	rates, err := b.currencyService.GetCurrencyRates(ctx)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка получения курсов валют. Попробуйте позже.")
		return nil, err
	}

	exchangeRates := buildExchangeRateMap(rates)
	lastUpdateTime := "неизвестно"
	if len(rates) > 0 {
		lastUpdateTime = rates[0].UpdatedAt.Format("02.01.2006 15:04")
	}

	var grandTotalRUB float64
	message := fmt.Sprintf("💰 ОБЩИЙ БАЛАНС 💰\n")
	message += fmt.Sprintf("🕐 Курсы на: %s\n\n", lastUpdateTime)

	if len(exchangeRates) > 1 {
		message += "📈 Курсы валют (ЦБ РФ):\n"
		var currencies []domain.CurrencyName
		for currency := range exchangeRates {
			if currency != domain.RUB {
				currencies = append(currencies, currency)
			}
		}
		sort.Slice(currencies, func(i, j int) bool {
			return string(currencies[i]) < string(currencies[j])
		})

		for _, currency := range currencies {
			rate := exchangeRates[currency]
			message += fmt.Sprintf("  %s: %s ₽\n", currency, utils.FormatRate(rate))
		}
		message += "\n"
	}

	if deposits, err := b.financeService.GetDepositsByUserID(ctx, UserID); err == nil && len(deposits) > 0 {
		summary, totalRUB := calculateDepositsSummary(deposits, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	if savingAccounts, err := b.financeService.GetSavingAccountsByUserID(ctx, UserID); err == nil && len(savingAccounts) > 0 {
		summary, totalRUB := calculateSavingAccountsSummary(savingAccounts, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	if brokerageAccounts, err := b.financeService.GetBrokerageAccountsByUserID(ctx, UserID); err == nil && len(brokerageAccounts) > 0 {
		summary, totalRUB := calculateBrokerageAccountsSummary(brokerageAccounts, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	if cashHoldings, err := b.financeService.GetCashHoldingsByUserID(ctx, UserID); err == nil && len(cashHoldings) > 0 {
		summary, totalRUB := calculateCashHoldingsSummary(cashHoldings, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	message += fmt.Sprintf("🎯 ИТОГО: %s ₽\n", utils.FormatNumber(grandTotalRUB))
	b.sendMessage(chatID, message)
	b.sendMainMenu(chatID)
	return nil, nil
}

func buildExchangeRateMap(rates []domain.CurrencyRate) map[domain.CurrencyName]float64 {
	exchangeRates := map[domain.CurrencyName]float64{
		domain.RUB: 1.0,
	}

	for _, rate := range rates {
		if rate.BaseCurrency == domain.RUB {
			exchangeRates[rate.Currency] = float64(rate.RateMinorUnits) / utils.DefaultMinorUnits
		}
	}

	return exchangeRates
}

func convertToRUB(amount float64, currency domain.CurrencyName, exchangeRates map[domain.CurrencyName]float64) float64 {
	if currency == domain.RUB {
		return amount
	}
	if rate, exists := exchangeRates[currency]; exists {
		return amount * rate
	}
	return amount
}

func calculateDepositsSummary(deposits []domain.Deposit, exchangeRates map[domain.CurrencyName]float64) (string, float64) {
	currencyTotals := make(map[domain.CurrencyName]float64)
	var totalRUB float64

	for _, deposit := range deposits {
		amount := float64(deposit.AmountMinorUnits) / utils.DefaultMinorUnits
		currency := deposit.Currency
		amountRUB := convertToRUB(amount, currency, exchangeRates)

		currencyTotals[currency] += amount
		totalRUB += amountRUB
	}

	return formatAccountSummary("Депозиты", int64(len(deposits)), currencyTotals, totalRUB, exchangeRates), totalRUB
}

func calculateSavingAccountsSummary(accounts []domain.SavingAccount, exchangeRates map[domain.CurrencyName]float64) (string, float64) {
	currencyTotals := make(map[domain.CurrencyName]float64)
	var totalRUB float64

	for _, account := range accounts {
		amount := float64(account.AmountMinorUnits) / utils.DefaultMinorUnits
		currency := account.Currency
		amountRUB := convertToRUB(amount, currency, exchangeRates)

		currencyTotals[currency] += amount
		totalRUB += amountRUB
	}

	return formatAccountSummary("Накопительные счета", int64(len(accounts)), currencyTotals, totalRUB, exchangeRates), totalRUB
}

func calculateBrokerageAccountsSummary(accounts []domain.BrokerageAccount, exchangeRates map[domain.CurrencyName]float64) (string, float64) {
	currencyTotals := make(map[domain.CurrencyName]float64)
	var totalRUB float64

	for _, account := range accounts {
		amount := float64(account.AmountMinorUnits) / utils.DefaultMinorUnits
		currency := account.Currency
		amountRUB := convertToRUB(amount, currency, exchangeRates)

		currencyTotals[currency] += amount
		totalRUB += amountRUB
	}

	return formatAccountSummary("Брокерские счета", int64(len(accounts)), currencyTotals, totalRUB, exchangeRates), totalRUB
}

func calculateCashHoldingsSummary(holdings []domain.CashHolding, exchangeRates map[domain.CurrencyName]float64) (string, float64) {
	currencyTotals := make(map[domain.CurrencyName]float64)
	var totalRUB float64

	for _, holding := range holdings {
		amount := float64(holding.AmountMinorUnits) / utils.DefaultMinorUnits
		currency := holding.Currency
		amountRUB := convertToRUB(amount, currency, exchangeRates)

		currencyTotals[currency] += amount
		totalRUB += amountRUB
	}

	return formatAccountSummary("Наличные", int64(len(holdings)), currencyTotals, totalRUB, exchangeRates), totalRUB
}

func formatAccountSummary(accountType string, count int64, currencyTotals map[domain.CurrencyName]float64, totalRUB float64, exchangeRates map[domain.CurrencyName]float64) string {
	message := fmt.Sprintf("📊 %s (%s):\n", accountType, utils.FormatInteger(count))

	var currencies []domain.CurrencyName
	for currency := range currencyTotals {
		currencies = append(currencies, currency)
	}
	sort.Slice(currencies, func(i, j int) bool {
		return string(currencies[i]) < string(currencies[j])
	})

	for _, currency := range currencies {
		amount := currencyTotals[currency]
		message += fmt.Sprintf("  %s: %s %s", currency, utils.FormatNumber(amount), currency.Symbol())

		if currency != domain.RUB {
			rubEquivalent := amount * exchangeRates[currency]
			message += fmt.Sprintf(" (%s ₽)", utils.FormatNumber(rubEquivalent))
		}
		message += "\n"
	}

	message += fmt.Sprintf("  Всего: %s ₽\n\n", utils.FormatNumber(totalRUB))
	return message
}
