package internal

import (
	"context"

	"github.com/go-apis/utils/xservice"
	multierror "github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
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
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(cfg.Url),
		otlptracegrpc.WithInsecure(),
	)
	return exporter, err
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
			semconv.ServiceNameKey.String(cfg.Service),
			semconv.ServiceVersionKey.String(cfg.Version),
		),
	)
	if err != nil {
		panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		// set the sampling rate based on the parent span to 60%
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, // W3C Trace Context format; https://www.w3.org/TR/trace-context/
		),
	)

	return &tracer{tp}
}
