package tracing

const (
	UserID = "userID"
)

const (
	ChatID = "chat.id"
)

const (
	Command = "command"
)

const (
	SessionID      = "session.id"
	SessionType    = "session.type"
	SessionActive  = "session.active"
	SessionStopped = "session.stopped"
	SessionCreated = "session.created"
)

const (
	MessageText = "message.text"
	UserMessage = "user.message"
)

const (
	TotalBalance       = "total.balance"
	AssetsCount        = "assets.count"
	CurrencyRatesCount = "currency.rates.count"

	AssetsDepositsCount  = "assets.deposits.count"
	AssetsBrokerageCount = "assets.brokerage.count"
	AssetsSavingsCount   = "assets.savings.count"
	AssetsCashCount      = "assets.cash.count"
)

const (
	OpenAIModel         = "openai.model"
	OpenAIMessagesTotal = "openai.messages.total"
	OpenAIMaxTokens     = "openai.max_tokens"
	OpenAITemperature   = "openai.temperature"

	OpenAIResponseID      = "openai.response.id"
	OpenAIResponseModel   = "openai.response.model"
	OpenAIResponseChoices = "openai.response.choices"

	OpenAITokensPrompt     = "openai.tokens.prompt"
	OpenAITokensCompletion = "openai.tokens.completion"
	OpenAITokensTotal      = "openai.tokens.total"
)

const (
	ResponseLength        = "response.length"
	InitialPrompt         = "initial.prompt"
	InitialResponseLength = "initial.response.length"

	PreviousMessagesCount = "previous.messages.count"
	FinancialAssetsCount  = "financial.assets.count"
)

const (
	FinancialAdviceCommand = "financial_advice"
)

const (
	ParamUserID      = "user_id"
	ParamChatID      = "chat_id"
	ParamSessionStep = "session_step"
	ParamSessionType = "session_type"
)
