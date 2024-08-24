package internal

import (
	"context"

	"github.com/go-apis/utils/xlog"
	"go.uber.org/zap"
)

type zl struct{}

func (s *zl) Start(ctx context.Context) error {
	return nil
}

func (s *zl) Shutdown(ctx context.Context) error {
	l := xlog.Logger(ctx)
	l.Info("Shutting down gracefully")

	// swallow error
	if err := l.Logger().Sync(); err != nil {
		l.Error("issue syncing logger", zap.Error(err))
	}
	return nil
}

func NewZapLog() Startable {
	return &zl{}
}
