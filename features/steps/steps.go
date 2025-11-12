package steps

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-redirect-proxy/service"
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
	ctx.Step(`^I should receive an empty response$`, c.iShouldReceiveAnEmptyResponse)
	ctx.Step(`^the response from the Proxied Service should be returned unmodified by the Proxy$`, c.iShouldReceiveTheSameUnmodifiedResponseFromProxiedService)
	ctx.Step(`^the response from the Wagtail Service should be returned unmodified by the Proxy$`, c.iShouldReceiveTheSameUnmodifiedResponseFromWagtailService)
	ctx.Step(`^the Proxy receives a GET request for "([^"]*)"$`, c.apiFeature.IGet)
	ctx.Step(`^the Proxy receives a POST request for "([^"]*)"$`, c.apiFeature.IPostToWithBody)
	ctx.Step(`^the Proxy receives a PUT request for "([^"]*)"$`, c.apiFeature.IPut)
	ctx.Step(`^the Proxy receives a PATCH request for "([^"]*)"$`, c.apiFeature.IPatch)
	ctx.Step(`^the Proxy receives a DELETE request for "([^"]*)"$`, c.apiFeature.IDelete)
	// TODO: abstract these to a common env var step definition
	ctx.Step(`^the feature flag EnableRedirects is set to "([^"]*)"$`, c.theFeatureFlagEnableRedirectsIsSetTo)
	ctx.Step(`^the feature flag EnableReleasesFallback is set to "([^"]*)"$`, c.theFeatureFlagEnableReleasesFallbackIsSetTo)
}

func (c *ProxyComponent) stopService() error {
	if err := c.svc.Close(context.Background()); err != nil {
		return fmt.Errorf("failed to stop existing service: %w", err)
	}
	c.ServiceRunning = false
	return nil
}

func (c *ProxyComponent) startService() error {
	var err error
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	c.ServiceRunning = true
	return nil
}

func (c *ProxyComponent) restartService() error {
	// Stop current service if running
	if c.ServiceRunning {
		if err := c.stopService(); err != nil {
			return err
		}
	}

	// Start service with updated config
	if err := c.startService(); err != nil {
		return err
	}
	return nil
}

func (c *ProxyComponent) SetEnableRedirects(enabled bool) error {
	if c.Config == nil {
		return fmt.Errorf("config is not initialized")
	}

	// Update the flag
	c.Config.EnableRedirects = enabled

	if err := c.restartService(); err != nil {
		return err
	}

	return nil
}

func (c *ProxyComponent) SetReleaseFallback(enabled bool) error {
	if c.Config == nil {
		return fmt.Errorf("config is not initialized")
	}

	// Update the flag
	c.Config.EnableReleasesFallback = enabled

	if err := c.restartService(); err != nil {
		return err
	}

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

func (c *ProxyComponent) theFeatureFlagEnableReleasesFallbackIsSetTo(value string) error {
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid boolean value: %q", value)
	}

	if err := c.SetReleaseFallback(enabled); err != nil {
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

func (c *ProxyComponent) checkResponseAgainstProxiedService(serviceFeature *ProxiedServiceFeature) error {
	// Ensure the body is the same
	body := &godog.DocString{Content: serviceFeature.Body}
	err := c.apiFeature.IShouldReceiveTheFollowingResponse(body)
	if err != nil {
		return err
	}

	// Ensure all the headers that the tester set in the mock ProxiedService response are present in the Proxy response
	for name, value := range serviceFeature.Headers {
		err = c.apiFeature.TheResponseHeaderShouldBe(name, value)
		if err != nil {
			return err
		}
	}

	// Ensure all the headers in the Proxy response are the same as the ones the tester set in the mock ProxiedService response
	for name, values := range c.apiFeature.HTTPResponse.Header {
		if shouldEvaluateHeader(name) {
			for _, value := range values {
				serviceHeaderValue := serviceFeature.Headers[name]
				errorMessage := fmt.Sprintf(`The Proxy response's %q header has a different value to the one sent by %s`, name, serviceFeature.Name)
				assert.Equal(c, serviceHeaderValue, value, errorMessage)
			}
		}
	}

	// Ensure the status code is the same
	serviceStatusCode := strconv.Itoa(serviceFeature.StatusCode)
	err = c.apiFeature.TheHTTPStatusCodeShouldBe(serviceStatusCode)
	if err != nil {
		return err
	}

	return nil
}

func (c *ProxyComponent) iShouldReceiveTheSameUnmodifiedResponseFromWagtailService() error {
	return c.checkResponseAgainstProxiedService(c.wagtailFeature)
}

func (c *ProxyComponent) iShouldReceiveTheSameUnmodifiedResponseFromProxiedService() error {
	return c.checkResponseAgainstProxiedService(c.proxiedServiceFeature)
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
