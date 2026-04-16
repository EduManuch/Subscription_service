package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *slog.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			logger.Debug("incoming request",
				"method", r.Method,
				"path", r.URL.Path,
			)

			next.ServeHTTP(rw, r)
			duration := time.Since(start)

			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", duration.String(),
			)
		})
	}
}
