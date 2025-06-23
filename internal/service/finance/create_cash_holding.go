package finance

import (
	"context"
	"fmt"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

func (s *FinanceService) CreateCashHolding(ctx context.Context, req *models.CreateCashHoldingRequest) (*domain.CashHolding, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err := s.validateCreateCashHoldingRequest(req); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		currency, err := s.validateAndNormalizeCurrency(req.Currency)
		if err != nil {
			return nil, fmt.Errorf("invalid currency: %w", err)
		}

		amountMinorUnits := int64(req.AmountRUB * 100)

		cashHolding := &domain.CashHolding{
			UserId:           req.UserID,
			Name:             strings.TrimSpace(req.Name),
			AmountMinorUnits: amountMinorUnits,
			Currency:         currency,
		}

		if err := s.repo.CreateCashHolding(ctx, cashHolding); err != nil {
			return nil, fmt.Errorf("failed to create cash holding: %w", err)
		}

		return cashHolding, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.CreateCashHolding")
	resultInterface, err := wrappedHandler(ctx, req)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.(*domain.CashHolding), nil
	}
	return nil, nil
}

func (s *FinanceService) validateCreateCashHoldingRequest(req *models.CreateCashHoldingRequest) error {
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

	return nil
}
