package steps

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/service"
	"github.com/ONSdigital/dis-redirect-proxy/service/mock"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"net/http"
	"time"
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
		DoGetHealthCheckFunc: c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:  c.DoGetHTTPServer,
	}

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

//func (c *ProxyComponent) Reset() *ProxyComponent {
//	c.apiFeature.Reset()
//	c.redisFeature.Reset()
//	return c
//}

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

//func (c *ProxyComponent) InitialiseService() (http.Handler, error) {
//	c.Config.BindAddr = "localhost:0"
//	var err error
//	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
//	if err != nil {
//		return nil, err
//	}
//	fmt.Println("I'm setting ServiceRunning to true in InitialiseService")
//	c.ServiceRunning = true
//	return c.HTTPServer.Handler, nil
//}

// InitialiseService returns the http.Handler that's contained within a specific ProxyComponent.
func (c *ProxyComponent) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}

func (c *ProxyComponent) DoGetHealthcheckOk(_ *config.Config, _, _, _ string) (service.HealthChecker, error) {
	// nolint:revive // param names give context here.
	return &mock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *ProxyComponent) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}
