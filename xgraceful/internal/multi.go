package internal

import (
	"context"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
)

type multi struct {
	services []Startable

	shutdownOnce sync.Once
	shutdownErr  error
}

func (m *multi) Start(ctx context.Context) error {
	errs, ctx := errgroup.WithContext(ctx)

	for _, s := range m.services {
		current := s
		errs.Go(func() error {
			return current.Start(ctx)
		})
	}

	// When the group context is cancelled — because a service's Start failed
	// or the parent context was cancelled — proactively shut everything down.
	// http.Server.ListenAndServe ignores context, so without this the other
	// services would run forever and errs.Wait would block, swallowing the
	// original failure.
	errs.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = m.Shutdown(shutdownCtx)
		return nil
	})

	return errs.Wait()
}

func (m *multi) Shutdown(ctx context.Context) error {
	// Idempotent: the cancel watcher in Start and the external signal handler
	// may both call this; only the first invocation tears things down.
	m.shutdownOnce.Do(func() {
		// Shut down in reverse registration order so request-serving services
		// stop before the telemetry/logging they depend on is torn down.
		for i := len(m.services) - 1; i >= 0; i-- {
			if err := m.services[i].Shutdown(ctx); err != nil {
				m.shutdownErr = multierror.Append(m.shutdownErr, err)
			}
		}
	})
	return m.shutdownErr
}

func NewMulti(services ...Startable) Startable {

	filtered := make([]Startable, 0, len(services))
	for _, item := range services {
		if item != nil {
			filtered = append(filtered, item)
		}
	}

	return &multi{services: filtered}
}
