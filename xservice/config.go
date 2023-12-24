package xservice

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/iamolegga/enviper"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var e = enviper.New(viper.New())

func init() {
	e.AddConfigPath("./config")
	e.SetConfigName("default")
}

type TracingConfig struct {
	Enabled bool
	Url     string
}
type MetricsConfig struct {
	Enabled bool
	Url     string
}

type ServiceConfig struct {
	Service    string
	Version    string
	SrvAddr    string
	HealthAddr string

	Tracing TracingConfig
	Metrics MetricsConfig
}

func StringExpandEnv() mapstructure.DecodeHookFuncKind {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.String {
			return data, nil
		}

		return os.ExpandEnv(data.(string)), nil
	}
}

func newConfig() *ServiceConfig {
	return &ServiceConfig{
		Service:    "service",
		Version:    "1.0.0",
		SrvAddr:    ":8080",
		HealthAddr: ":8082",
	}
}

func (c *ServiceConfig) Parse(cfg interface{}) error {
	if err := e.Unmarshal(cfg, viper.DecodeHook(StringExpandEnv())); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

func NewConfig(ctx context.Context) (*ServiceConfig, error) {
	cfg := newConfig()
	if err := e.Unmarshal(cfg, viper.DecodeHook(StringExpandEnv())); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
