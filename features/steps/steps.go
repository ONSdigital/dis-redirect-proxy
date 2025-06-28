package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-redirect-proxy/service"
	"github.com/ONSdigital/dis-redirect-proxy/service/mock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// HealthCheckTest represents a test healthcheck struct that mimics the real healthcheck struct
type HealthCheckTest struct {
	Status    string                  `json:"status"`
	Version   healthcheck.VersionInfo `json:"version"`
	Uptime    time.Duration           `json:"uptime"`
	StartTime time.Time               `json:"start_time"`
	Checks    []*Check                `json:"checks"`
}

// Check represents a health status of a registered app that mimics the real check struct
// As the component test needs to access fields that are not exported in the real struct
type Check struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StatusCode  int        `json:"status_code"`
	Message     string     `json:"message"`
	LastChecked *time.Time `json:"last_checked"`
	LastSuccess *time.Time `json:"last_success"`
	LastFailure *time.Time `json:"last_failure"`
}

func (c *ProxyComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the redirect proxy is running$`, c.theRedirectProxyIsRunning)
	ctx.Step(`^I should receive the following health JSON response:$`, c.iShouldReceiveTheFollowingHealthJSONResponse)
	ctx.Step(`^I should receive an empty response$`, c.iShouldReceiveAnEmptyResponse)
	ctx.Step(`^the response from the Proxied Service should be returned unmodified by the Proxy$`, c.iShouldReceiveTheSameUnmodifiedResponseFromProxiedService)
	ctx.Step(`^the Proxy receives a GET request for "([^"]*)"$`, c.apiFeature.IGet)
	ctx.Step(`^the Proxy receives a POST request for "([^"]*)"$`, c.apiFeature.IPostToWithBody)
	ctx.Step(`^the Proxy receives a PUT request for "([^"]*)"$`, c.apiFeature.IPut)
	ctx.Step(`^the Proxy receives a PATCH request for "([^"]*)"$`, c.apiFeature.IPatch)
	ctx.Step(`^the Proxy receives a DELETE request for "([^"]*)"$`, c.apiFeature.IDelete)
	ctx.Step(`^the feature flag EnableRedirects is set to "([^"]*)"$`, c.theFeatureFlagEnableRedirectsIsSetTo)
}

func (c *ProxyComponent) SetEnableRedirects(enabled bool) error {
	if c.Config == nil {
		return fmt.Errorf("config is not initialized")
	}

	// Stop current service if running
	if c.ServiceRunning {
		if err := c.svc.Close(context.Background()); err != nil {
			return fmt.Errorf("failed to stop existing service: %w", err)
		}
		c.ServiceRunning = false
	}

	// Update the flag
	c.Config.EnableRedirects = enabled

	// Ensure required fields are filled
	c.Config.ProxiedServiceURL = c.proxiedServiceFeature.Server.URL
	c.Config.RedisAddress = c.redisFeature.Server.Addr()
	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.Config.BindAddr = bindAddress

	// Service initialiser mock setup
	initMock := &mock.InitialiserMock{
		DoGetHTTPServerFunc:        c.DoGetHTTPServer,
		DoGetHealthCheckFunc:       c.getHealthCheckOK,
		DoGetRequestMiddlewareFunc: c.DoGetRequestMiddleware,
	}
	c.svcList = service.NewServiceList(initMock)

	// Restart service with updated config
	var err error
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}
	c.ServiceRunning = true

	return nil
}

func (c *ProxyComponent) theFeatureFlagEnableRedirectsIsSetTo(value string) error {
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid boolean value: %q", value)
	}

	if err := c.SetEnableRedirects(enabled); err != nil {
		return err
	}

	return nil
}

func (c *ProxyComponent) theRedirectProxyIsRunning() {
	assert.Equal(c, true, c.ServiceRunning)
}

func (c *ProxyComponent) iShouldReceiveAnEmptyResponse() error {
	emptyResponse := &godog.DocString{Content: ""}
	return c.apiFeature.IShouldReceiveTheFollowingResponse(emptyResponse)
}

func (c *ProxyComponent) iShouldReceiveTheSameUnmodifiedResponseFromProxiedService() error {
	// Ensure the body is the same
	proxiedServiceBody := &godog.DocString{Content: c.proxiedServiceFeature.Body}
	err := c.apiFeature.IShouldReceiveTheFollowingResponse(proxiedServiceBody)
	if err != nil {
		return err
	}

	// Ensure all the headers that the tester set in the mock ProxiedService response are present in the Proxy response
	for name, value := range c.proxiedServiceFeature.Headers {
		err = c.apiFeature.TheResponseHeaderShouldBe(name, value)
		if err != nil {
			return err
		}
	}

	// Ensure all the headers in the Proxy response are the same as the ones the tester set in the mock ProxiedService response
	for name, values := range c.apiFeature.HTTPResponse.Header {
		if shouldEvaluateHeader(name) {
			for _, value := range values {
				proxiedServiceHeaderValue := c.proxiedServiceFeature.Headers[name]
				errorMessage := fmt.Sprintf(`The Proxy response's %q header has a different value to the one sent by ProxiedService`, name)
				assert.Equal(c, proxiedServiceHeaderValue, value, errorMessage)
			}
		}
	}

	// Ensure the status code is the same
	proxiedServiceStatusCode := strconv.Itoa(c.proxiedServiceFeature.StatusCode)
	err = c.apiFeature.TheHTTPStatusCodeShouldBe(proxiedServiceStatusCode)
	if err != nil {
		return err
	}

	return nil
}

func (c *ProxyComponent) iShouldReceiveTheFollowingHealthJSONResponse(expectedResponse *godog.DocString) error {
	var healthResponse, expectedHealth HealthCheckTest

	responseBody, err := io.ReadAll(c.apiFeature.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response of proxy component - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &healthResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response of proxy component - error: %v", err)
	}

	err = json.Unmarshal([]byte(expectedResponse.Content), &expectedHealth)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected health response - error: %v", err)
	}

	c.validateHealthCheckResponse(healthResponse, expectedHealth)

	return c.StepError()
}

func (c *ProxyComponent) validateHealthCheckResponse(healthResponse, expectedResponse HealthCheckTest) {
	maxExpectedStartTime := c.StartTime.Add((c.Config.HealthCheckInterval + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedResponse.Status, healthResponse.Status)
	assert.True(&c.ErrorFeature, healthResponse.StartTime.After(c.StartTime))
	assert.True(&c.ErrorFeature, healthResponse.StartTime.Before(maxExpectedStartTime))
	assert.Greater(&c.ErrorFeature, healthResponse.Uptime.Seconds(), float64(0))

	c.validateHealthVersion(healthResponse.Version, expectedResponse.Version, maxExpectedStartTime)

	for i, checkResponse := range healthResponse.Checks {
		c.validateHealthCheck(checkResponse, expectedResponse.Checks[i])
	}
}

func (c *ProxyComponent) validateHealthVersion(versionResponse, expectedVersion healthcheck.VersionInfo, maxExpectedStartTime time.Time) {
	assert.True(&c.ErrorFeature, versionResponse.BuildTime.Before(maxExpectedStartTime))
	assert.Equal(&c.ErrorFeature, expectedVersion.GitCommit, versionResponse.GitCommit)
	assert.Equal(&c.ErrorFeature, expectedVersion.Language, versionResponse.Language)
	assert.NotEmpty(&c.ErrorFeature, versionResponse.LanguageVersion)
	assert.Equal(&c.ErrorFeature, expectedVersion.Version, versionResponse.Version)
}

func (c *ProxyComponent) validateHealthCheck(checkResponse, expectedCheck *Check) {
	maxExpectedHealthCheckTime := c.StartTime.Add((c.Config.HealthCheckInterval + c.Config.HealthCheckCriticalTimeout + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedCheck.Name, checkResponse.Name)
	assert.Equal(&c.ErrorFeature, expectedCheck.Status, checkResponse.Status)
	assert.Equal(&c.ErrorFeature, expectedCheck.StatusCode, checkResponse.StatusCode)
	assert.Equal(&c.ErrorFeature, expectedCheck.Message, checkResponse.Message)
	assert.True(&c.ErrorFeature, checkResponse.LastChecked.Before(maxExpectedHealthCheckTime))
	assert.True(&c.ErrorFeature, checkResponse.LastChecked.After(c.StartTime))

	if expectedCheck.StatusCode == 200 {
		assert.True(&c.ErrorFeature, checkResponse.LastSuccess.Before(maxExpectedHealthCheckTime))
		assert.True(&c.ErrorFeature, checkResponse.LastSuccess.After(c.StartTime))
	} else {
		assert.True(&c.ErrorFeature, checkResponse.LastFailure.Before(maxExpectedHealthCheckTime))
		assert.True(&c.ErrorFeature, checkResponse.LastFailure.After(c.StartTime))
	}
}

// shouldEvaluateHeader helps determine which headers should be skipped when comparing the ProxiedService and the Proxy response
func shouldEvaluateHeader(headerName string) bool {
	switch headerName {
	case "Content-Length", "Content-Type", "Date":
		return false
	default:
		return true
	}
}
