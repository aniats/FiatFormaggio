package app

import (
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

const DefaultMinorUnits = 100.0

func FormatAmount(amount float64, currency string) string {
	switch currency {
	case "RUB":
		return fmt.Sprintf("%.2f ₽", amount)
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

func FormatAccountType(accountType domain.BrokerageType) string {
	switch accountType {
	case domain.Regular:
		return "Обычный"
	case domain.IIS:
		return "ИИС"
	case domain.IIS3:
		return "ИИС-3"
	case domain.IRA:
		return "ИРА"
	case domain.Margin:
		return "Маржинальный"
	default:
		return string(accountType)
	}
}

/*
TRACING WRAPPER EXAMPLES:

// Before refactoring (boilerplate code):
func (s *SomeService) DoSomething(ctx context.Context, userID int64, data string) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "SomeService.DoSomething")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("data", data),
	)

	// actual business logic here
	return nil
}

// After refactoring (clean and simple):
func (s *SomeService) DoSomething(ctx context.Context, userID int64, data string) error {
	return WithTrace(ctx, "SomeService.DoSomething", func(ctx context.Context) error {
		// actual business logic here
		return nil
	}, TraceOptions{
		Attributes: []attribute.KeyValue{
			TraceAttribute("user.id", userID),
			TraceAttribute("data", data),
		},
	})
}

// For functions with return values:
func (s *SomeService) GetSomething(ctx context.Context, id int64) (string, error) {
	return WithTraceFunc(ctx, "SomeService.GetSomething", func(ctx context.Context) (string, error) {
		// actual business logic here
		return "result", nil
	}, TraceOptions{
		Attributes: []attribute.KeyValue{
			TraceAttribute("id", id),
		},
	})
}

// For functions with no error return:
func (s *SomeService) LogSomething(ctx context.Context, message string) {
	WithTraceNoError(ctx, "SomeService.LogSomething", func(ctx context.Context) {
		// actual business logic here
	}, TraceOptions{
		Attributes: []attribute.KeyValue{
			TraceAttribute("message", message),
		},
	})
}
*/