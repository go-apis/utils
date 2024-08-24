package xservice

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
	Tracing    TracingConfig
	Metrics    MetricsConfig
}
