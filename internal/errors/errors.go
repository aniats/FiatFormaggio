package errors

import (
	"fmt"
)

type ErrorType string

const (
	ErrorTypeTechnical ErrorType = "technical"
	ErrorTypeBusiness  ErrorType = "business"
)

type AppError struct {
	Code         string    `json:"code"`
	TechnicalMsg string    `json:"technical_message"`
	Type         ErrorType `json:"type"`

	UserMsg string `json:"user_message"`

	Cause   error                  `json:"-"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.TechnicalMsg, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.TechnicalMsg)
}

func (e *AppError) UserMessage() string {
	if e.UserMsg == "" {
		return "❌ Произошла ошибка. Попробуйте позже."
	}
	return e.UserMsg
}

func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

func (e *AppError) IsTechnical() bool {
	return e.Type == ErrorTypeTechnical
}

func (e *AppError) IsBusiness() bool {
	return e.Type == ErrorTypeBusiness
}

func NewTechnicalError(code, technicalMsg string) *AppError {
	return &AppError{
		Code:         code,
		TechnicalMsg: technicalMsg,
		Type:         ErrorTypeTechnical,
		UserMsg:      "❌ Произошла техническая ошибка. Попробуйте позже.", // Generic user message
		Context:      make(map[string]interface{}),
	}
}

func NewBusinessError(code, userMsg string) *AppError {
	return &AppError{
		Code:         code,
		TechnicalMsg: fmt.Sprintf("Business logic error: %s", code),
		Type:         ErrorTypeBusiness,
		UserMsg:      userMsg,
		Context:      make(map[string]interface{}),
	}
}

func WrapError(cause error, code, technicalMsg, userMsg string) *AppError {
	return &AppError{
		Code:         code,
		TechnicalMsg: technicalMsg,
		Type:         ErrorTypeTechnical,
		UserMsg:      userMsg,
		Cause:        cause,
		Context:      make(map[string]interface{}),
	}
}

func WrapValidationError(cause error) *AppError {
	return NewTechnicalError(CodeValidationError, "validation failed").WithCause(cause)
}

func WrapRepositoryError(cause error) *AppError {
	return NewTechnicalError(CodeRepositoryError, "repository operation failed").WithCause(cause)
}

func WrapServiceError(cause error) *AppError {
	return NewTechnicalError(CodeServiceError, "service operation failed").WithCause(cause)
}

func WrapExternalAPIError(cause error) *AppError {
	return NewTechnicalError(CodeExternalAPIError, "external API call failed").WithCause(cause)
}

func WrapDatabaseError(cause error) *AppError {
	return NewTechnicalError(CodeDatabaseError, "database operation failed").WithCause(cause)
}

func WrapParseError(cause error) *AppError {
	return NewTechnicalError(CodeParseError, "parsing failed").WithCause(cause)
}

func WrapConfigError(cause error) *AppError {
	return NewTechnicalError(CodeConfigError, "configuration error").WithCause(cause)
}

func NewTechnicalValidationError(message string) *AppError {
	return NewTechnicalError(CodeValidationError, message)
}

func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

func GetUserMessage(err error) string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.UserMessage()
	}
	return "❌ Произошла ошибка. Попробуйте позже."
}
