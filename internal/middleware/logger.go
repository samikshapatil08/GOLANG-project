package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (writer *loggingResponseWriter) WriteHeader(code int) {
	writer.statusCode = code
	writer.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		start := time.Now()
		wrapped := &loggingResponseWriter{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, req)

		duration := time.Since(start)
		slog.Info("http request",
			"method", req.Method,
			"path", req.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"ip", req.RemoteAddr,
		)
	})
}
