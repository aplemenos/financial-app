package logger

import (
	"context"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// ContextKey is
type ContextKey string

const (
	// ContextKeyReqID - is the context key for RequestID
	ContextKeyReqID ContextKey = "requestID"

	// HTTPHeaderNameRequestID - is the name of the header for request ID
	HTTPHeaderNameRequestID = "X-Request-ID"
)

// GetReqID - returns reqID from a http request as a string
func GetReqID(ctx context.Context) string {
	reqID := ctx.Value(ContextKeyReqID)
	if ret, ok := reqID.(string); ok {
		return ret
	}

	return ""
}

// AttachReqID - attaches a brand new request ID to a http request
func AttachReqID(ctx context.Context) context.Context {
	reqID := uuid.NewV4()

	return context.WithValue(ctx, ContextKeyReqID, reqID.String())
}

// Middleware - attaches the reqID to the http.Request,
// and adds reqID to http header in the response
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := AttachReqID(r.Context())
		r = r.WithContext(ctx)

		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		// Log request details before being served
		log.WithFields(log.Fields{
			"uri":       uri,
			"method":    method,
			"requestID": GetReqID(ctx),
		}).Debug("request handled")

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		// Log request details when completed
		log.WithFields(log.Fields{
			"duration":  duration.Seconds(),
			"requestID": GetReqID(ctx),
		}).Debug("request completed")

		h := w.Header()
		h.Add(HTTPHeaderNameRequestID, GetReqID(ctx))
	})
}
