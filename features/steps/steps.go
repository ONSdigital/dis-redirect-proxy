package steps

import (
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	c.proxiedServiceFeature.RegisterSteps(ctx)

	ctx.Step(`^I should receive an empty response$`, c.iShouldReceiveAnEmptyResponse)
	ctx.Step(`^the response from the Proxied Service should be returned unmodified by the Proxy$`, c.iShouldReceiveTheSameUnmodifiedResponseFromProxiedService)
	ctx.Step(`^the Proxy receives a GET request for "([^"]*)"$`, c.apiFeature.IGet)
	ctx.Step(`^the Proxy receives a POST request for "([^"]*)"$`, c.apiFeature.IPostToWithBody)
	ctx.Step(`^the Proxy receives a PUT request for "([^"]*)"$`, c.apiFeature.IPut)
	ctx.Step(`^the Proxy receives a PATCH request for "([^"]*)"$`, c.apiFeature.IPatch)
	ctx.Step(`^the Proxy receives a DELETE request for "([^"]*)"$`, c.apiFeature.IDelete)
}

func (c *Component) iShouldReceiveAnEmptyResponse() error {
	emptyResponse := &godog.DocString{Content: ""}
	return c.apiFeature.IShouldReceiveTheFollowingResponse(emptyResponse)
}

func (c *Component) iShouldReceiveTheSameUnmodifiedResponseFromProxiedService() error {
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

	return c.StepError()
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
