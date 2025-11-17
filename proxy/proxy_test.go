package proxy_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	clientMocks "github.com/ONSdigital/dis-redirect-proxy/clients/mock"
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
		redisClientMock := &clientMocks.RedisMock{}
		redirectProxy, err := proxy.Setup(ctx, r, cfg, redisClientMock)
		So(err, ShouldBeNil)

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
		redisClientMock := &clientMocks.RedisMock{
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
				_, err := w.Write([]byte("This is the non-redirect URL response"))
				if err != nil {
					t.Fatalf("unexpected err writing mock response: %v", err)
				}
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
			redirectProxy, err := proxy.Setup(ctx, r, cfg, redisClientMock)
			So(err, ShouldBeNil)

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
			redirectProxy, err := proxy.Setup(ctx, r, cfg, redisClientMock)
			So(err, ShouldBeNil)

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
				t.Fatalf("unexpected err writing mock response: %v", err)
			}
		}))
		defer mockTargetServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:   false,
			ProxiedServiceURL: mockTargetServer.URL}
		redisCli := &clientMocks.RedisMock{}
		testProxy, err := proxy.Setup(ctx, router, cfg, redisCli)
		So(err, ShouldBeNil)

		Convey("When a request is sent", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

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
		redisCli := &clientMocks.RedisMock{}
		testProxy, err := proxy.Setup(ctx, router, cfg, redisCli)
		So(err, ShouldBeNil)

		Convey("When a request is sent", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return a 502 Bad Gateway", func() {
				So(w.Code, ShouldEqual, http.StatusBadGateway)
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
				t.Fatalf("unexpected err writing mock response: %v", err)
			}
		}))
		defer mockTargetServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:   false,
			ProxiedServiceURL: mockTargetServer.URL}
		redisCli := &clientMocks.RedisMock{}
		testProxy, err := proxy.Setup(ctx, router, cfg, redisCli)
		So(err, ShouldBeNil)

		Convey("When a request is sent", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy response should match the target's custom headers and body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Custom Body Content")
				So(w.Header().Get("Custom-Header"), ShouldEqual, "HeaderValue")
			})
		})
	})
}

func TestProxyHandleFallback(t *testing.T) {
	Convey("Given a Proxy with releases fallback enabled and a mock Wagtail that returns 404", t, func() {
		// Mock Wagtail server that always returns 404
		mockWagtailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("Mock Wagtail Response"))
			if err != nil {
				t.Fatalf("unexpected err writing mock response: %v", err)
			}
		}))
		defer mockWagtailServer.Close()

		// Mock Proxied service that returns 200 OK
		mockProxiedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Mock Proxied Server Response"))
			if err != nil {
				t.Fatalf("unexpected err writing mock response: %v", err)
			}
		}))
		defer mockProxiedServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:        false,
			EnableReleasesFallback: true,
			ProxiedServiceURL:      mockProxiedServer.URL,
			WagtailURL:             mockWagtailServer.URL,
		}
		redisCli := &clientMocks.RedisMock{}
		testProxy, err := proxy.Setup(ctx, router, cfg, redisCli)
		So(err, ShouldBeNil)

		Convey("When a request is sent to a /releases/ path", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/releases/some-release", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return the proxied service response", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Mock Proxied Server Response")
			})
		})
	})

	Convey("Given a Proxy with releases fallback enabled and a mock Wagtail that returns 200 OK", t, func() {
		// Mock Wagtail server that returns 200 OK
		mockWagtailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Println("Wagtail server received request")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Wagtail Response"))
			if err != nil {
				panic(err)
			}
		}))
		defer mockWagtailServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:        false,
			EnableReleasesFallback: true,
			ProxiedServiceURL:      "http://localhost:9999", // Unused in this test but must be valid.
			WagtailURL:             mockWagtailServer.URL,
		}
		redisCli := &clientMocks.RedisMock{}
		testProxy, err := proxy.Setup(ctx, router, cfg, redisCli)
		So(err, ShouldBeNil)

		Convey("When a request is sent to a /releases/ path", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/releases/some-release", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return the Wagtail response", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Wagtail Response")
			})
		})
	})

	Convey("Given a Proxy with releases fallback enabled and a mock Wagtail shouldn't be called", t, func() {
		// Mock Wagtail server that always returns 404
		mockWagtailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected call to test server: %s %s", r.Method, r.URL.Path)
		}))
		defer mockWagtailServer.Close()

		// Mock Proxied service that returns 200 OK
		mockProxiedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Mock Proxied Server Response"))
			if err != nil {
				t.Fatalf("unexpected err writing mock response: %v", err)
			}
		}))
		defer mockProxiedServer.Close()

		ctx := context.Background()
		router := mux.NewRouter()
		cfg := &config.Config{
			EnableRedirects:        false,
			EnableReleasesFallback: true,
			ProxiedServiceURL:      mockProxiedServer.URL,
			WagtailURL:             mockWagtailServer.URL,
		}
		redisCli := &clientMocks.RedisMock{}
		testProxy, err := proxy.Setup(ctx, router, cfg, redisCli)
		So(err, ShouldBeNil)

		Convey("When a request is sent to a non /releases/ path", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/some-other-path", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return the proxied service response", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Mock Proxied Server Response")

				Convey("And the mock Wagtail server should not have been called", func() {
					// Wagtail handler would have caused the test to fail if it was called
				})
			})
		})

		Convey("When a request is sent to the /releases path", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/releases", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return the proxied service response", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Mock Proxied Server Response")

				Convey("And the mock Wagtail server should not have been called", func() {
					// Wagtail handler would have caused the test to fail if it was called
				})
			})
		})

		Convey("When a request is sent to the /release/calendar path", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/release/calendar", http.NoBody)
			testProxy.Router.ServeHTTP(w, r)

			Convey("Then the proxy should return the proxied service response", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "Mock Proxied Server Response")

				Convey("And the mock Wagtail server should not have been called", func() {
					// Wagtail handler would have caused the test to fail if it was called
				})
			})
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
