package internal

import (
	"context"
	"net/http"

	"github.com/contextcloud/goutils/xservice"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type metricser struct {
	server *http.Server
	mp     *sdkmetric.MeterProvider
}

func (s *metricser) Start(ctx context.Context) error {
	// Run the server
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
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
	if err := s.server.Shutdown(ctx); err != nil {
		all = multierror.Append(all, err)
	}
	return all
}

func NewMetrics(ctx context.Context, cfg *xservice.ServiceConfig) Startable {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
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
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(mp)

	server := &http.Server{Addr: cfg.MetricsAddr, Handler: promhttp.Handler()}
	return &metricser{server, mp}
}
