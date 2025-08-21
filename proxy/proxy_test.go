package proxy_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/clients/mock"
	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/proxy"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const nonRedirectURL = "/non-redirect-url"

func TestSetup(t *testing.T) {
	Convey("Given a Proxy instance", t, func() {
		ctx := context.Background()
		r := mux.NewRouter()
		cfg := &config.Config{}
		redisClientMock := &mock.RedisClientMock{}
		redirectProxy := proxy.Setup(ctx, r, cfg, redisClientMock)

		Convey("When created, all HTTP methods should be accepted", func() {
			So(hasRoute(redirectProxy.Router, "/", http.MethodGet), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodPost), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodPut), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodDelete), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodHead), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodConnect), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodOptions), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodTrace), ShouldBeTrue)
			So(hasRoute(redirectProxy.Router, "/", http.MethodPatch), ShouldBeTrue)
		})
	})
}

func TestProxyHandleRequestWithRedirect(t *testing.T) {
	Convey("Given a Proxy instance with a mock Redis client", t, func() {
		// Create a mock Redis client with inline method definition for GetValue
		redisClientMock := &mock.RedisClientMock{
			GetValueFunc: func(ctx context.Context, key string) (string, error) {
				switch key {
				case "/old-url":
					return "http://localhost:8081/new-url", nil
				case nonRedirectURL:
					return "", nil
				default:
					return "", nil
				}
			},
		}

		// Create a local mock server to simulate the target server for proxying requests
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/new-url":
				http.Redirect(w, r, "/final-url", http.StatusPermanentRedirect)
			case nonRedirectURL:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("This is the non-redirect URL response"))
			default:
				http.NotFound(w, r)
			}
		}))
		defer mockServer.Close()

		Convey("When EnableRedisRedirect is true", func() {
			// Set the ProxiedServiceURL to the mock server's URL
			cfg := &config.Config{
				EnableRedirects:   true,           // Enable the feature flag to test redirect
				ProxiedServiceURL: mockServer.URL, // Set the ProxiedServiceURL to the mock server's URL
			}

			ctx := context.Background()
			r := mux.NewRouter()
			redirectProxy := proxy.Setup(ctx, r, cfg, redisClientMock)

			Convey("It should register routes for all HTTP methods", func() {
				methods := []string{
					http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
					http.MethodHead, http.MethodConnect, http.MethodOptions, http.MethodTrace, http.MethodPatch,
				}
				for _, method := range methods {
					So(hasRoute(redirectProxy.Router, "/", method), ShouldBeTrue)
				}
			})

			Convey("Then the redirect middleware should be in place", func() {
				Convey("When a request triggers a redirect", func() {
					req, err := http.NewRequest("GET", "/old-url", http.NoBody)
					So(err, ShouldBeNil)
					rr := httptest.NewRecorder()
					redirectProxy.Router.ServeHTTP(rr, req)

					// Assert that a 308 redirect status is returned
					So(rr.Code, ShouldEqual, http.StatusPermanentRedirect)

					// Extract and parse the Location header
					location := rr.Header().Get("Location")
					So(location, ShouldNotBeEmpty)

					parsedURL, err := url.Parse(location)
					So(err, ShouldBeNil)

					// Assert the final redirect path
					So(parsedURL.Path, ShouldEqual, "/new-url")
				})

				Convey("When a request does not trigger a redirect", func() {
					req, err := http.NewRequest("GET", nonRedirectURL, http.NoBody)
					So(err, ShouldBeNil)
					rr := httptest.NewRecorder()
					redirectProxy.Router.ServeHTTP(rr, req)

					// Assert that a normal response (200 OK) is returned
					So(rr.Code, ShouldEqual, http.StatusOK)
				})
			})
		})

		Convey("When EnableRedisRedirect is false", func() {
			// Set the ProxiedServiceURL to the mock server's URL
			cfg := &config.Config{
				EnableRedirects:   false,          // Disable the feature flag to avoid redirect
				ProxiedServiceURL: mockServer.URL, // Set the ProxiedServiceURL to the mock server's URL
			}

			ctx := context.Background()
			r := mux.NewRouter()
			redirectProxy := proxy.Setup(ctx, r, cfg, redisClientMock)

			Convey("Then the middleware should forward the request", func() {
				// Test a request that should not trigger a redirect
				req, err := http.NewRequest("GET", nonRedirectURL, http.NoBody)
				So(err, ShouldBeNil)
				rr := httptest.NewRecorder()
				redirectProxy.Router.ServeHTTP(rr, req)

				// Assert that a normal response (200 OK) is returned
				So(rr.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestProxyHandleRequestOK(t *testing.T) {
	Convey("Given a Proxy and a mock target service", t, func() {
		mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("mock-header", "test")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Mock Target Response"))
			if err != nil {
				panic(err)
			}
		}))
		defer mockTargetServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:   false,
			ProxiedServiceURL: mockTargetServer.URL}
		redisCli := &mock.RedisClientMock{}
		proxy := proxy.Setup(ctx, router, cfg, redisCli)

		Convey("When a request is sent", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
			proxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy response should match the target response", func() {
				// Log the response for debugging
				fmt.Printf("Response Body: %s\n", w.Body.String())

				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Mock Target Response")
				So(w.Header().Get("mock-header"), ShouldEqual, "test")
			})
		})
	})
}

func TestProxyHandleRequestError(t *testing.T) {
	Convey("Given a Proxy with an invalid target URL", t, func() {
		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:   false,
			ProxiedServiceURL: "http://invalid-url"}
		redisCli := &mock.RedisClientMock{}
		proxy := proxy.Setup(ctx, router, cfg, redisCli)

		Convey("When a request is sent", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
			proxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return a 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestProxyHandleCustomHeaderAndBody(t *testing.T) {
	Convey("Given a Proxy and a mock target service with custom headers and body", t, func() {
		mockTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Custom-Header", "HeaderValue")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Custom Body Content"))
			if err != nil {
				panic(err)
			}
		}))
		defer mockTargetServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:   false,
			ProxiedServiceURL: mockTargetServer.URL}
		redisCli := &mock.RedisClientMock{}
		proxy := proxy.Setup(ctx, router, cfg, redisCli)

		Convey("When a request is sent", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
			proxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy response should match the target's custom headers and body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Custom Body Content")
				So(w.Header().Get("Custom-Header"), ShouldEqual, "HeaderValue")
			})
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
