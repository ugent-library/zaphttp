package zapchi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ugent-library/zaphttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func RequestID(l *zap.Logger, r *http.Request) *zap.Logger {
	return l.With(zap.String("requestID", middleware.GetReqID(r.Context())))
}

func RequestLogger() middleware.RequestLogger {
	return &requestLogger{}
}

type requestLogger struct{}

func (l *requestLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	logger := zaphttp.Logger(r.Context()).With(
		zap.String("remoteAddr", r.RemoteAddr),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
	)
	return &logEntry{
		logger: logger,
	}
}

type logEntry struct {
	logger *zap.Logger
}

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	lvl := zapcore.InfoLevel
	if status >= 500 {
		lvl = zapcore.ErrorLevel
	}

	e.logger.Log(lvl, "request",
		zap.Int("status", status),
		zap.Duration("latency", elapsed),
		zap.Int("bytes", bytes),
	)
}

func (e *logEntry) Panic(v any, stack []byte) {
	e.logger.Log(zapcore.PanicLevel, "request",
		zap.Any("panic", v),
		zap.ByteString("stack", stack),
	)
}
