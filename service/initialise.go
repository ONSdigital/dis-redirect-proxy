package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck bool
	Init        Initialiser
	RedisCli    Redis
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		HealthCheck: false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

//var GetRedisClient = func(ctx context.Context) (disRedis.Client, error) {
//	clientConfig := &disRedis.ClientConfig{}
//	redisClient, err := disRedis.NewClient(ctx, clientConfig)
//
//	if err != nil {
//		log.Error(ctx, "Failed to create dis-redis client", err)
//		return disRedis.Client{}, err
//	}
//
//	return *redisClient, nil
//}

var GetRedisClient = func(ctx context.Context) (Redis, error) {
	clientConfig := &disRedis.ClientConfig{}
	redisClient, err := disRedis.NewClient(ctx, clientConfig)

	if err != nil {
		log.Error(ctx, "Failed to create dis-redis client", err)
		return nil, err
	}

	return redisClient, nil
}
