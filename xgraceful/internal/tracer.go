package internal

import (
	"context"
	"fmt"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/contextcloud/goutils/xservice"
	multierror "github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type tracer struct {
	tp *trace.TracerProvider
}

func (t *tracer) Start(ctx context.Context) error {
	return nil
}

func (t *tracer) Shutdown(ctx context.Context) error {
	var all error
	if err := t.tp.ForceFlush(ctx); err != nil {
		all = multierror.Append(all, err)
	}
	if err := t.tp.Shutdown(ctx); err != nil {
		all = multierror.Append(all, err)
	}
	return all
}

func traceExporter(ctx context.Context, cfg xservice.TracingConfig) (trace.SpanExporter, error) {
	switch cfg.Type {
	case "zipkin":
		return zipkin.New(cfg.Url)
	case "gcp":
		return texporter.New(texporter.WithProjectID(cfg.Url), texporter.WithContext(ctx))
	default:
		return nil, fmt.Errorf("unknown tracing exporter type: %s", cfg.Type)
	}
}

func NewTracer(ctx context.Context, cfg *xservice.ServiceConfig) Startable {
	if !cfg.Tracing.Enabled {
		return nil
	}

	exp, err := traceExporter(ctx, cfg.Tracing)
	if err != nil {
		panic(err)
	}

	res, err := resource.New(ctx,
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Name),
			semconv.ServiceVersionKey.String(cfg.Version),
		),
	)
	if err != nil {
		panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return &tracer{tp}
}
