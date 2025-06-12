package domain

type AccountType string

const (
	//Deposit   AccountType = "deposit"
	Saving    AccountType = "saving"
	Brokerage AccountType = "brokerage"
	Cash      AccountType = "cash"
)
