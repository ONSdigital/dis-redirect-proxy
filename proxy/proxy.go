package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ONSdigital/dis-redirect-proxy/clients"
	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/proxy/alt"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// Proxy provides a struct to wrap the proxy around
type Proxy struct {
	Router      *mux.Router
	RedisClient clients.Redis
}

// Setup function sets up the proxy and returns a Proxy
func Setup(_ context.Context, r *mux.Router, cfg *config.Config, redisCli clients.Redis) *Proxy {
	proxy := &Proxy{
		Router:      r,
		RedisClient: redisCli,
	}

	// Only create middleware with Redis check if feature flag is enabled
	if cfg.EnableRedirects {
		// Middleware for redirect check
		r.Use(proxy.redirectMiddleware(redisCli))
	}

	proxiedUrl, err := url.Parse(cfg.ProxiedServiceURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse proxied service url: %w", err)) //TODO handle error
	}
	proxyHandler := newReverseProxy(proxiedUrl)

	wagtailProxy, err := url.Parse(cfg.WagtailURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse wagtail proxied service url: %w", err))
	}
	wagtailProxyHandler := newReverseProxy(wagtailProxy)

	alternativeHandler := alt.Try(wagtailProxyHandler).WhenStatus(http.StatusNotFound).Then(proxyHandler)
	// TODO  feature flagged
	r.PathPrefix("/releases/*").Name("Release alternative").Handler(alternativeHandler)

	r.PathPrefix("/").Name("Proxy Catch-All").Handler(proxyHandler)
	return proxy
}

// redirectMiddleware checks Redis for a redirect URL
func (proxy *Proxy) redirectMiddleware(redisCli clients.Redis) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			redirectURL, err := proxy.checkRedirect(req.URL.String(), req.Context(), redisCli)
			if err == nil && redirectURL != "" {
				// Redirect with 308 Permanent Redirect
				http.Redirect(w, req, redirectURL, http.StatusPermanentRedirect)
				return
			}

			// If Redis returns an error (e.g., timeout, unavailable), we do not fail the request.
			// This is intentional: redirect support is a non-blocking enhancement,
			// and we prefer to serve the original content (legacy website) via the proxy
			// rather than return a 5xx to the user.
			// This ensures the service remains highly available even if Redis is degraded.

			// Proceed to the next handler if no redirect is found or feature flag is off
			next.ServeHTTP(w, req)
		})
	}
}

// checkRedirect checks if a redirect exists in Redis
func (proxy *Proxy) checkRedirect(url string, ctx context.Context, redisClient clients.Redis) (string, error) {
	// Get the redirect URL from Redis based on the incoming URL
	redirectURL, err := redisClient.GetValue(ctx, url)
	if err == redis.Nil {
		// If the key does not exist, return an empty string
		return "", nil
	} else if err != nil {
		// If an error occurs while checking Redis, log it and return the error
		log.Error(ctx, "error checking Redis for redirect", err)
		return "", err
	}

	// Return the found redirect URL
	return redirectURL, nil
}

func newReverseProxy(proxiedUrl *url.URL) *httputil.ReverseProxy {
	// TODO add end request logging
	// TODO consider other proxy options eg. timeouts, proxy-from-env etc. (see dp-frontend-router main.go for similar)
	reverseProxy := httputil.NewSingleHostReverseProxy(proxiedUrl)
	dir := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		log.Info(req.Context(), "forwarding request to target", log.Data{"request_url": req.URL.String(), "target": proxiedUrl.String()})
		dir(req)
	}

	return reverseProxy
}

func getTargetURL(requestURL string, cfg *config.Config) string {
	return cfg.ProxiedServiceURL + requestURL
}
