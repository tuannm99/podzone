package telemetry

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerProvider is a wrapper around the OpenTelemetry TracerProvider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
}

// Shutdown stops the trace provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	return tp.provider.Shutdown(ctx)
}

// InitTracer creates a new trace provider
func InitTracer(serviceName, endpoint string) (*TracerProvider, error) {
	if endpoint == "" {
		log.Println("Telemetry endpoint not specified, using default (localhost:4317)")
		endpoint = "localhost:4317"
	}

	// Create a connection to the collector
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithGRPCConn(conn),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global trace provider
	otel.SetTracerProvider(provider)

	// Set the global propagator
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &TracerProvider{provider: provider}, nil
}

// Tracer creates a new tracer with the given name
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return Tracer("").Start(ctx, name)
}

// AddAttribute adds an attribute to the current span
func AddAttribute(ctx context.Context, key string, value string) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(semconv.ServiceNameKey.String(value))
}

// EndSpan ends the current span
func EndSpan(span trace.Span) {
	span.End()
}

// Error adds an error to the current span
func Error(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

// GRPCMiddleware returns a gRPC interceptor for tracing
func GRPCMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Start a new span
		ctx, span := StartSpan(ctx, info.FullMethod)
		defer span.End()

		// Handle the request
		resp, err := handler(ctx, req)
		if err != nil {
			Error(ctx, err)
		}

		return resp, err
	}
}
