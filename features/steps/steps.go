package steps

import (
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

//
//import (
//	"context"
//	"fmt"
//	"github.com/maxcnunes/httpfake"
//	"io"
//	"net/http"
//	"strconv"
//	"strings"
//	"time"
//
//	"github.com/ONSdigital/dis-redirect-proxy/config"
//	"github.com/ONSdigital/dis-redirect-proxy/service"
//	"github.com/ONSdigital/dis-redirect-proxy/service/mock"
//	"github.com/ONSdigital/dp-healthcheck/healthcheck"
//	"github.com/ONSdigital/log.go/v2/log"
//	"github.com/cucumber/godog"
//	"github.com/stretchr/testify/assert"
//)
//
//const MsgHealthy = "redis is healthy"
//

func (c *ProxyComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the redirect proxy is running$`, c.theRedirectProxyIsRunning)
	//ctx.Step(`^I should receive a hello-world response$`, c.iShouldReceiveAHelloworldResponse)
	//ctx.Step(`^redis is healthy$`, c.redisIsHealthy)
	//ctx.Step(`^the redirect proxy is running$`, c.theRedirectProxyIsRunning)
	//ctx.Step(`^the redirect proxy is initialised$`, c.theRedirectProxyIsInitialised)
	//ctx.Step(`^I run the redirect proxy$`, c.iRunTheRedirectProxy)
	//ctx.Step(`^I should receive the following health JSON response:$`, c.iShouldReceiveTheFollowingHealthJSONResponse)
}

func (c *ProxyComponent) theRedirectProxyIsRunning() error {
	assert.Equal(c, true, c.ServiceRunning)

	return c.StepError()
}

//
//func (c *Component) iShouldReceiveAHelloworldResponse() error {
//	responseBody := c.apiFeature.HTTPResponse.Body
//	body, _ := io.ReadAll(responseBody)
//
//	assert.Equal(c, `{"message":"Hello, World!"}`, strings.TrimSpace(string(body)))
//
//	return c.StepError()
//}
//
//func (c *Component) theRedirectProxyIsRunning() error {
//	ctx := context.Background()
//	initFunctions := &mock.InitialiserMock{
//		DoGetHTTPServerFunc:  c.getHTTPServer,
//		DoGetHealthCheckFunc: c.getHealthCheckOK,
//	}
//
//	c.svcList = service.NewServiceList(initFunctions)
//
//	svcErrors := make(chan error, 1)
//	c.StartTime = time.Now()
//	var err error
//	c.svc, err = service.Run(ctx, c.Config, c.svcList, "1", gitCommitHash, appVersion, svcErrors)
//	if err != nil {
//		log.Error(ctx, "failed to init service", err)
//		return err
//	}
//	c.ServiceRunning = true
//	return nil
//}
//
//func (c *Component) theRedirectProxyIsInitialised() error {
//	initFunctions := &mock.InitialiserMock{
//		DoGetHTTPServerFunc:  c.getHTTPServer,
//		DoGetHealthCheckFunc: c.getHealthCheckOK,
//	}
//	c.svcList = service.NewServiceList(initFunctions)
//	return nil
//}
//
//func (c *Component) iRunTheRedirectProxy() error {
//	ctx := context.Background()
//	svcErrors := make(chan error, 1)
//	c.StartTime = time.Now()
//	var err error
//	fmt.Println("here 1")
//	c.svc, err = service.Run(ctx, c.Config, c.svcList, "1", gitCommitHash, appVersion, svcErrors)
//	if err != nil {
//		log.Error(ctx, "failed to run service", err)
//		return err
//	}
//	c.ServiceRunning = true
//	return nil
//}
//
//func (c *Component) getHealthCheckOK(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
//	componentBuildTime := strconv.Itoa(int(time.Now().Unix()))
//	versionInfo, err := healthcheck.NewVersionInfo(componentBuildTime, gitCommitHash, appVersion)
//	if err != nil {
//		return nil, err
//	}
//	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
//	return &hc, nil
//}
//
//func (c *Component) redisIsHealthy() error {
//	//redisClientMock := &mock.RedisClientMock{
//	//	CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
//	//		if state == nil {
//	//			state = &healthcheck.CheckState{}
//	//		}
//	//		if updateErr := state.Update(healthcheck.StatusOK, MsgHealthy, 200); updateErr != nil {
//	//			return updateErr
//	//		}
//	//		return nil
//	//	},
//	//}
//	//
//	//c.svcList.RedisCli = redisClientMock
//
//	c.fakeRedis.healthRequest.Lock()
//	defer c.fakeRedis.healthRequest.Unlock()
//
//	c.fakeRedis.healthRequest.CustomHandle = healthCheckStatusHandle(200)
//
//	return nil
//}
//
//func (c *Component) iShouldReceiveTheFollowingHealthJSONResponse(expectedResponse *godog.DocString) error {
//	responseBody := c.apiFeature.HTTPResponse.Body
//	body, _ := io.ReadAll(responseBody)
//
//	assert.Equal(c, expectedResponse, strings.TrimSpace(string(body)))
//
//	return c.StepError()
//}
//
//func healthCheckStatusHandle(status int) httpfake.Responder {
//	return func(w http.ResponseWriter, r *http.Request, rh *httpfake.Request) {
//		rh.Lock()
//		defer rh.Unlock()
//		w.WriteHeader(status)
//	}
//}
