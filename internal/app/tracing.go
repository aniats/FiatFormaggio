package app

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otelTrace "go.opentelemetry.io/otel/trace"
)

func InitTracing() (func(), error) {
	ctx := context.Background()
	
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("fiat-formaggio"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	var exporter trace.SpanExporter
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	
	if jaegerEndpoint != "" {
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(jaegerEndpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			log.Printf("Failed to create Jaeger exporter, falling back to stdout: %v", err)
			exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
			if err != nil {
				return nil, err
			}
		} else {
			log.Printf("Using Jaeger exporter at %s", jaegerEndpoint)
		}
	} else {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		log.Println("Using stdout exporter (set JAEGER_ENDPOINT to use Jaeger)")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	log.Println("Tracing initialized successfully")

	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}

	return cleanup, nil
}

const tracerName = "fiat-formaggio"

// TraceOptions holds optional attributes for tracing
type TraceOptions struct {
	Attributes []attribute.KeyValue
}

// TraceAttribute creates a key-value attribute for tracing
func TraceAttribute(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, "unknown")
	}
}

// WithTrace wraps a function with OpenTelemetry tracing
func WithTrace(ctx context.Context, operationName string, fn func(ctx context.Context) error, opts ...TraceOptions) error {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, operationName)
	defer span.End()

	// Add attributes if provided
	if len(opts) > 0 && len(opts[0].Attributes) > 0 {
		span.SetAttributes(opts[0].Attributes...)
	}

	return fn(ctx)
}

// WithTraceFunc wraps a function with return value and OpenTelemetry tracing
func WithTraceFunc[T any](ctx context.Context, operationName string, fn func(ctx context.Context) (T, error), opts ...TraceOptions) (T, error) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, operationName)
	defer span.End()

	// Add attributes if provided
	if len(opts) > 0 && len(opts[0].Attributes) > 0 {
		span.SetAttributes(opts[0].Attributes...)
	}

	return fn(ctx)
}

// WithTraceNoError wraps a function with no error return and OpenTelemetry tracing
func WithTraceNoError(ctx context.Context, operationName string, fn func(ctx context.Context), opts ...TraceOptions) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, operationName)
	defer span.End()

	// Add attributes if provided
	if len(opts) > 0 && len(opts[0].Attributes) > 0 {
		span.SetAttributes(opts[0].Attributes...)
	}

	fn(ctx)
}

// SetSpanAttributes is a helper to add attributes to the current span
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := otelTrace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(attrs...)
	}
}