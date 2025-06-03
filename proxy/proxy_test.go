package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given a Proxy instance", t, func() {
		ctx := context.Background()
		r := mux.NewRouter()
		cfg := &config.Config{}
		legacyCacheProxy := Setup(ctx, r, cfg)

		Convey("When created, all HTTP methods should be accepted", func() {
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodGet), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodPost), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodPut), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodDelete), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodHead), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodConnect), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodOptions), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodTrace), ShouldBeTrue)
			So(hasRoute(legacyCacheProxy.Router, "/", http.MethodPatch), ShouldBeTrue)
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
		cfg := &config.Config{ProxiedServiceURL: mockTargetServer.URL}

		proxy := Setup(ctx, router, cfg)

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
		cfg := &config.Config{ProxiedServiceURL: "http://invalid-url"}
		proxy := Setup(ctx, router, cfg)

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
		cfg := &config.Config{ProxiedServiceURL: mockTargetServer.URL}

		proxy := Setup(ctx, router, cfg)

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
