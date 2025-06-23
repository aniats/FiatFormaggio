package finance

import (
	"context"
	"fmt"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

func (s *FinanceService) CreateSavingAccount(ctx context.Context, req *models.CreateSavingAccountRequest) (*domain.SavingAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err := s.validateCreateSavingAccountRequest(req); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		currency, err := s.validateAndNormalizeCurrency(req.Currency)
		if err != nil {
			return nil, fmt.Errorf("invalid currency: %w", err)
		}

		amountMinorUnits := int64(req.AmountRUB * 100)

		var interestRateBasisPoints int64
		if req.InterestRatePercent != nil {
			interestRateBasisPoints = int64(*req.InterestRatePercent * 100)
		}

		account := &domain.SavingAccount{
			UserId:                  req.UserID,
			Name:                    strings.TrimSpace(req.Name),
			AmountMinorUnits:        amountMinorUnits,
			Currency:                currency,
			InterestRateBasisPoints: interestRateBasisPoints,
			ExpirationDate:          req.ExpirationDate,
		}

		if err := s.repo.CreateSavingAccount(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to create saving account: %w", err)
		}

		return account, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.CreateSavingAccount")
	resultInterface, err := wrappedHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.(*domain.SavingAccount), nil
	}
	return nil, nil
}

func (s *FinanceService) validateCreateSavingAccountRequest(req *models.CreateSavingAccountRequest) error {
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

	if req.InterestRatePercent != nil {
		if *req.InterestRatePercent < 0 || *req.InterestRatePercent > 100 {
			return fmt.Errorf("interest rate must be between 0 and 100 percent")
		}

		if *req.InterestRatePercent > 50 {
			return fmt.Errorf("interest rate seems too high (max 50%% for safety)")
		}
	}

	return nil
}
