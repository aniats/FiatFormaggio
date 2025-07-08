package errors

const (
	CodeDatabaseError    = "DB_ERROR"
	CodeExternalAPIError = "EXT_API_ERROR"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeValidationError  = "ValidATION_ERROR"
	CodeRepositoryError  = "REPOSITORY_ERROR"
	CodeServiceError     = "SERVICE_ERROR"
	CodeConfigError      = "CONFIG_ERROR"
	CodeUnknownHandler   = "UNKNOWN_HANDLER"
	CodeUnknownStep      = "UNKNOWN_STEP"
	CodeParseError       = "PARSE_ERROR"

	CodeInValidAmount   = "INValid_AMOUNT"
	CodeInValidCurrency = "INValid_CURRENCY"
	CodeInValidName     = "INValid_NAME"
	CodeInValidInput    = "INValid_INPUT"
	CodeAmountTooLarge  = "AMOUNT_TOO_LARGE"
	CodeAmountTooSmall  = "AMOUNT_TOO_SMALL"
	CodeNameTooLong     = "NAME_TOO_LONG"
	CodeNameEmpty       = "NAME_EMPTY"
	CodeDateInPast      = "DATE_IN_PAST"
	CodeDateTooFar      = "DATE_TOO_FAR"
	CodeRateOutOfRange  = "RATE_OUT_OF_RANGE"
	CodeUnsupportedType = "UNSUPPORTED_TYPE"
)

var (
	ErrInValidCurrency = NewBusinessError(
		CodeInValidCurrency,
		"❌ Неподдерживаемая валюта. Доступные: RUB, USD, EUR, GBP, JPY, CNY, RSD, XBT, KZT.",
	)

	ErrInValidInput = NewBusinessError(
		CodeInValidInput,
		"❌ Неверный формат ввода. Попробуйте еще раз.",
	)

	ErrCurrencyRatesUnavailable = NewBusinessError(
		CodeExternalAPIError,
		"❌ Курсы валют временно недоступны. Попробуйте позже.",
	)

	ErrNameEmpty = NewBusinessError(
		CodeNameEmpty,
		"❌ Название не может быть пустым.",
	)

	ErrNameTooLong = NewBusinessError(
		CodeNameTooLong,
		"❌ Название слишком длинное (максимум 255 символов).",
	)

	ErrAmountNotPositive = NewBusinessError(
		CodeInValidAmount,
		"❌ Сумма должна быть положительной.",
	)

	ErrAmountTooLarge = NewBusinessError(
		CodeAmountTooLarge,
		"❌ Слишком большая сумма (максимум 1,000,000,000).",
	)

	ErrAmountTooSmall = NewBusinessError(
		CodeAmountTooSmall,
		"❌ Минимальная сумма: 1.",
	)

	ErrRateOutOfRange = NewBusinessError(
		CodeRateOutOfRange,
		"❌ Процентная ставка должна быть от 0 до 50%.",
	)

	ErrDateInPast = NewBusinessError(
		CodeDateInPast,
		"❌ Дата не может быть в прошлом.",
	)

	ErrDateTooFar = NewBusinessError(
		CodeDateTooFar,
		"❌ Слишком далекая дата (максимум 10 лет).",
	)

	ErrInValidDateFormat = NewBusinessError(
		CodeParseError,
		"❌ Неверный формат даты.",
	)

	ErrUnknownStep = NewBusinessError(
		CodeUnknownStep,
		"❌ Неизвестный шаг операции.",
	)

	ErrUnsupportedAccountType = NewBusinessError(
		CodeUnsupportedType,
		"❌ Неподдерживаемый тип счета.",
	)

	ErrUnknownSessionHandler = NewBusinessError(
		CodeUnknownHandler,
		"❌ Неизвестный тип сессии.",
	)

	ErrRequestNil = NewBusinessError(
		CodeInValidInput,
		"❌ Запрос не может быть пустым.",
	)

	ErrInValidUserID = NewBusinessError(
		CodeInValidInput,
		"❌ ID пользователя должен быть положительным.",
	)

	ErrNameRequired = NewBusinessError(
		CodeNameEmpty,
		"❌ Название обязательно.",
	)

	ErrAmountNegative = NewBusinessError(
		CodeInValidAmount,
		"❌ Сумма не может быть отрицательной.",
	)

	ErrCurrencyRequired = NewBusinessError(
		CodeInValidCurrency,
		"❌ Валюта обязательна.",
	)

	ErrAccountTypeRequired = NewBusinessError(
		CodeInValidInput,
		"❌ Тип счета обязателен.",
	)
)

func NewCurrencyError(currency string) *AppError {
	return ErrInValidCurrency.WithContext("currency", currency)
}

func NewUnknownSessionHandlerError(sessionType string) *AppError {
	return ErrUnknownSessionHandler.WithContext("sessionType", sessionType)
}

func NewUnsupportedCurrencyError(currency string) *AppError {
	return NewBusinessError(
		CodeInValidCurrency,
		"❌ Неподдерживаемая валюта '"+currency+"'. Поддерживаемые валюты: RUB, USD, EUR, CNY, GBP.",
	).WithContext("currency", currency)
}

func NewCurrencyNotAllowedForDepositsError(currency string) *AppError {
	return NewBusinessError(
		CodeInValidCurrency,
		"❌ Валюта '"+currency+"' не поддерживается для депозитов.",
	).WithContext("currency", currency)
}

func NewUnsupportedAccountTypeError(accountType string) *AppError {
	return NewBusinessError(
		CodeUnsupportedType,
		"❌ Неподдерживаемый тип счета: "+accountType+". Поддерживаемые типы: regular, iis, iis3, ira, margin.",
	).WithContext("accountType", accountType)
}
