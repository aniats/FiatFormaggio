package domain

import "time"

type Deposit struct {
	ID                      int64
	UserID                  UserID
	Name                    string
	AmountMinorUnits        int64
	InterestRateBasisPoints int64
	ExpirationDate          *time.Time
	Currency                CurrencyName
}
