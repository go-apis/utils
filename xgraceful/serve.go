package xgraceful

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-apis/utils/xgraceful/internal"
	"github.com/go-apis/utils/xservice"
)

func run(ctx context.Context, s internal.Startable) {
	serverCtx, serverStopCtx := context.WithCancel(ctx)

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Print(err)
		}

		shutdownCancel()
		serverStopCtx()
	}()

	// start it.
	if err := s.Start(serverCtx); err != nil {
		panic(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

// defaultGoroutineThreshold preserves the historical liveness behavior when no
// option overrides it.
const defaultGoroutineThreshold = 100

type options struct {
	goroutineThreshold int
}

// Option configures Serve.
type Option func(*options)

// WithGoroutineThreshold sets the max goroutine count for the health server's
// liveness check. Use a higher value for services that hold many goroutines
// (e.g. one per long-lived stream). A value <= 0 disables the goroutine check.
func WithGoroutineThreshold(n int) Option {
	return func(o *options) { o.goroutineThreshold = n }
}

// WithoutGoroutineCheck disables the goroutine-count liveness check entirely.
func WithoutGoroutineCheck() Option {
	return func(o *options) { o.goroutineThreshold = -1 }
}

func Serve(ctx context.Context, cfg *xservice.ServiceConfig, handler interface{}, opts ...Option) {
	o := options{goroutineThreshold: defaultGoroutineThreshold}
	for _, opt := range opts {
		opt(&o)
	}

	multi := internal.NewMulti(
		internal.NewZapLog(),
		internal.NewTracer(ctx, cfg),
		internal.NewMetrics(ctx, cfg),
		internal.NewHealth(cfg.HealthAddr, o.goroutineThreshold),
		internal.NewStartable(cfg, handler),
	)
	run(ctx, multi)
}
