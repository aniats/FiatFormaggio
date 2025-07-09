package models

import (
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

type CreateDepositRequest struct {
	UserID              domain.UserID
	Name                string
	AmountRUB           float64
	InterestRatePercent *float64
	ExpirationDate      *time.Time
	Currency            string
}
