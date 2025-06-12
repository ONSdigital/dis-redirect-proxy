package steps

import (
	"context"
	"net/http"
	"strconv"
	"strings"
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
	svcList               *service.ExternalServiceList
	svc                   *service.Service
	errorChan             chan error
	Config                *config.Config
	HTTPServer            *http.Server
	ServiceRunning        bool
	apiFeature            *componentTest.APIFeature
	redisFeature          *componentTest.RedisFeature
	StartTime             time.Time
	proxiedServiceFeature *ProxiedServiceFeature
}

func NewProxyComponent(redisFeat *componentTest.RedisFeature, proxiedServiceFeat *ProxiedServiceFeature) (*ProxyComponent, error) {
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

	c.proxiedServiceFeature = proxiedServiceFeat
	c.Config.ProxiedServiceURL = c.proxiedServiceFeature.Server.URL

	c.redisFeature = redisFeat
	c.Config.RedisAddress = c.redisFeature.Server.Addr()

	initMock := &mock.InitialiserMock{
		DoGetHTTPServerFunc:        c.DoGetHTTPServer,
		DoGetHealthCheckFunc:       c.getHealthCheckOK,
		DoGetRequestMiddlewareFunc: c.DoGetRequestMiddleware,
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.svcList = service.NewServiceList(initMock)

	c.Config.BindAddr = "localhost:0"
	c.StartTime = time.Now()
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return nil, err
	}
	c.ServiceRunning = true

	return c, nil
}

func (c *ProxyComponent) InitAPIFeature() *componentTest.APIFeature {
	c.apiFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.apiFeature
}

func (c *ProxyComponent) Reset() *ProxyComponent {
	c.apiFeature.Reset()
	return c
}

func (c *ProxyComponent) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.proxiedServiceFeature.Server.Close()
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
	c.HTTPServer = &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
		Addr:              bindAddr,
		Handler:           router,
	}
	return c.HTTPServer
}

func (c *ProxyComponent) DoGetRequestMiddleware() service.RequestMiddleware {
	return &HTTPTestRequestMiddleware{}
}

type HTTPTestRequestMiddleware struct{}

func (rm HTTPTestRequestMiddleware) GetMiddlewareFunction() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The APIFeature in dp-component-test appends "http://foo" to the request. In production, the scheme and
			// host are not set. This middleware removes them, so that the request looks like it would in production.
			r.URL.Scheme = ""
			r.URL.Host = ""

			requestURI, _ := strings.CutPrefix(r.RequestURI, "http://foo")
			r.RequestURI = requestURI

			next.ServeHTTP(w, r)
		})
	}
}
