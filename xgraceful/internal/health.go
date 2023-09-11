package internal

import (
	"github.com/heptiolabs/healthcheck"
)

func NewHealth(healthAddr string) Startable {
	health := healthcheck.NewHandler()
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

	return NewStandard(healthAddr, health)
}
