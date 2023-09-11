package internal

import (
	"context"
	"time"

	"github.com/contextcloud/goutils/xservice"
	multierror "github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type metricser struct {
	mp *sdkmetric.MeterProvider
}

func (s *metricser) Start(ctx context.Context) error {
	return nil
}

func (s *metricser) Shutdown(ctx context.Context) error {
	var all error
	if err := s.mp.ForceFlush(ctx); err != nil {
		all = multierror.Append(all, err)
	}
	if err := s.mp.Shutdown(ctx); err != nil {
		all = multierror.Append(all, err)
	}
	return all
}

func NewMetrics(ctx context.Context, cfg *xservice.ServiceConfig) Startable {
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(cfg.Metrics.Url),
	)
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

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second)),
		),
	)
	otel.SetMeterProvider(mp)

	return &metricser{mp}
}
