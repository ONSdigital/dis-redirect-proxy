package alt

import (
	"bytes"
	"io"
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

	// TODO slurping in the original rquest's body if there is one.
	var readBody []byte
	if r.Body != nil {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		readBody = body
	}

	w1 := responseWriter{}
	r1 := r.Clone(r.Context())
	r1.Body = io.NopCloser(bytes.NewBuffer(readBody))
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
	r2.Body = io.NopCloser(bytes.NewBuffer(readBody))
	alternative.ThenHandler.ServeHTTP(w, r2)
}

func (alternative *Alternative) WhenStatus(status int) *when {
	return &when{
		tryHandler: alternative,
		whenStatus: status,
	}
}
