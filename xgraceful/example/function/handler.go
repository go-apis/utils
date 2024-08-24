package function

import (
	"context"
	"log"
	"net/http"

	"github.com/go-apis/utils/xservice"
)

func NewHandler(ctx context.Context, cfg *xservice.Service) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello World"))
		if err != nil {
			log.Fatal(err)
		}
	}), nil
}
