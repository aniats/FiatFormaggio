package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (b *Bot) sendTotalBalanceCommand(ctx context.Context, chatID int64, userID domain.UserId) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		params := input.(map[string]interface{})
		chatID := params["chatID"].(int64)
		userID := params["userID"].(domain.UserId)
		
		return b.processTotalBalanceCommand(ctx, chatID, userID)
	}
	
	params := map[string]interface{}{
		"chatID": chatID,
		"userID": userID,
	}
	
	wrappedHandler := b.interceptor.Chain(handler, "Bot.sendTotalBalanceCommand")
	_, _ = wrappedHandler(ctx, params)
}

func (b *Bot) processTotalBalanceCommand(ctx context.Context, chatID int64, userID domain.UserId) (interface{}, error) {

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
			message += fmt.Sprintf("  %s: %s ₽\n", currency, FormatRate(rate))
		}
		message += "\n"
	}

	if deposits, err := b.financeService.GetDepositsByUserID(ctx, userID); err == nil && len(deposits) > 0 {
		summary, totalRUB := calculateDepositsSummary(deposits, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	if savingAccounts, err := b.financeService.GetSavingAccountsByUserID(ctx, userID); err == nil && len(savingAccounts) > 0 {
		summary, totalRUB := calculateSavingAccountsSummary(savingAccounts, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	if brokerageAccounts, err := b.financeService.GetBrokerageAccountsByUserID(ctx, userID); err == nil && len(brokerageAccounts) > 0 {
		summary, totalRUB := calculateBrokerageAccountsSummary(brokerageAccounts, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	if cashHoldings, err := b.financeService.GetCashHoldingsByUserID(ctx, userID); err == nil && len(cashHoldings) > 0 {
		summary, totalRUB := calculateCashHoldingsSummary(cashHoldings, exchangeRates)
		message += summary
		grandTotalRUB += totalRUB
	}

	message += fmt.Sprintf("🎯 ИТОГО: %s ₽\n", FormatNumber(grandTotalRUB))
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
			exchangeRates[rate.Currency] = float64(rate.RateMinorUnits) / DefaultMinorUnits
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
		amount := float64(deposit.AmountMinorUnits) / DefaultMinorUnits
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
		amount := float64(account.AmountMinorUnits) / DefaultMinorUnits
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
		amount := float64(account.AmountMinorUnits) / DefaultMinorUnits
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
		amount := float64(holding.AmountMinorUnits) / DefaultMinorUnits
		currency := holding.Currency
		amountRUB := convertToRUB(amount, currency, exchangeRates)

		currencyTotals[currency] += amount
		totalRUB += amountRUB
	}

	return formatAccountSummary("Наличные", int64(len(holdings)), currencyTotals, totalRUB, exchangeRates), totalRUB
}

func formatAccountSummary(accountType string, count int64, currencyTotals map[domain.CurrencyName]float64, totalRUB float64, exchangeRates map[domain.CurrencyName]float64) string {
	message := fmt.Sprintf("📊 %s (%s):\n", accountType, FormatInteger(count))

	var currencies []domain.CurrencyName
	for currency := range currencyTotals {
		currencies = append(currencies, currency)
	}
	sort.Slice(currencies, func(i, j int) bool {
		return string(currencies[i]) < string(currencies[j])
	})

	for _, currency := range currencies {
		amount := currencyTotals[currency]
		message += fmt.Sprintf("  %s: %s %s", currency, FormatNumber(amount), currency.Symbol())

		if currency != domain.RUB {
			rubEquivalent := amount * exchangeRates[currency]
			message += fmt.Sprintf(" (%s ₽)", FormatNumber(rubEquivalent))
		}
		message += "\n"
	}

	message += fmt.Sprintf("  Всего: %s ₽\n\n", FormatNumber(totalRUB))
	return message
}

