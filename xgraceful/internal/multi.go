package internal

import (
	"context"

	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
)

type multi struct {
	services []Startable
}

func (m *multi) Start(ctx context.Context) error {
	errs, ctx := errgroup.WithContext(ctx)

	for _, s := range m.services {
		current := s
		errs.Go(func() error {
			return current.Start(ctx)
		})
	}

	return errs.Wait()
}

func (m *multi) Shutdown(ctx context.Context) error {
	var all error
	for _, s := range m.services {
		if err := s.Shutdown(ctx); err != nil {
			all = multierror.Append(all, err)
		}
	}
	return all
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
