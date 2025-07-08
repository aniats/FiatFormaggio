package tracing

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracing(appName string) (func(), error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(appName),
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
		if err = tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provIDer: %v", err)
		}
	}

	return cleanup, nil
}
