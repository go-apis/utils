package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/contextcloud/goutils/xservice"
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
		return NewStandard(cfg.SrvAddr, start)
	default:
		panic(fmt.Errorf("unknown service type: %T", h))
	}
}
