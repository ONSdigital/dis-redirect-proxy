package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

type measuredTransport struct {
	base  http.RoundTripper
	count atomic.Uint64
}

func newMeasuredTransport(base http.RoundTripper) *measuredTransport {
	return &measuredTransport{
		base:  base,
		count: atomic.Uint64{},
	}
}

func (transport *measuredTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	var mu sync.Mutex
	var connReady, firstByte time.Duration
	var reused bool

	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			mu.Lock()
			connReady, reused = time.Since(start), info.Reused
			mu.Unlock()
		},
		GotFirstResponseByte: func() {
			mu.Lock()
			firstByte = time.Since(start)
			mu.Unlock()
		},
	}

	traced := req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := transport.base.RoundTrip(traced)
	toHeaders := time.Since(start)

	if toHeaders > 500*time.Millisecond || transport.count.Add(1)%100 == 0 {
		mu.Lock()
		data := log.Data{
			"connection_ready": connReady.String(),
			"first_byte":       firstByte.String(),
			"response_headers": toHeaders.String(),
			"reused":           reused,
			"error":            fmt.Sprint(err),
		}
		mu.Unlock()
		log.Info(req.Context(), "outbound timing", data)
	}
	return resp, err
}
