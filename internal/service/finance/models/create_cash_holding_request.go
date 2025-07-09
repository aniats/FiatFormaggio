package models

import (
	"github.com/aniats/FiatFormaggio/internal/domain"
)

type CreateCashHoldingRequest struct {
	UserID    domain.UserID
	Name      string
	AmountRUB float64
	Currency  string
}
