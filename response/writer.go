package response

import (
	"context"
	"io"
	"net/http"

	"github.com/ONSdigital/dis-redirect-proxy/config"
	"github.com/ONSdigital/log.go/v2/log"
)

func WriteResponse(ctx context.Context, w http.ResponseWriter, serviceResponse *http.Response, req *http.Request, cfg *config.Config) {
	writeUnmodifiedResponse(ctx, w, serviceResponse)
}

func writeResponse(ctx context.Context, w http.ResponseWriter, serviceResponse *http.Response, overrideHeaders map[string]string) {
	// Copy the service response's headers
	for name, values := range serviceResponse.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set any new headers or overwrite existing
	for name, value := range overrideHeaders {
		w.Header().Set(name, value)
	}

	// Copy the service response's status code
	w.WriteHeader(serviceResponse.StatusCode)

	buf := make([]byte, 128*1024)

	// Copy the service response's body
	if _, err := io.CopyBuffer(w, serviceResponse.Body, buf); err != nil {
		log.Error(ctx, "error copying the proxy response's body", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func writeUnmodifiedResponse(ctx context.Context, w http.ResponseWriter, serviceResponse *http.Response) {
	noAdditionalHeaders := map[string]string{}
	writeResponse(ctx, w, serviceResponse, noAdditionalHeaders)
}
