package function

import (
	"context"
	"net/http"

	"github.com/go-apis/utils/xservice"
)

func NewHandler(ctx context.Context, cfg *xservice.Service) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}), nil
}
