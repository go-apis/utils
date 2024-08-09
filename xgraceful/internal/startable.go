package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-apis/utils/xservice"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Startable interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func NewStartable(cfg *xservice.ServiceConfig, h interface{}) Startable {
	switch start := h.(type) {
	case Startable:
		return start
	case http.Handler:
		inner := otelhttp.NewHandler(start, cfg.Service)
		return NewStandard(cfg.SrvAddr, inner)
	default:
		panic(fmt.Errorf("unknown service type: %T", h))
	}
}
