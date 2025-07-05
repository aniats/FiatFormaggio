package models

import (
	"github.com/aniats/FiatFormaggio/internal/domain"
	"time"
)

type CreateSavingAccountRequest struct {
	UserID              domain.UserId
	Name                string
	AmountRUB           float64
	Currency            string
	InterestRatePercent *float64
	ExpirationDate      *time.Time
}