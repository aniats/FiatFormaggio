package errors

import (
	"context"
	"log"
)

type ErrorHandler struct {
	logTechnicalErrors bool
	logBusinessErrors  bool
}

func DefaultErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		logTechnicalErrors: true,
		logBusinessErrors:  false,
	}
}

func (h *ErrorHandler) Handle(ctx context.Context, err error, operation string) string {
	if err == nil {
		return ""
	}

	if appErr, ok := AsAppError(err); ok {
		return h.handleAppError(ctx, appErr, operation)
	}

	return h.handleGenericError(ctx, err, operation)
}

func (h *ErrorHandler) handleAppError(ctx context.Context, appErr *AppError, operation string) string {
	if appErr.Context["operation"] == nil {
		appErr.WithContext("operation", operation)
	}

	if appErr.IsTechnical() && h.logTechnicalErrors {
		h.logTechnicalError(ctx, appErr, operation)
	} else if appErr.IsBusiness() && h.logBusinessErrors {
		h.logBusinessError(ctx, appErr, operation)
	}

	return appErr.UserMessage()
}

func (h *ErrorHandler) handleGenericError(ctx context.Context, err error, operation string) string {
	if h.logTechnicalErrors {
		log.Printf("[TECHNICAL_ERROR] Operation: %s, Error: %v", operation, err)
	}

	return "❌ Произошла ошибка. Попробуйте позже."
}

func (h *ErrorHandler) logTechnicalError(ctx context.Context, appErr *AppError, operation string) {
	logMsg := "[TECHNICAL_ERROR]"
	if appErr.Code != "" {
		logMsg += " Code: " + appErr.Code
	}
	if operation != "" {
		logMsg += " Operation: " + operation
	}
	if appErr.TechnicalMsg != "" {
		logMsg += " Message: " + appErr.TechnicalMsg
	}
	if appErr.Cause != nil {
		logMsg += " Cause: " + appErr.Cause.Error()
	}
	if len(appErr.Context) > 0 {
		logMsg += " Context: %+v"
		log.Printf(logMsg, appErr.Context)
	} else {
		log.Print(logMsg)
	}
}

func (h *ErrorHandler) logBusinessError(ctx context.Context, appErr *AppError, operation string) {
	log.Printf("[BUSINESS_ERROR] Code: %s, Operation: %s, User: %v",
		appErr.Code, operation, appErr.Context["user_id"])
}

