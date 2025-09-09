package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingConfig holds the configuration for distributed tracing
type TracingConfig struct {
	ServiceName     string
	ServiceVersion  string
	Environment     string
	JaegerEndpoint  string
	Enabled         bool
	SampleRate      float64
}

// InitTracing initializes OpenTelemetry with Jaeger exporter
func InitTracing(cfg TracingConfig, logger *zap.Logger) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		logger.Info("Distributed tracing is disabled")
		return sdktrace.NewTracerProvider(), nil
	}

	// Create Jaeger exporter
	jaegerExporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(cfg.JaegerEndpoint),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(jaegerExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to handle trace context and baggage
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	logger.Info("OpenTelemetry tracing initialized",
		zap.String("service", cfg.ServiceName),
		zap.String("version", cfg.ServiceVersion),
		zap.String("environment", cfg.Environment),
		zap.String("jaeger_endpoint", cfg.JaegerEndpoint),
		zap.Float64("sample_rate", cfg.SampleRate))

	return tp, nil
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider, logger *zap.Logger) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := tp.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown tracer provider", zap.Error(err))
	} else {
		logger.Info("Tracer provider shutdown successfully")
	}
}

// GetTracer returns a tracer for the cart service
func GetTracer() trace.Tracer {
	return otel.Tracer("cart-service")
}
