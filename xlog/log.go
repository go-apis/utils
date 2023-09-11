package xlog

import (
	"context"
	"sync"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

var (
	once   sync.Once
	logger *otelzap.Logger
)

func newLogger() *otelzap.Logger {
	var l *zap.Logger
	if IsDebugging() {
		l = zap.Must(zap.NewDevelopment())
	} else {
		l = zap.Must(zap.NewProduction())
	}
	return otelzap.New(l)
}

// Logger ensures that the caller does not forget to pass the context.
func Logger(ctx context.Context) otelzap.LoggerWithCtx {
	once.Do(func() {
		logger = newLogger()
	})
	return logger.Ctx(ctx)
}
