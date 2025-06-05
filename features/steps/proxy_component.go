package steps

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/service"
	"github.com/ONSdigital/dis-redirect-proxy/service/mock"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	gitCommitHash = "132a3b8570fdfc9098757d841c8c058ddbd1c8fc"
	appVersion    = "v1.2.3"
)

type ProxyComponent struct {
	componentTest.ErrorFeature
	svcList        *service.ExternalServiceList
	svc            *service.Service
	errorChan      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	apiFeature     *componentTest.APIFeature
	redisFeature   *componentTest.RedisFeature
	StartTime      time.Time
}

func NewProxyComponent(redisFeat *componentTest.RedisFeature) (*ProxyComponent, error) {
	fmt.Println("I'm setting ServiceRunning to false in NewProxyComponent")
	c := &ProxyComponent{
		errorChan:      make(chan error),
		ServiceRunning: false,
		HTTPServer: &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.redisFeature = redisFeat
	c.Config.RedisConfig.Address = c.redisFeature.Server.Addr()

	initMock := &mock.InitialiserMock{
		DoGetHTTPServerFunc:  c.DoGetHTTPServer,
		DoGetHealthCheckFunc: c.getHealthCheckOK,
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.svcList = service.NewServiceList(initMock)

	c.Config.BindAddr = "localhost:0"
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return nil, err
	}
	fmt.Println("I'm setting ServiceRunning to true in NewProxyComponent")
	c.ServiceRunning = true

	return c, nil
}

func (c *ProxyComponent) InitAPIFeature() *componentTest.APIFeature {
	fmt.Println("In InitAPIFeature - calling InitialiseService")
	c.apiFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.apiFeature
}

func (c *ProxyComponent) Close() error {
	fmt.Println("I'm setting ServiceRunning to false in Close")
	if c.svc != nil && c.ServiceRunning {
		c.redisFeature.Server.Close()
		if err := c.svc.Close(context.Background()); err != nil {
			return err
		}
		c.ServiceRunning = false
	}
	return nil
}

// InitialiseService returns the http.Handler that's contained within a specific ProxyComponent.
func (c *ProxyComponent) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}

func (c *ProxyComponent) getHealthCheckOK(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	componentBuildTime := strconv.Itoa(int(time.Now().Unix()))
	versionInfo, err := healthcheck.NewVersionInfo(componentBuildTime, gitCommitHash, appVersion)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (c *ProxyComponent) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}
