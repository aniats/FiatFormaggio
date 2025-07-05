package finance

import (
	"context"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

func (s *FinanceService) CreateBrokerageAccount(ctx context.Context, req *models.CreateBrokerageAccountRequest) (*domain.BrokerageAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err := s.validateCreateBrokerageAccountRequest(req); err != nil {
			return nil, errors.WrapValidationError(err)
		}

		currency, err := s.validateAndNormalizeCurrency(req.Currency)
		if err != nil {
			return nil, errors.WrapValidationError(err)
		}

		accountType, err := s.parseBrokerageType(req.AccountType)
		if err != nil {
			return nil, errors.WrapValidationError(err)
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
			return nil, errors.WrapRepositoryError(err)
		}

		return account, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.CreateBrokerageAccount")
	resultInterface, err := wrappedHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.(*domain.BrokerageAccount), nil
	}
	return nil, nil
}

func (s *FinanceService) validateCreateBrokerageAccountRequest(req *models.CreateBrokerageAccountRequest) error {
	if req == nil {
		return errors.ErrRequestNil
	}

	if req.UserID <= 0 {
		return errors.ErrInvalidUserID
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.ErrNameRequired
	}

	if len(name) > 255 {
		return errors.ErrNameTooLong
	}

	if req.AmountRUB < 0 {
		return errors.ErrAmountNegative
	}

	if req.AmountRUB > 1000000000 {
		return errors.ErrAmountTooLarge
	}

	if req.Currency == "" {
		return errors.ErrCurrencyRequired
	}

	if req.AccountType == "" {
		return errors.ErrAccountTypeRequired
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
		return "", errors.NewUnsupportedAccountTypeError(accountTypeStr)
	}
}
