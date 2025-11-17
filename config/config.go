package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dis-redirect-proxy
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	EnableReleasesFallback     bool          `envconfig:"ENABLE_RELEASES_FALLBACK"`
	EnableRedirects            bool          `envconfig:"ENABLE_REDIRECTS"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	RedisAddress               string        `envconfig:"REDIS_ADDRESS"`
	ProxiedServiceURL          string        `envconfig:"PROXIED_SERVICE_URL"`
	WagtailURL                 string        `envconfig:"WAGTAIL_URL"` // TODO consider naming
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:30000",
		EnableReleasesFallback:     false,
		EnableRedirects:            false,
		ProxiedServiceURL:          "http://localhost:20000",
		WagtailURL:                 "http://localhost:8000",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		OTBatchTimeout:             5 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dis-redirect-proxy",
		OtelEnabled:                false,
		RedisAddress:               "localhost:6379",
	}

	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	if cfg.ProxiedServiceURL == "" {
		return nil, fmt.Errorf("missing required config: PROXIED_SERVICE_URL")
	}

	return cfg, nil
}
