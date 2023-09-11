package xservice

import (
	"context"
	"os"
	"testing"
)

type Demo struct {
	MyCustom string
}

func Test_It(t *testing.T) {
	os.Setenv("TRACING_ENABLED", "true")
	os.Setenv("TRACING_URL", "http://localhost:8080")
	os.Setenv("TRACING_TYPE", "jaeger")
	os.Setenv("HEALTHADDR", ":9000")
	os.Setenv("MYCUSTOM", "custom")

	ctx := context.Background()
	cfg, err := NewConfig(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	demo := &Demo{}
	if err := cfg.Parse(demo); err != nil {
		t.Error(err)
		return
	}

	if cfg == nil {
		t.Error("cfg is nil")
		return
	}
	if cfg.Tracing.Enabled != true {
		t.Error("cfg.Tracing.Enabled is not true")
		return
	}
	if cfg.Tracing.Url != "http://localhost:8080" {
		t.Error("cfg.Tracing.Url is not http://localhost:8080")
		return
	}
	if cfg.Service != "service" {
		t.Error("cfg.ServiceName is not service")
		return
	}
	if cfg.HealthAddr != ":9000" {
		t.Error("cfg.HealthAddr is not :9000")
		return
	}
	if demo.MyCustom != "custom" {
		t.Error("demo.MyCustom is not custom")
		return
	}
}
