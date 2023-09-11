package main

import (
	"context"

	"github.com/contextcloud/goutils/xgraceful"
	"github.com/contextcloud/goutils/xgraceful/example/function"
	"github.com/contextcloud/goutils/xservice"
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
