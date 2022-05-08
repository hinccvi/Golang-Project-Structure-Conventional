package log

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"gopkg.in/natefinch/lumberjack.v2"
)

var fileName = "app.log"

// Logger is a logger that supports log levels, context and structured logging.
type Logger interface {
	// With returns a logger based off the root logger and decorates it with the given context and arguments.
	With(ctx context.Context, args ...interface{}) Logger

	// Debug uses fmt.Sprint to construct and log a message at DEBUG level
	Debug(args ...interface{})
	// Info uses fmt.Sprint to construct and log a message at INFO level
	Info(args ...interface{})
	// Error uses fmt.Sprint to construct and log a message at ERROR level
	Error(args ...interface{})

	// Debugf uses fmt.Sprintf to construct and log a message at DEBUG level
	Debugf(format string, args ...interface{})
	// Infof uses fmt.Sprintf to construct and log a message at INFO level
	Infof(format string, args ...interface{})
	// Errorf uses fmt.Sprintf to construct and log a message at ERROR level
	Errorf(format string, args ...interface{})
}

type logger struct {
	*zap.SugaredLogger
}

type contextKey int

const (
	RequestIDKey contextKey = iota
	CorrelationIDKey
)

// Customize log encoder
func Encoder(mode string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format("2006-01-02T15:04:05Z0700"))
	}
	if mode == "local" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// New creates a new logger
func New(mode string) Logger {
	c := new(zapcore.Core)

	if mode == "local" {
		*c = zapcore.NewTee(zapcore.NewCore(Encoder(mode), zapcore.Lock(os.Stdout), zap.DebugLevel))
	} else {
		*c = zapcore.NewTee(zapcore.NewCore(Encoder(mode), WriteSyncer(), zap.DebugLevel))
	}

	l := zap.New(*c, zap.AddCaller(), zap.AddCallerSkip(0))

	return NewWithZap(l)
}

func WriteSyncer() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    500,
		MaxBackups: 7,
		MaxAge:     30,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// NewWithZap creates a new logger using the preconfigured zap logger.
func NewWithZap(l *zap.Logger) Logger {
	return &logger{l.Sugar()}
}

// NewForTest returns a new logger and the corresponding observed logs which can be used in unit tests to verify log entries.
func NewForTest() (Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	return NewWithZap(zap.New(core)), recorded
}

// With returns a logger based off the root logger and decorates it with the given context and arguments.
//
// If the context contains request ID and/or correlation ID information (recorded via WithRequestID()
// and WithCorrelationID()), they will be added to every log message generated by the new logger.
//
// The arguments should be specified as a sequence of name, value pairs with names being strings.
// The arguments will also be added to every log message generated by the logger.
func (l *logger) With(ctx context.Context, args ...interface{}) Logger {
	if ctx != nil {
		if id, ok := ctx.Value(RequestIDKey).(string); ok {
			args = append(args, zap.String("request_id", id))
		}
		if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
			args = append(args, zap.String("correlation_id", id))
		}
	}
	if len(args) > 0 {
		return &logger{l.SugaredLogger.With(args...)}
	}
	return l
}

// WithRequest returns a context which knows the request ID and correlation ID in the given request.
func WithRequest(ctx context.Context, req *http.Request) context.Context {
	id := getRequestID(req)
	if id == "" {
		id = uuid.New().String()
	}
	ctx = context.WithValue(ctx, RequestIDKey, id)
	if id := getCorrelationID(req); id != "" {
		ctx = context.WithValue(ctx, CorrelationIDKey, id)
	}
	return ctx
}

// getCorrelationID extracts the correlation ID from the HTTP request
func getCorrelationID(req *http.Request) string {
	return req.Header.Get("X-Correlation-ID")
}

// getRequestID extracts the correlation ID from the HTTP request
func getRequestID(req *http.Request) string {
	return req.Header.Get("X-Request-ID")
}