package steps

import (
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

type ProxiedServiceFeature struct {
	Server     *httptest.Server
	Body       string
	StatusCode int
	Headers    map[string]string
}

func NewProxiedServiceFeature() *ProxiedServiceFeature {
	f := ProxiedServiceFeature{
		Headers: make(map[string]string),
	}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for headerName, headerValue := range f.Headers {
			w.Header().Set(headerName, headerValue)
		}

		w.WriteHeader(f.StatusCode)

		if _, err := w.Write([]byte(f.Body)); err != nil {
			panic(err)
		}
	}))

	return &f
}

func (f *ProxiedServiceFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the Proxied Service will send the following response:$`, f.proxiedServiceWillSendTheFollowingResponse)
	ctx.Step(`^the Proxied Service will send the following response with status "([^"]*)":$`, f.proxiedServiceWillSendTheFollowingResponseWithStatus)
	ctx.Step(`^the Proxied Service will set the "([^"]*)" header to "([^"]*)"$`, f.proxiedServiceWillSetTheHeaderTo)
	ctx.Step(`^the Proxied Service will set the HTTP status code to "([^"]*)"$`, f.proxiedServiceWillSetTheHTTPStatusCodeTo)
}

func (f *ProxiedServiceFeature) proxiedServiceWillSendTheFollowingResponse(proxiedServiceBody *godog.DocString) error {
	return f.proxiedServiceWillSendTheFollowingResponseWithStatus("200", proxiedServiceBody)
}

func (f *ProxiedServiceFeature) proxiedServiceWillSendTheFollowingResponseWithStatus(statusCodeStr string, proxiedServiceBody *godog.DocString) error {
	f.Body = strings.TrimSpace(proxiedServiceBody.Content)

	return f.proxiedServiceWillSetTheHTTPStatusCodeTo(statusCodeStr)
}

func (f *ProxiedServiceFeature) proxiedServiceWillSetTheHeaderTo(headerName, headerValue string) error {
	canonicalHeaderName := textproto.CanonicalMIMEHeaderKey(headerName)
	f.Headers[canonicalHeaderName] = headerValue

	return nil
}

func (f *ProxiedServiceFeature) proxiedServiceWillSetTheHTTPStatusCodeTo(statusCodeStr string) error {
	statusCode, err := strconv.Atoi(statusCodeStr)
	if err != nil {
		return err
	}

	f.StatusCode = statusCode

	return nil
}
