package steps

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/service"
	"github.com/ONSdigital/dis-redirect-proxy/service/mock"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	gitCommitHash = "3t7e5s1t4272646ef477f8ed755"
	appVersion    = "v1.2.3"
)

type Component struct {
	componentTest.ErrorFeature
	svcList        *service.ExternalServiceList
	svc            *service.Service
	svcErrors      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	apiFeature     *componentTest.APIFeature
	StartTime      time.Time
}

func NewRedirectProxyComponent() (c *Component, err error) {
	c = &Component{
		HTTPServer: &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		},
		svcErrors: make(chan error),
	}

	ctx := context.Background()

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second

	log.Info(ctx, "configuration for component test", log.Data{"config": c.Config})

	return c, nil
}

func (c *Component) InitAPIFeature() *componentTest.APIFeature {
	c.apiFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.apiFeature
}

func (c *Component) Reset() *Component {
	c.apiFeature.Reset()
	return c
}

func (c *Component) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}
	return nil
}

func (c *Component) InitialiseService() (http.Handler, error) {
	var err error
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", gitCommitHash, appVersion, c.svcErrors)
	if err != nil {
		return nil, err
	}

	c.ServiceRunning = true
	return c.HTTPServer.Handler, nil
}

func (c *Component) DoGetHealthcheckOk(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	return &mock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *Component) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func (c *Component) getHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}
