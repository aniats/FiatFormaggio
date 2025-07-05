package models

import (
	"github.com/aniats/FiatFormaggio/internal/domain"
	"time"
)

type CreateDepositRequest struct {
	UserID              domain.UserId
	Name                string
	AmountRUB           float64
	InterestRatePercent *float64
	ExpirationDate      *time.Time
	Currency            string
}
