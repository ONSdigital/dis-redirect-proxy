package proxy

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/dis-redirect-proxy/response"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// Proxy provides a struct to wrap the proxy around
type Proxy struct {
	Router *mux.Router
}

// Setup function sets up the proxy and returns a Proxy
func Setup(_ context.Context, r *mux.Router, cfg *config.Config) *Proxy {
	proxy := &Proxy{
		Router: r,
	}

	r.PathPrefix("/").Name("Proxy Catch-All").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		proxy.manage(req.Context(), w, req, cfg)
	})
	return proxy
}

func (proxy *Proxy) manage(ctx context.Context, w http.ResponseWriter, req *http.Request, cfg *config.Config) {
	// Get the target URL based on the original request and the configuration
	targetURL := getTargetURL(req.URL.String(), cfg)

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
	return cfg.ProxiedServiceURL + requestURL
}
