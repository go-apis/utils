package internal

import (
	"context"

	"github.com/contextcloud/goutils/xlog"
)

type zl struct{}

func (s *zl) Start(ctx context.Context) error {
	return nil
}

func (s *zl) Shutdown(ctx context.Context) error {
	l := xlog.Logger(ctx)
	l.Info("Shutting down gracefully")

	// swallow error
	l.Logger().Sync()
	return nil
}

func NewZapLog() Startable {
	return &zl{}
}
