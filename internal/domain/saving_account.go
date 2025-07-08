package domain

import "time"

type SavingAccount struct {
	ID                      int64
	UserID                  UserID
	Name                    string
	AmountMinorUnits        int64
	ExpirationDate          *time.Time
	InterestRateBasisPoints int64
	Currency                CurrencyName
}
