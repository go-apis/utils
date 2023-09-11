package internal

import (
	"context"
	"fmt"
	"net/http"
)

type Startable interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func NewStartable(srvAddr string, h interface{}) Startable {
	switch start := h.(type) {
	case Startable:
		return start
	case http.Handler:
		handler := WithMetricsRecorder(start)
		return NewStandard(srvAddr, handler)
	default:
		panic(fmt.Errorf("unknown service type: %T", h))
	}
}
