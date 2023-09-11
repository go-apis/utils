package internal

import (
	"context"
	"net/http"

	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

type standard struct {
	server *http.Server
}

func (s *standard) Start(ctx context.Context) error {
	// Run the server
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *standard) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func WithMetricsRecorder(h http.Handler) http.Handler {
	// Create our middleware.
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	// Wrap our main handler, we pass empty handler ID so the middleware
	// the handler label from the URL.
	return std.Handler("", mdlw, h)
}

func NewStandard(addr string, h http.Handler) Startable {
	server := &http.Server{Addr: addr, Handler: h}
	return &standard{server}
}
