package domain

type CommandType string

const (
	CommandStart                  CommandType = "start"
	CommandTotal                  CommandType = "total"
	CommandDeposits               CommandType = "deposits"
	CommandCreateDeposit          CommandType = "create_deposit"
	CommandBrokerageAccounts      CommandType = "brokerage_accounts"
	CommandCreateBrokerageAccount CommandType = "create_brokerage_account"
	CommandSavingAccounts         CommandType = "saving_accounts"
	CommandCreateSavingAccount    CommandType = "create_saving_account"
	CommandCashHoldings           CommandType = "cash_holdings"
	CommandCreateCashHolding      CommandType = "create_cash_holding"
	CommandRates                  CommandType = "rates"
	CommandFinancialAdvice        CommandType = "financial_advice"
)

const (
	CallbackConfirmDepositYes   = "confirm_deposit_yes"
	CallbackConfirmDepositNo    = "confirm_deposit_no"
	CallbackConfirmBrokerageYes = "confirm_brokerage_yes"
	CallbackConfirmBrokerageNo  = "confirm_brokerage_no"
	CallbackConfirmSavingYes    = "confirm_saving_yes"
	CallbackConfirmSavingNo     = "confirm_saving_no"
	CallbackConfirmCashYes      = "confirm_cash_yes"
	CallbackConfirmCashNo       = "confirm_cash_no"

	CallbackCurrencyPrefix = "currency_"
	CallbackCurrencyRUB    = "currency_RUB"
	CallbackCurrencyUSD    = "currency_USD"
	CallbackCurrencyEUR    = "currency_EUR"
	CallbackCurrencyGBP    = "currency_GBP"
	CallbackCurrencyJPY    = "currency_JPY"
	CallbackCurrencyCNY    = "currency_CNY"
	CallbackCurrencyRSD    = "currency_RSD"
	CallbackCurrencyXBT    = "currency_XBT"
	CallbackCurrencyKZT    = "currency_KZT"
	CallbackCurrencySkip   = "currency_skip"

	CallbackMenuTotal           = "menu_total"
	CallbackMenuDeposits        = "menu_deposits"
	CallbackMenuCreateDeposit   = "menu_create_deposit"
	CallbackMenuBrokerage       = "menu_brokerage"
	CallbackMenuCreateBrokerage = "menu_create_brokerage"
	CallbackMenuSaving          = "menu_saving"
	CallbackMenuCreateSaving    = "menu_create_saving"
	CallbackMenuCash            = "menu_cash"
	CallbackMenuCreateCash      = "menu_create_cash"
	CallbackMenuRates           = "menu_rates"
	CallbackMenuFinancialAdvice = "menu_financial_advice"
	CallbackStopFinancialAdvice = "stop_financial_advice"
)
