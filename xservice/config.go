package xservice

import (
	"context"
	"fmt"

	"github.com/iamolegga/enviper"
	"github.com/spf13/viper"
)

var e = enviper.New(viper.New())

func init() {
	e.AddConfigPath("./config")
	e.SetConfigName("default")
}

type TracingConfig struct {
	Enabled bool
	Type    string
	Url     string
}

type ServiceConfig struct {
	Name        string
	Version     string
	SrvAddr     string
	MetricsAddr string
	HealthAddr  string
	Tracing     TracingConfig
}

func newConfig() *ServiceConfig {
	return &ServiceConfig{
		Name:        "service",
		Version:     "1.0.0",
		SrvAddr:     ":8080",
		MetricsAddr: ":8081",
		HealthAddr:  ":8082",
	}
}

func (c *ServiceConfig) Parse(cfg interface{}) error {
	if err := e.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

func NewConfig(ctx context.Context) (*ServiceConfig, error) {
	cfg := newConfig()
	if err := e.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
