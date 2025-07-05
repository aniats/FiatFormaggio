package domain

import "time"

type Deposit struct {
	Id                      int64
	UserId                  UserId
	Name                    string
	AmountMinorUnits        int64
	InterestRateBasisPoints int64
	ExpirationDate          *time.Time
	Currency                CurrencyName
}
