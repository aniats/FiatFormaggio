package chatgpt

import (
	"context"
	"fmt"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/tracing"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Client struct {
	client *openai.Client
}

type FinanceDataProvider interface {
	GetDepositsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.Deposit, error)
	GetBrokerageAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.BrokerageAccount, error)
	GetSavingAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.SavingAccount, error)
	GetCashHoldingsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.CashHolding, error)
	GetTotalBalance(ctx context.Context, UserID domain.UserID) (float64, error)
}

type CurrencyService interface {
	GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error)
}

type ChatMessage struct {
	Role    string
	Content string
}

type FinancialContext struct {
	TotalBalanceRUB string
	Assets          []string
	CurrencyRates   map[string]float64
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c *Client) GetFinancialAdvice(ctx context.Context, financialContext *FinancialContext, userMessage string, previousMessages []ChatMessage) (string, error) {
	tracer := otel.Tracer("chatgpt-service")
	ctx, span := tracer.Start(ctx, "ChatGPT.GetFinancialAdvice")
	defer span.End()

	span.SetAttributes(
		attribute.String(tracing.UserMessage, userMessage),
		attribute.Int(tracing.PreviousMessagesCount, len(previousMessages)),
		attribute.String(tracing.TotalBalance, financialContext.TotalBalanceRUB),
		attribute.Int(tracing.FinancialAssetsCount, len(financialContext.Assets)),
	)

	// Build the system message with financial context
	systemMessage := c.buildSystemMessage(financialContext)

	// Convert chat messages to OpenAI format
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessage,
		},
	}

	// Add previous messages
	for _, msg := range previousMessages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	// Make the API call
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    messages,
		MaxTokens:   500,
		Temperature: 0.7,
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get ChatGPT response")
		return "", fmt.Errorf("failed to get ChatGPT response: %w", err)
	}

	if len(resp.Choices) == 0 {
		span.SetStatus(codes.Error, "No response from ChatGPT")
		return "", fmt.Errorf("no response from ChatGPT")
	}

	response := resp.Choices[0].Message.Content
	span.SetAttributes(
		attribute.Int(tracing.ResponseLength, len(response)),
	)
	span.SetStatus(codes.Ok, "Financial advice provided successfully")

	return response, nil
}

func (c *Client) buildSystemMessage(financialContext *FinancialContext) string {
	var systemMessage strings.Builder

	systemMessage.WriteString("Ты — персональный финансовый консультант. Отвечай на русском языке.\n\n")
	systemMessage.WriteString("Информация о портфеле пользователя:\n")
	systemMessage.WriteString(fmt.Sprintf("• Общий баланс: %s ₽\n", financialContext.TotalBalanceRUB))

	if len(financialContext.Assets) > 0 {
		systemMessage.WriteString("• Активы:\n")
		for _, asset := range financialContext.Assets {
			systemMessage.WriteString(fmt.Sprintf("  - %s\n", asset))
		}
	}

	if len(financialContext.CurrencyRates) > 0 {
		systemMessage.WriteString("• Актуальные курсы валют (к рублю):\n")
		for currency, rate := range financialContext.CurrencyRates {
			systemMessage.WriteString(fmt.Sprintf("  - %s: %.4f ₽\n", currency, rate))
		}
	}

	systemMessage.WriteString("\nДавай персонализированные советы основываясь на этой информации. ")
	systemMessage.WriteString("Будь кратким и практичным. Не давай советы по покупке конкретных акций или криптовалют.")

	return systemMessage.String()
}

func (c *Client) GetFinancialAdviceWithContext(ctx context.Context, userID domain.UserID, userMessage string, previousMessages []ChatMessage, financeProvider FinanceDataProvider, currencyService CurrencyService) (string, error) {
	tracer := otel.Tracer("chatgpt-service")
	ctx, span := tracer.Start(ctx, "ChatGPT.GetFinancialAdviceWithContext")
	defer span.End()

	span.SetAttributes(
		attribute.Int64(tracing.UserID, int64(userID)),
		attribute.String(tracing.UserMessage, userMessage),
		attribute.Int(tracing.PreviousMessagesCount, len(previousMessages)),
	)

	financialContext, err := c.buildFinancialContext(ctx, userID, financeProvider, currencyService)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to build financial context")
		return "", fmt.Errorf("failed to build financial context: %w", err)
	}

	response, err := c.GetFinancialAdvice(ctx, financialContext, userMessage, previousMessages)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get AI advice")
		return "", fmt.Errorf("failed to get AI advice: %w", err)
	}

	span.SetAttributes(
		attribute.Int(tracing.ResponseLength, len(response)),
		attribute.Int(tracing.FinancialAssetsCount, len(financialContext.Assets)),
	)
	span.SetStatus(codes.Ok, "Financial advice provided successfully")

	return response, nil
}

func (c *Client) buildFinancialContext(ctx context.Context, userID domain.UserID, financeProvider FinanceDataProvider, currencyService CurrencyService) (*FinancialContext, error) {
	tracer := otel.Tracer("chatgpt-service")
	ctx, span := tracer.Start(ctx, "ChatGPT.buildFinancialContext")
	defer span.End()

	span.SetAttributes(attribute.Int64(tracing.UserID, int64(userID)))

	totalBalance, err := financeProvider.GetTotalBalance(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get total balance")
		return nil, fmt.Errorf("failed to get total balance: %w", err)
	}

	span.SetAttributes(attribute.String(tracing.TotalBalance, fmt.Sprintf("%.2f", totalBalance)))

	var assets []string

	deposits, err := financeProvider.GetDepositsByUserID(ctx, userID)
	if err == nil && len(deposits) > 0 {
		currencyBreakdown := make(map[string]float64)
		for _, deposit := range deposits {
			amount := float64(deposit.AmountMinorUnits) / 100.0
			currencyBreakdown[string(deposit.Currency)] += amount
		}

		var assetDetails []string
		for currency, amount := range currencyBreakdown {
			assetDetails = append(assetDetails, fmt.Sprintf("%.2f %s", amount, currency))
		}
		assets = append(assets, fmt.Sprintf("Депозиты (%d счетов): %s", len(deposits), strings.Join(assetDetails, ", ")))
	}

	brokerageAccounts, err := financeProvider.GetBrokerageAccountsByUserID(ctx, userID)
	if err == nil && len(brokerageAccounts) > 0 {
		currencyBreakdown := make(map[string]float64)
		for _, account := range brokerageAccounts {
			amount := float64(account.AmountMinorUnits) / 100.0
			currencyBreakdown[string(account.Currency)] += amount
		}

		var assetDetails []string
		for currency, amount := range currencyBreakdown {
			assetDetails = append(assetDetails, fmt.Sprintf("%.2f %s", amount, currency))
		}
		assets = append(assets, fmt.Sprintf("Брокерские счета (%d счетов): %s", len(brokerageAccounts), strings.Join(assetDetails, ", ")))
	}

	savingAccounts, err := financeProvider.GetSavingAccountsByUserID(ctx, userID)
	if err == nil && len(savingAccounts) > 0 {
		currencyBreakdown := make(map[string]float64)
		for _, account := range savingAccounts {
			amount := float64(account.AmountMinorUnits) / 100.0
			currencyBreakdown[string(account.Currency)] += amount
		}

		var assetDetails []string
		for currency, amount := range currencyBreakdown {
			assetDetails = append(assetDetails, fmt.Sprintf("%.2f %s", amount, currency))
		}
		assets = append(assets, fmt.Sprintf("Накопительные счета (%d счетов): %s", len(savingAccounts), strings.Join(assetDetails, ", ")))
	}

	cashHoldings, err := financeProvider.GetCashHoldingsByUserID(ctx, userID)
	if err == nil && len(cashHoldings) > 0 {
		currencyBreakdown := make(map[string]float64)
		for _, holding := range cashHoldings {
			amount := float64(holding.AmountMinorUnits) / 100.0
			currencyBreakdown[string(holding.Currency)] += amount
		}

		var assetDetails []string
		for currency, amount := range currencyBreakdown {
			assetDetails = append(assetDetails, fmt.Sprintf("%.2f %s", amount, currency))
		}
		assets = append(assets, fmt.Sprintf("Наличные (%d счетов): %s", len(cashHoldings), strings.Join(assetDetails, ", ")))
	}

	currencyRates := make(map[string]float64)
	rates, err := currencyService.GetCurrencyRates(ctx)
	if err == nil {
		for _, rate := range rates {
			currencyRates[string(rate.Currency)] = float64(rate.RateMinorUnits) / 100.0
		}
	}

	span.SetAttributes(
		attribute.Int(tracing.AssetsCount, len(assets)),
		attribute.Int(tracing.CurrencyRatesCount, len(currencyRates)),
	)

	assetCounts := make(map[string]int)
	for _, asset := range assets {
		if strings.Contains(asset, "Депозиты") {
			assetCounts["deposits"]++
		} else if strings.Contains(asset, "Брокерские") {
			assetCounts["brokerage"]++
		} else if strings.Contains(asset, "Накопительные") {
			assetCounts["savings"]++
		} else if strings.Contains(asset, "Наличные") {
			assetCounts["cash"]++
		}
	}

	for assetType, count := range assetCounts {
		switch assetType {
		case "deposits":
			span.SetAttributes(attribute.Int(tracing.AssetsDepositsCount, count))
		case "brokerage":
			span.SetAttributes(attribute.Int(tracing.AssetsBrokerageCount, count))
		case "savings":
			span.SetAttributes(attribute.Int(tracing.AssetsSavingsCount, count))
		case "cash":
			span.SetAttributes(attribute.Int(tracing.AssetsCashCount, count))
		}
	}

	span.SetStatus(codes.Ok, "Financial context built successfully")

	return &FinancialContext{
		TotalBalanceRUB: fmt.Sprintf("%.2f", totalBalance),
		Assets:          assets,
		CurrencyRates:   currencyRates,
	}, nil
}
