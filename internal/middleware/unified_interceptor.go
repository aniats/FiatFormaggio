package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UnifiedInterceptor struct {
	tracer      trace.Tracer
	middlewares []Middleware
	config      *InterceptorConfig
}

type InterceptorConfig struct {
	EnableTracing      bool
	EnableMetrics      bool
	EnableProfiling    bool
	EnableSQLProfiling bool
	ServiceName        string
	ProfileSQLQueries  bool
	ProfileThreshold   time.Duration
}

type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, input interface{}) (interface{}, error)

func DefaultConfig(serviceName string) *InterceptorConfig {
	return &InterceptorConfig{
		EnableTracing:      true,
		EnableMetrics:      true,
		EnableProfiling:    true,
		EnableSQLProfiling: true,
		ServiceName:        serviceName,
		ProfileSQLQueries:  true,
		ProfileThreshold:   50 * time.Millisecond,
	}
}

func NewUnifiedInterceptor(config *InterceptorConfig) *UnifiedInterceptor {
	if config == nil {
		config = DefaultConfig(domain.AppName)
	}

	ui := &UnifiedInterceptor{
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

func (ui *UnifiedInterceptor) Use(middleware Middleware) *UnifiedInterceptor {
	ui.middlewares = append(ui.middlewares, middleware)
	return ui
}

func (ui *UnifiedInterceptor) InterceptFunc(fn interface{}, operationName ...string) interface{} {
	fnValue := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() != reflect.Func {
		panic("InterceptFunc expects a function")
	}

	opName := ui.getOperationName(fn, operationName...)

	return reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		ctx := ui.extractContext(args)

		handler := func(ctx context.Context, input interface{}) (interface{}, error) {
			if ctx != nil {
				ui.updateContextInArgs(args, ctx)
			}

			results := fnValue.Call(args)

			var result interface{}
			var err error

			if len(results) > 0 {
				result = results[0].Interface()
			}
			if len(results) > 1 {
				if errValue := results[len(results)-1]; !errValue.IsNil() {
					err = errValue.Interface().(error)
				}
			}

			return result, err
		}

		wrappedHandler := ui.Chain(handler, opName)

		result, err := wrappedHandler(ctx, args)

		returnValues := make([]reflect.Value, fnType.NumOut())

		if fnType.NumOut() > 0 {
			if result != nil {
				returnValues[0] = reflect.ValueOf(result)
			} else {
				returnValues[0] = reflect.Zero(fnType.Out(0))
			}
		}

		if fnType.NumOut() > 1 {
			if err != nil {
				returnValues[fnType.NumOut()-1] = reflect.ValueOf(err)
			} else {
				returnValues[fnType.NumOut()-1] = reflect.Zero(fnType.Out(fnType.NumOut() - 1))
			}
		}

		return returnValues
	}).Interface()
}

func (ui *UnifiedInterceptor) InterceptService(service interface{}, serviceName ...string) interface{} {
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service)

	if serviceType.Kind() == reflect.Ptr {
		serviceType = serviceType.Elem()
		serviceValue = serviceValue.Elem()
	}

	name := ui.getServiceName(serviceType, serviceName...)

	wrapper := &ServiceWrapper{
		service:     service,
		interceptor: ui,
		serviceName: name,
		methodCache: make(map[string]reflect.Value),
	}

	wrapper.wrapAllMethods()

	return wrapper
}

type ServiceWrapper struct {
	service     interface{}
	interceptor *UnifiedInterceptor
	serviceName string
	methodCache map[string]reflect.Value
}

func (sw *ServiceWrapper) wrapAllMethods() {
	serviceValue := reflect.ValueOf(sw.service)
	serviceType := reflect.TypeOf(sw.service)

	if serviceType.Kind() == reflect.Ptr {
		serviceType = serviceType.Elem()
		serviceValue = serviceValue.Elem()
	}

	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		methodName := fmt.Sprintf("%s.%s", sw.serviceName, method.Name)

		// Get the original method
		originalMethod := serviceValue.Method(i)

		// Create wrapped version
		wrappedMethod := sw.interceptor.InterceptFunc(originalMethod.Interface(), methodName)
		sw.methodCache[method.Name] = reflect.ValueOf(wrappedMethod)
	}
}

// CallMethod calls a wrapped method by name
func (sw *ServiceWrapper) CallMethod(methodName string, args ...interface{}) ([]interface{}, error) {
	if wrappedMethod, exists := sw.methodCache[methodName]; exists {
		// Convert args to reflect.Value
		argValues := make([]reflect.Value, len(args))
		for i, arg := range args {
			argValues[i] = reflect.ValueOf(arg)
		}

		// Call the wrapped method
		results := wrappedMethod.Call(argValues)

		// Convert results back to interface{}
		resultInterfaces := make([]interface{}, len(results))
		for i, result := range results {
			resultInterfaces[i] = result.Interface()
		}

		// Check if last result is an error
		if len(results) > 0 {
			if err, ok := results[len(results)-1].Interface().(error); ok {
				return resultInterfaces[:len(resultInterfaces)-1], err
			}
		}

		return resultInterfaces, nil
	}

	return nil, errors.NewTechnicalError(errors.CodeUnknownHandler, fmt.Sprintf("method %s not found", methodName))
}

func (sw *ServiceWrapper) GetWrappedMethod(methodName string) (interface{}, bool) {
	if wrappedMethod, exists := sw.methodCache[methodName]; exists {
		return wrappedMethod.Interface(), true
	}
	return nil, false
}

func (ui *UnifiedInterceptor) InterceptSQL(db interface{}) interface{} {
	if !ui.config.EnableSQLProfiling {
		return db
	}

	if sqlDB, ok := db.(*sql.DB); ok {
		return ProfileDatabase(sqlDB, ui.config.EnableProfiling, ui.config.EnableTracing, ui.config.ProfileThreshold)
	}

	return db
}

func (ui *UnifiedInterceptor) tracingMiddleware(next Handler) Handler {
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

func (ui *UnifiedInterceptor) metricsMiddleware(next Handler) Handler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		start := time.Now()

		result, err := next(ctx, input)

		duration := time.Since(start)
		opName := ui.extractOperationName(ctx)

		metrics.RecordRequest(opName, duration)

		return result, err
	}
}

func (ui *UnifiedInterceptor) profilingMiddleware(next Handler) Handler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		if !ui.config.EnableProfiling {
			return next(ctx, input)
		}

		start := time.Now()
		opName := ui.extractOperationName(ctx)

		// Create profiling labels with operation name
		labels := pprof.Labels("operation", opName)
		ctx = pprof.WithLabels(ctx, labels)

		var result interface{}
		var err error

		// Execute the operation with profiling labels
		pprof.Do(ctx, labels, func(labeledCtx context.Context) {
			result, err = next(labeledCtx, input)
		})

		// Record if this operation took longer than threshold
		duration := time.Since(start)
		if duration > ui.config.ProfileThreshold {
			// Additional profiling for slow operations can be added here
		}

		return result, err
	}
}

func (ui *UnifiedInterceptor) Chain(handler Handler, operationName string) Handler {
	// Create a handler that has the operation name available in context
	finalHandler := func(ctx context.Context, input interface{}) (interface{}, error) {
		// Set operation name in context before applying any middleware
		ctx = context.WithValue(ctx, "operation_name", operationName)

		// Create the actual handler chain
		currentHandler := handler

		// Apply middleware in normal order (first added, first executed)
		for _, middleware := range ui.middlewares {
			currentHandler = middleware(currentHandler)
		}

		return currentHandler(ctx, input)
	}

	return finalHandler
}

func (ui *UnifiedInterceptor) extractContext(args []reflect.Value) context.Context {
	for _, arg := range args {
		if ctx, ok := arg.Interface().(context.Context); ok {
			return ctx
		}
	}
	return context.Background()
}

func (ui *UnifiedInterceptor) updateContextInArgs(args []reflect.Value, ctx context.Context) {
	for i, arg := range args {
		if _, ok := arg.Interface().(context.Context); ok {
			args[i] = reflect.ValueOf(ctx)
			break
		}
	}
}

func (ui *UnifiedInterceptor) getOperationName(fn interface{}, custom ...string) string {
	if len(custom) > 0 && custom[0] != "" {
		return custom[0]
	}

	fnValue := reflect.ValueOf(fn)
	name := runtime.FuncForPC(fnValue.Pointer()).Name()

	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown_operation"
}

func (ui *UnifiedInterceptor) getServiceName(serviceType reflect.Type, custom ...string) string {
	if len(custom) > 0 && custom[0] != "" {
		return custom[0]
	}
	return serviceType.Name()
}

func (ui *UnifiedInterceptor) extractOperationName(ctx context.Context) string {
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

func (ui *UnifiedInterceptor) extractAttributes(input interface{}) []attribute.KeyValue {
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

func (ui *UnifiedInterceptor) findField(value reflect.Value, names ...string) reflect.Value {
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

func (ui *UnifiedInterceptor) InterceptBotHandler(handler func(context.Context, interface{}) error, operationName string) func(context.Context, interface{}) error {
	return ui.InterceptFunc(handler, operationName).(func(context.Context, interface{}) error)
}

func (ui *UnifiedInterceptor) InterceptServiceMethod(method interface{}, operationName string) interface{} {
	return ui.InterceptFunc(method, operationName)
}

func (ui *UnifiedInterceptor) InterceptRepository(repo interface{}) interface{} {
	return ui.InterceptService(repo, "Repository")
}
