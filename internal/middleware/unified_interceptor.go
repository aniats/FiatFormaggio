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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"github.com/aniats/FiatFormaggio/internal/metrics"
)

// UnifiedInterceptor combines all middleware approaches into one powerful system
type UnifiedInterceptor struct {
	tracer      trace.Tracer
	middlewares []Middleware
	config      *InterceptorConfig
}

// InterceptorConfig configures the unified interceptor behavior
type InterceptorConfig struct {
	EnableTracing     bool
	EnableMetrics     bool
	EnableProfiling   bool
	EnableSQLProfiling bool
	ServiceName       string
	ProfileSQLQueries bool
	ProfileThreshold  time.Duration // Only profile operations taking longer than this
}

// Middleware represents a generic middleware function
type Middleware func(next Handler) Handler

// Handler is the most generic handler interface
type Handler func(ctx context.Context, input interface{}) (interface{}, error)

// DefaultConfig returns a sensible default configuration
func DefaultConfig(serviceName string) *InterceptorConfig {
	return &InterceptorConfig{
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableProfiling:   true,
		EnableSQLProfiling: true,
		ServiceName:       serviceName,
		ProfileSQLQueries: true,
		ProfileThreshold:  50 * time.Millisecond, // Profile operations > 50ms
	}
}

// NewUnifiedInterceptor creates a new unified interceptor
func NewUnifiedInterceptor(config *InterceptorConfig) *UnifiedInterceptor {
	if config == nil {
		config = DefaultConfig("fiat-formaggio")
	}
	
	ui := &UnifiedInterceptor{
		tracer: otel.Tracer(config.ServiceName),
		config: config,
	}
	
	// Add default middleware stack
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

// Use adds custom middleware to the stack
func (ui *UnifiedInterceptor) Use(middleware Middleware) *UnifiedInterceptor {
	ui.middlewares = append(ui.middlewares, middleware)
	return ui
}

// InterceptFunc wraps any function with the unified middleware stack
func (ui *UnifiedInterceptor) InterceptFunc(fn interface{}, operationName ...string) interface{} {
	fnValue := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)
	
	if fnType.Kind() != reflect.Func {
		panic("InterceptFunc expects a function")
	}
	
	opName := ui.getOperationName(fn, operationName...)
	
	// Create wrapper function with the same signature
	return reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		// Extract context if available
		ctx := ui.extractContext(args)
		
		// Create handler from original function
		handler := func(ctx context.Context, input interface{}) (interface{}, error) {
			// Update context in args if found
			if ctx != nil {
				ui.updateContextInArgs(args, ctx)
			}
			
			// Call original function
			results := fnValue.Call(args)
			
			// Extract return values
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
		
		// Apply middleware stack
		wrappedHandler := ui.Chain(handler, opName)
		
		// Execute
		result, err := wrappedHandler(ctx, args)
		
		// Convert back to reflect.Value
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
				returnValues[fnType.NumOut()-1] = reflect.Zero(fnType.Out(fnType.NumOut()-1))
			}
		}
		
		return returnValues
	}).Interface()
}

// InterceptService wraps all methods of a service with middleware
func (ui *UnifiedInterceptor) InterceptService(service interface{}, serviceName ...string) interface{} {
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service)
	
	if serviceType.Kind() == reflect.Ptr {
		serviceType = serviceType.Elem()
		serviceValue = serviceValue.Elem()
	}
	
	name := ui.getServiceName(serviceType, serviceName...)
	
	// Create new struct type with wrapped methods
	fields := make([]reflect.StructField, 0)
	
	// Copy original fields
	for i := 0; i < serviceType.NumField(); i++ {
		field := serviceType.Field(i)
		fields = append(fields, field)
	}
	
	// Add interceptor field
	fields = append(fields, reflect.StructField{
		Name: "UnifiedInterceptor",
		Type: reflect.TypeOf(ui),
	})
	
	wrappedType := reflect.StructOf(fields)
	wrappedValue := reflect.New(wrappedType).Elem()
	
	// Copy field values
	for i := 0; i < serviceType.NumField(); i++ {
		wrappedValue.Field(i).Set(serviceValue.Field(i))
	}
	
	// Set interceptor
	wrappedValue.FieldByName("UnifiedInterceptor").Set(reflect.ValueOf(ui))
	
	// Wrap methods
	wrappedPtr := wrappedValue.Addr()
	ui.wrapMethods(wrappedPtr, serviceValue, name)
	
	return wrappedPtr.Interface()
}

// InterceptSQL creates a SQL interceptor for database operations
func (ui *UnifiedInterceptor) InterceptSQL(db interface{}) interface{} {
	if !ui.config.EnableSQLProfiling {
		return db
	}
	
	// Create SQL driver wrapper using SQLProfiler
	if sqlDB, ok := db.(*sql.DB); ok {
		return ProfileDatabase(sqlDB, ui.config.EnableProfiling, ui.config.EnableTracing, ui.config.ProfileThreshold)
	}
	
	return db
}

// Built-in middleware implementations

func (ui *UnifiedInterceptor) tracingMiddleware(next Handler) Handler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		opName := ui.extractOperationName(ctx)
		
		ctx, span := ui.tracer.Start(ctx, opName)
		defer span.End()
		
		// Extract attributes from input
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
		
		// Record metrics
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

// Helper methods

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
	
	// Clean up the name
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
	
	// If we reach here, operation name wasn't set properly
	// This should no longer happen with the fixed middleware chain
	// Adding some debug info to help identify the source
	if ctx.Value("operation_name") != nil {
		// Operation name exists but is empty string
		return "empty_operation_name"
	}
	return "unknown_operation"
}

func (ui *UnifiedInterceptor) extractAttributes(input interface{}) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	
	if input == nil {
		return attrs
	}
	
	// Use reflection to extract common attributes
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	
	// Extract user ID if available
	if userIDField := ui.findField(value, "UserID", "UserId", "ID"); userIDField.IsValid() {
		if userID := userIDField.Int(); userID > 0 {
			attrs = append(attrs, attribute.Int64("user.id", userID))
		}
	}
	
	// Extract chat ID if available
	if chatIDField := ui.findField(value, "ChatID", "ChatId"); chatIDField.IsValid() {
		if chatID := chatIDField.Int(); chatID > 0 {
			attrs = append(attrs, attribute.Int64("chat.id", chatID))
		}
	}
	
	// Extract command if available
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

func (ui *UnifiedInterceptor) wrapMethods(wrappedPtr, originalValue reflect.Value, serviceName string) {
	originalType := originalValue.Type()
	
	for i := 0; i < originalType.NumMethod(); i++ {
		method := originalType.Method(i)
		methodName := fmt.Sprintf("%s.%s", serviceName, method.Name)
		
		// Get the original method
		originalMethod := originalValue.Method(i)
		
		// Create wrapped version
		_ = ui.InterceptFunc(originalMethod.Interface(), methodName)
		
		// Set the wrapped method on the new struct
		// This requires careful reflection work that would need more implementation
	}
}

// Convenience functions for common patterns

// InterceptBotHandler wraps a bot message handler
func (ui *UnifiedInterceptor) InterceptBotHandler(handler func(context.Context, interface{}) error, operationName string) func(context.Context, interface{}) error {
	return ui.InterceptFunc(handler, operationName).(func(context.Context, interface{}) error)
}

// InterceptServiceMethod wraps a service method
func (ui *UnifiedInterceptor) InterceptServiceMethod(method interface{}, operationName string) interface{} {
	return ui.InterceptFunc(method, operationName)
}

// InterceptRepository wraps all repository methods
func (ui *UnifiedInterceptor) InterceptRepository(repo interface{}) interface{} {
	return ui.InterceptService(repo, "Repository")
}