package service

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/redisClient.go -pkg mock . RedisClient
//go:generate moq -out mock/requestMiddleware.go -pkg mock . RequestMiddleware

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetRequestMiddleware() RequestMiddleware
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// RedisClient defines the required methods for RedisClient
type RedisClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetValue(ctx context.Context, key string) (string, error)
	GetTotalKeys(ctx context.Context) (int64, error)
	SetValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeleteValue(ctx context.Context, key string) error
}

// RequestMiddleware defines a method to get a middleware function that can modify the request.
// It is only needed to facilitate testing: the APIFeature in dp-component-test modifies the request to add a fake host
// and scheme. We need to remove this when running component tests, but don't need to do anything while in production
// mode. This can be achieved by injecting a middleware function (which will be different in production and test mode).
type RequestMiddleware interface {
	GetMiddlewareFunction() func(http.Handler) http.Handler
}
