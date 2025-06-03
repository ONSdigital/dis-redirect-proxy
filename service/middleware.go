package service

import "net/http"

type NoOpRequestMiddleware struct{}

func (rm NoOpRequestMiddleware) GetMiddlewareFunction() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
