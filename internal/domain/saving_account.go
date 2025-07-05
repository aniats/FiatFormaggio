package domain

import "time"

type SavingAccount struct {
	Id                      int64
	UserId                  UserId
	Name                    string
	AmountMinorUnits        int64
	ExpirationDate          *time.Time
	InterestRateBasisPoints int64
	Currency                CurrencyName
}
