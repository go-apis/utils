package function

import (
	"context"
	"net/http"

	"github.com/contextcloud/goutils/xservice"
)

func NewHandler(ctx context.Context, cfg *xservice.ServiceConfig) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}), nil
}
