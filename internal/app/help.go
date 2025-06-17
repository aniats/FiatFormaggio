package app

func (b *Bot) sendHelp(chatID int64) {
	helpText := `Доступные команды:
		/total - Общий баланс по всем счетам в рублях с курсами валют
		/deposits - Показать депозиты
		/create_deposit - Создать депозит
		/brokerage_accounts - Показать брокерские счета
		/create_brokerage_account - Создать брокерский счет
		/saving_accounts - Показать накопительные счета
		/create_saving_account - Создать накопительный счет
		/cash_holdings - Показать наличные счета
		/create_cash_holding - Создать наличный счет
		/rates - Курсы валют ЦБ РФ
	`

	b.sendMessage(chatID, helpText)
}
