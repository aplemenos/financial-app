package http

import (
	"context"
	"financial-app/util/logger"
	"net/http"
	"time"
)

func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(w, r)
	})
}

func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware - a handy middleware function that logs out incoming requests
// using request id through the context.
func LoggingMiddleware(next http.Handler) http.Handler {
	return logger.Middleware(next)
}
