package xservice

import (
	"context"
	"fmt"

	"github.com/iamolegga/enviper"
	"github.com/spf13/viper"
)

type Service struct {
	*ServiceConfig

	v *enviper.Enviper
}

func (c *Service) Parse(cfg interface{}) error {
	if err := c.v.Unmarshal(cfg, viper.DecodeHook(StringExpandEnv())); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

func NewService(ctx context.Context, v *viper.Viper) (*Service, error) {
	var e = enviper.New(v)
	e.AddConfigPath("./config")
	e.SetConfigName("default")

	svc := &Service{
		ServiceConfig: &ServiceConfig{
			Service:    "service",
			Version:    "1.0.0",
			SrvAddr:    ":8080",
			HealthAddr: ":8082",
		},
		v: e,
	}

	return svc, svc.Parse(svc.ServiceConfig)
}
