package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/response"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// Proxy provides a struct to wrap the proxy around
type Proxy struct {
	Router      *mux.Router
	RedisClient RedisClient
}

// Setup function sets up the proxy and returns a Proxy
func ProxySetup(_ context.Context, r *mux.Router, cfg *config.Config, redisCli RedisClient) *Proxy {
	proxy := &Proxy{
		Router:      r,
		RedisClient: redisCli,
	}

	// Middleware for redirect check
	r.Use(proxy.redirectMiddleware(cfg, redisCli))

	r.PathPrefix("/").Name("Proxy Catch-All").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		proxy.manage(req.Context(), w, req, cfg)
	})
	return proxy
}

// redirectMiddleware checks Redis for a redirect URL
func (proxy *Proxy) redirectMiddleware(cfg *config.Config, redisCli RedisClient) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Only proceed with Redis check if feature flag is enabled
			if cfg.EnableRedirects {
				redirectURL, err := proxy.checkRedirect(req.URL.String(), req.Context(), redisCli)
				if err == nil && redirectURL != "" {
					// Redirect with 308 Permanent Redirect
					http.Redirect(w, req, redirectURL, http.StatusPermanentRedirect)
					return
				}
			}

			// Proceed to the next handler if no redirect is found or feature flag is off
			next.ServeHTTP(w, req)
		})
	}
}

// checkRedirect checks if a redirect exists in Redis
func (proxy *Proxy) checkRedirect(url string, ctx context.Context, redisClient RedisClient) (string, error) {
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

func (proxy *Proxy) manage(ctx context.Context, w http.ResponseWriter, req *http.Request, cfg *config.Config) {
	// Get the target URL based on the original request and the configuration
	targetURL := getTargetURL(req.URL.String(), cfg)

	log.Info(ctx, "Forwarding request to target", log.Data{"target_url": targetURL})

	// Create a new proxy request using the original request's context and body
	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, req.Body)
	if err != nil {
		// Log the error and respond with a 400 Bad Request
		log.Error(ctx, "error creating the proxy request", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Copy the headers from the original request to the proxy request
	proxyReq.Header = req.Header
	// Set the Host header explicitly to match the original request
	proxyReq.Host = req.Host

	// Create a new HTTP client with a custom redirect handler
	client := &http.Client{
		// Prevent automatic redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Send the proxy request to the target service
	serviceResponse, err := client.Do(proxyReq)
	if err != nil {
		log.Error(ctx, "error sending the proxy request", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Ensure the response body is closed once the function returns
	defer func() {
		if closeErr := serviceResponse.Body.Close(); closeErr != nil {
			log.Error(ctx, "error closing the response body", closeErr)
		}
	}()

	// Log the response status code for debugging purposes
	log.Info(ctx, "service response received", log.Data{"status_code": serviceResponse.StatusCode})

	response.WriteResponse(ctx, w, serviceResponse, req, cfg)
}

func getTargetURL(requestURL string, cfg *config.Config) string {
	// Ensure that ProxiedServiceURL contains a valid URL scheme
	if cfg.ProxiedServiceURL == "" {
		log.Error(context.Background(), "ProxiedServiceURL is not set", nil)
		return ""
	}
	return cfg.ProxiedServiceURL + requestURL
}
