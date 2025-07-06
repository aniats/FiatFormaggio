package middleware

import (
	"context"
	"reflect"
	"runtime/pprof"
	"time"

	"github.com/aniats/FiatFormaggio/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TracerInterface interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

type Interceptor struct {
	tracer      TracerInterface
	middlewares []Middleware
	config      *InterceptorConfig
}

type InterceptorConfig struct {
	EnableTracing    bool
	EnableMetrics    bool
	EnableProfiling  bool
	ServiceName      string
	ProfileThreshold time.Duration
}

type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, input interface{}) (interface{}, error)

func DefaultConfig(serviceName string) *InterceptorConfig {
	return &InterceptorConfig{
		EnableTracing:    true,
		EnableMetrics:    true,
		EnableProfiling:  true,
		ServiceName:      serviceName,
		ProfileThreshold: 50 * time.Millisecond,
	}
}

func NewInterceptor(config *InterceptorConfig, appName string) *Interceptor {
	if config == nil {
		config = DefaultConfig(appName)
	}

	ui := &Interceptor{
		tracer: otel.Tracer(config.ServiceName),
		config: config,
	}

	if config.EnableTracing {
		ui.middlewares = append(ui.middlewares, ui.tracingMiddleware)
	}
	if config.EnableMetrics {
		ui.middlewares = append(ui.middlewares, ui.metricsMiddleware)
	}
	if config.EnableProfiling {
		ui.middlewares = append(ui.middlewares, ui.profilingMiddleware)
	}

	return ui
}

type ServiceWrapper struct {
	service     interface{}
	interceptor *Interceptor
	serviceName string
	methodCache map[string]reflect.Value
}

func (ui *Interceptor) tracingMiddleware(next Handler) Handler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		opName := ui.extractOperationName(ctx)

		ctx, span := ui.tracer.Start(ctx, opName)
		defer span.End()

		attrs := ui.extractAttributes(input)
		if len(attrs) > 0 {
			span.SetAttributes(attrs...)
		}

		result, err := next(ctx, input)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return result, err
	}
}

func (ui *Interceptor) metricsMiddleware(next Handler) Handler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		start := time.Now()

		result, err := next(ctx, input)

		duration := time.Since(start)
		opName := ui.extractOperationName(ctx)

		metrics.RecordRequest(opName, duration)

		return result, err
	}
}

func (ui *Interceptor) profilingMiddleware(next Handler) Handler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		if !ui.config.EnableProfiling {
			return next(ctx, input)
		}

		start := time.Now()
		opName := ui.extractOperationName(ctx)

		labels := pprof.Labels("operation", opName)
		ctx = pprof.WithLabels(ctx, labels)

		var (
			result interface{}
			err    error
		)

		pprof.Do(ctx, labels, func(labeledCtx context.Context) {
			result, err = next(labeledCtx, input)
		})

		duration := time.Since(start)
		if duration > ui.config.ProfileThreshold {
			// Additional profiling for slow operations can be added here
		}

		return result, err
	}
}

func (ui *Interceptor) Chain(handler Handler, operationName string) Handler {
	finalHandler := func(ctx context.Context, input interface{}) (interface{}, error) {
		ctx = context.WithValue(ctx, "operation_name", operationName)

		currentHandler := handler

		for _, middleware := range ui.middlewares {
			currentHandler = middleware(currentHandler)
		}

		return currentHandler(ctx, input)
	}

	return finalHandler
}

func (ui *Interceptor) extractOperationName(ctx context.Context) string {
	if opName, ok := ctx.Value("operation_name").(string); ok {
		if opName != "" {
			return opName
		}
	}

	if ctx.Value("operation_name") != nil {
		return "empty_operation_name"
	}
	return "unknown_operation"
}

func (ui *Interceptor) extractAttributes(input interface{}) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	if input == nil {
		return attrs
	}

	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if userIDField := ui.findField(value, "UserID", "UserId", "ID"); userIDField.IsValid() {
		if userID := userIDField.Int(); userID > 0 {
			attrs = append(attrs, attribute.Int64("user.id", userID))
		}
	}

	if chatIDField := ui.findField(value, "ChatID", "ChatId"); chatIDField.IsValid() {
		if chatID := chatIDField.Int(); chatID > 0 {
			attrs = append(attrs, attribute.Int64("chat.id", chatID))
		}
	}

	if cmdField := ui.findField(value, "Command", "Cmd"); cmdField.IsValid() {
		if cmd := cmdField.String(); cmd != "" {
			attrs = append(attrs, attribute.String("command", cmd))
		}
	}

	return attrs
}

func (ui *Interceptor) findField(value reflect.Value, names ...string) reflect.Value {
	if value.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	for _, name := range names {
		if field := value.FieldByName(name); field.IsValid() {
			return field
		}
	}

	return reflect.Value{}
}
