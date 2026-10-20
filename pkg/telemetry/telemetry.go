package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config configures OpenTelemetry tracing.
type Config struct {
	Enabled     bool   `json:"enabled"`
	Endpoint    string `json:"endpoint"`
	Insecure    bool   `json:"insecure"`
	ServiceName string `json:"service-name"`
}

// DefaultConfig returns the default tracing configuration (disabled).
func DefaultConfig() Config {
	return Config{
		Enabled:     false,
		Endpoint:    "localhost:4318",
		Insecure:    true,
		ServiceName: "harns-platform",
	}
}

// Setup initialises the global tracer provider and returns a shutdown func.
// When tracing is disabled it returns a no-op shutdown.
func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("telemetry endpoint must not be empty")
	}
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter failed: %w", err)
	}

	name := cfg.ServiceName
	if name == "" {
		name = "harns-platform"
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(attribute.String("service.name", name)),
	)
	if err != nil {
		return nil, fmt.Errorf("create tracing resource failed: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	return tp.Shutdown, nil
}

// Middleware creates a span per request using the matched route template.
func Middleware(serviceName string) gin.HandlerFunc {
	if serviceName == "" {
		serviceName = "harns-platform"
	}
	tracer := otel.Tracer(serviceName)
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+route)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.target", c.Request.URL.Path),
			attribute.String("client.address", c.ClientIP()),
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", status))
		if status >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
		}
	}
}
