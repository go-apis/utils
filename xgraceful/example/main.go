package main

import (
	"context"

	"github.com/go-apis/utils/xgraceful"
	"github.com/go-apis/utils/xgraceful/example/function"
	"github.com/go-apis/utils/xservice"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := xservice.NewConfig(ctx)
	if err != nil {
		panic(err)
	}

	handler, err := function.NewHandler(ctx, cfg)
	if err != nil {
		panic(err)
	}

	xgraceful.Serve(ctx, cfg, handler)
	cancel()

	<-ctx.Done()
}
