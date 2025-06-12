package domain

import "time"

type Deposit struct {
	User                    User
	Name                    string
	AmountMinorUnits        int64
	InterestRateBasisPoints int64
	ExpirationDate          time.Time
	CurrencyName            CurrencyName
}
