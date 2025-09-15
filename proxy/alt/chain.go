package alt

import (
	"net/http"
)

type try struct {
	tryHandler http.Handler
}

// Try wraps a `http.Handler`
func Try(h http.Handler) *try {
	return &try{tryHandler: h}
}

type when struct {
	tryHandler http.Handler
	whenStatus int
}

func (t *try) WhenStatus(status int) *when {
	return &when{
		tryHandler: t.tryHandler,
		whenStatus: status,
	}
}

type Alternative struct {
	TryHandler   http.Handler
	WhenStatusIs int
	ThenHandler  http.Handler
}

func (w *when) Then(handler http.Handler) *Alternative {
	return &Alternative{
		TryHandler:   w.tryHandler,
		WhenStatusIs: w.whenStatus,
		ThenHandler:  handler,
	}
}

type responseWriter struct {
	header     http.Header
	statusCode int
	body       []byte
}

var _ http.ResponseWriter = &responseWriter{}

func (t *responseWriter) Header() http.Header {
	m := t.header
	if m == nil {
		m = make(http.Header)
		t.header = m
	}
	return m
}

func (t *responseWriter) Write(bytes []byte) (int, error) {
	// TODO initialise body
	t.body = append(t.body, bytes...)
	return len(bytes), nil
}

func (t *responseWriter) WriteHeader(statusCode int) {
	t.statusCode = statusCode
}

func (t *responseWriter) Body() []byte {
	return t.body
}

func (t *responseWriter) StatusCode() int {
	return t.statusCode
}

func (alternative *Alternative) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Split the request's body reader so that it can be read by both handlers
	splitter := NewReadCloserSplitter(r.Body)
	bodyReadCloser1 := splitter.NewReadCloser()
	defer bodyReadCloser1.Close() // TODO not sure we even need to do this, will handlers defer close anyway?
	bodyReadCloser2 := splitter.NewReadCloser()
	defer bodyReadCloser2.Close()

	w1 := responseWriter{} // TODO don't slurp response body of first request
	r1 := r.Clone(r.Context())
	r1.Body = bodyReadCloser1
	alternative.TryHandler.ServeHTTP(&w1, r1)

	if w1.statusCode != alternative.WhenStatusIs {
		for k, vs := range w1.header {
			w.Header()[k] = vs
		}
		w.WriteHeader(w1.statusCode)
		w.Write(w1.body)
		return
	}

	r2 := r.Clone(r.Context())
	r2.Body = bodyReadCloser2
	alternative.ThenHandler.ServeHTTP(w, r2)
}

func (alternative *Alternative) WhenStatus(status int) *when {
	return &when{
		tryHandler: alternative,
		whenStatus: status,
	}
}
