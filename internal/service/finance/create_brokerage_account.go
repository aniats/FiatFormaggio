package finance

import (
	"context"
	"fmt"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

func (s *FinanceService) CreateBrokerageAccount(ctx context.Context, req *models.CreateBrokerageAccountRequest) (*domain.BrokerageAccount, error) {
	if err := s.validateCreateBrokerageAccountRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	currency, err := s.validateAndNormalizeCurrency(req.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	accountType, err := s.parseBrokerageType(req.AccountType)
	if err != nil {
		return nil, fmt.Errorf("invalid account type: %w", err)
	}

	amountMinorUnits := int64(req.AmountRUB * 100)

	account := &domain.BrokerageAccount{
		UserId:           req.UserID,
		Name:             strings.TrimSpace(req.Name),
		AmountMinorUnits: amountMinorUnits,
		Currency:         currency,
		Broker:           req.Broker,
		AccountType:      accountType,
	}

	if err := s.repo.CreateBrokerageAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create brokerage account: %w", err)
	}

	return account, nil
}

func (s *FinanceService) validateCreateBrokerageAccountRequest(req *models.CreateBrokerageAccountRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if req.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return fmt.Errorf("name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("name too long (max 255 characters)")
	}

	if req.AmountRUB < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	if req.AmountRUB > 1000000000 {
		return fmt.Errorf("amount too large (max 1,000,000,000)")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if req.AccountType == "" {
		return fmt.Errorf("account type is required")
	}

	return nil
}

func (s *FinanceService) parseBrokerageType(accountTypeStr string) (domain.BrokerageType, error) {
	accountTypeStr = strings.TrimSpace(strings.ToLower(accountTypeStr))

	switch accountTypeStr {
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
		return "", fmt.Errorf("unsupported account type: %s. Supported types: regular, iis, iis3, ira, margin", accountTypeStr)
	}
}
