// Package zaphttp contains middleware to help use the [zap] structured logger in your HTTP handlers.
//
// [zap]]: https://github.com/uber-go/zap
package zaphttp

import (
	"context"
	"net/http"

	"github.com/felixge/httpsnoop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var loggerKey = contextKey("logger")

// Retrieve the zap logger set with the SetLogger middleware from Context.
func Logger(c context.Context) *zap.Logger {
	if l := c.Value(loggerKey); l != nil {
		return l.(*zap.Logger)
	}
	return nil
}

// SetLogger is a middleware to set a zap logger in the request Context. If a
// X-Request-ID header is present, the logger will be set to a request scoped
// logger with a requestID field.
func SetLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logger
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				l = l.With(zap.String("requestID", requestID))
			}
			c := context.WithValue(r.Context(), loggerKey, l)
			next.ServeHTTP(w, r.WithContext(c))
		})
	}
}

// LogRequests is a middleware to log requests to a zap logger.
// The message will be set to "request" and the following request fields will be logged:
//   - requestID (string)
//   - method (string)
//   - url (string)
//   - status (int)
//   - latency (time.Duration)
//   - bytes (int64)
//
// The requestID field will only be set if a X-Request-ID header is present.
//
// The log level will be set to error if status >= 500, info otherwise.
func LogRequests(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logger
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				l = l.With(zap.String("requestID", requestID))
			}

			m := httpsnoop.CaptureMetrics(next, w, r)

			lvl := zapcore.InfoLevel
			if m.Code >= 500 {
				lvl = zapcore.ErrorLevel
			}

			l.Log(lvl, "request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int("status", m.Code),
				zap.Duration("latency", m.Duration),
				zap.Int64("bytes", m.Written),
			)
		})
	}
}
