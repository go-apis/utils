package main

import (
	"context"

	"github.com/go-apis/utils/xgraceful"
	"github.com/go-apis/utils/xgraceful/example/function"
	"github.com/go-apis/utils/xservice"
	"github.com/spf13/viper"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	v := viper.New()

	cfg, err := xservice.NewConfig(ctx, v)
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
