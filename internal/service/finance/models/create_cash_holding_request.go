package models

import (
	"github.com/aniats/FiatFormaggio/internal/domain"
)

type CreateCashHoldingRequest struct {
	UserID    domain.UserId
	Name      string
	AmountRUB float64
	Currency  string
}