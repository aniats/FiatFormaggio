package models

import (
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

type CreateSavingAccountRequest struct {
	UserID              domain.UserID
	Name                string
	AmountRUB           float64
	Currency            string
	InterestRatePercent *float64
	ExpirationDate      *time.Time
}
