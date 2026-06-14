package internal

import (
	"github.com/heptiolabs/healthcheck"
)

// NewHealth builds the health server. goroutineThreshold sets the max goroutine
// count for the liveness check; a value <= 0 disables that check entirely (the
// endpoint then only reflects process liveness). This matters for services that
// legitimately hold many goroutines (e.g. one per long-lived stream), where a
// fixed count would false-positive and trigger restarts.
func NewHealth(healthAddr string, goroutineThreshold int) Startable {
	health := healthcheck.NewHandler()
	if goroutineThreshold > 0 {
		health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(goroutineThreshold))
	}

	return NewStandard(healthAddr, "", "", health)
}
