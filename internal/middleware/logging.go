package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging is a middleware that logs HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		wrapped := wrapResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request details
		log.Printf(
			"%s %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.status,
			time.Since(start),
		)
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// wrapResponseWriter creates a new responseWriter
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

// WriteHeader captures the status code and passes it to the wrapped ResponseWriter
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}