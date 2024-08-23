package xservice

import (
	"context"
	"fmt"

	"github.com/iamolegga/enviper"
	"github.com/spf13/viper"
)

type TracingConfig struct {
	Enabled bool
	Url     string
}
type MetricsConfig struct {
	Enabled bool
	Url     string
}

type ServiceConfig struct {
	v *enviper.Enviper

	Service    string
	Version    string
	SrvAddr    string
	HealthAddr string
	Tracing    TracingConfig
	Metrics    MetricsConfig
}

func (c *ServiceConfig) Parse(cfg interface{}) error {
	if err := c.v.Unmarshal(cfg, viper.DecodeHook(StringExpandEnv())); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

func NewConfig(ctx context.Context, v *viper.Viper) (*ServiceConfig, error) {
	var e = enviper.New(v)
	e.AddConfigPath("./config")
	e.SetConfigName("default")

	cfg := &ServiceConfig{
		v:          e,
		Service:    "service",
		Version:    "1.0.0",
		SrvAddr:    ":8080",
		HealthAddr: ":8082",
	}
	if err := e.Unmarshal(cfg, viper.DecodeHook(StringExpandEnv())); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
