package models

import (
	"github.com/aniats/FiatFormaggio/internal/domain"
)

type CreateBrokerageAccountRequest struct {
	UserID      domain.UserId
	Name        string
	AmountRUB   float64
	Currency    string
	Broker      *string
	AccountType string
}
