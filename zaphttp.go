// Package zaphttp contains middleware to help use the [zap] structured logger in your HTTP handlers.
//
// [zap]]: https://github.com/uber-go/zap
package zaphttp

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var loggerKey = contextKey("logger")

type Option func(*zap.Logger, *http.Request) *zap.Logger

// Retrieve the zap logger set with the SetLogger middleware from Context.
func Logger(c context.Context) *zap.Logger {
	if l := c.Value(loggerKey); l != nil {
		return l.(*zap.Logger)
	}
	return nil
}

// SetLogger is a middleware to set a zap logger in the request Context.
func SetLogger(logger *zap.Logger, opts ...Option) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logger
			for _, o := range opts {
				l = o(l, r)
			}
			c := context.WithValue(r.Context(), loggerKey, l)
			next.ServeHTTP(w, r.WithContext(c))
		})
	}
}
