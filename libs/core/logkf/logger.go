package logkf

import (
	"encoding/json"
	"fmt"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO move "with" keys to const

const (
	Instance = "instance"
	Platform = "platform"
	Service  = "service"

	ComponentId     = "componentId"
	ComponentCommit = "componentCommit"
	ComponentName   = "componentName"

	EventId   = "eventId"
	EventType = "eventType"
	TraceId   = "traceId"
	SpanId    = "spanId"
)

var (
	Global *Logger
)

type Logger struct {
	wrapped *zap.SugaredLogger
}

func init() {
	Global, _ = BuildLogger("console", "debug")
}

func BuildLogger(format, level string) (*Logger, error) {
	var cfg zap.Config
	switch format {
	case "console":
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04.05")
	case "json":
		cfg = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("unrecognized log format: %s", format)
	}
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg.Level = lvl

	z, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	// ensures log messages are shown to be from caller instead of this logger
	z = z.WithOptions(zap.AddCallerSkip(1))

	return &Logger{wrapped: z.Sugar()}, nil
}

func (log *Logger) Unwrap() *zap.SugaredLogger {
	return log.wrapped
}

func (log *Logger) Named(name string) *Logger {
	return &Logger{wrapped: log.wrapped.Named(name)}
}

// With adds a variadic number of fields to the logging context. It accepts a
// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
// processing pairs, the first element of the pair is used as the field key and
// the second as the field value.
//
// For example,
//
//	 logger.With(
//	   "hello", "world",
//	   "failure", errors.New("oh no"),
//	   Stack(),
//	   "count", 42,
//	   "user", User{Name: "alice"},
//	)
//
// Note that the keys in key-value pairs should be strings. In console mode,
// passing a non-string key panics. In production, the logger is more forgiving:
// a separate error is logged, but the key-value pair is skipped and execution
// continues. Passing an orphaned key triggers similar behavior: panics in
// console mode and errors in production.
func (log *Logger) With(args ...interface{}) *Logger {
	return &Logger{wrapped: log.wrapped.With(args...)}
}

func (log *Logger) WithInstance(val string) *Logger {
	return log.With(Instance, val)
}

func (log *Logger) WithPlatform(val string) *Logger {
	return log.With(Platform, val)
}

func (log *Logger) WithService(val string) *Logger {
	return log.With(Service, val)
}

func (log *Logger) WithComponent(comp *kubefox.Component) *Logger {
	if comp == nil {
		return log
	}
	return log.With(
		ComponentId, comp.Id,
		ComponentCommit, comp.Commit,
		ComponentName, comp.Name,
	)
}

func (log *Logger) WithEvent(evt *kubefox.Event) *Logger {
	if evt == nil {
		return log
	}
	return log.With(
		EventId, evt.Id,
		EventType, evt.Type,
		TraceId, evt.GetTraceId(),
		SpanId, evt.GetSpanId(),
	)
}

func (log *Logger) WithSpan(traceId, spanId string) *Logger {
	return log.With(
		TraceId, traceId,
		SpanId, spanId,
	)
}

// DebugInterface marshals the interface to indented JSON. Since this operation
// is expensive it is only performed if the logger is at debug level.
func (log *Logger) DebugInterface(msg string, v interface{}) {
	if !log.wrapped.Desugar().Core().Enabled(zap.DebugLevel) {
		return
	}
	out, _ := json.MarshalIndent(v, "", "\t")
	log.wrapped.Debugf("%s\n%s", msg, out)
}

func (log *Logger) DebugEw(msg string, evt *kubefox.Event, keysAndValues ...interface{}) {
	if evt != nil {
		keysAndValues = append(keysAndValues,
			EventId, evt.Id,
			EventType, evt.Type,
			TraceId, evt.GetTraceId(),
			SpanId, evt.GetSpanId(),
		)
	}
	log.wrapped.Debugw(msg, keysAndValues...)
}

func (log *Logger) ErrorEw(msg string, evt *kubefox.Event, keysAndValues ...interface{}) {
	if evt != nil {
		keysAndValues = append(keysAndValues,
			EventId, evt.Id,
			EventType, evt.Type,
			TraceId, evt.GetTraceId(),
			SpanId, evt.GetSpanId(),
		)
	}

	log.wrapped.Errorw(msg, keysAndValues...)
}

// ErrorN creates an error with fmt.Errorf, logs it, and returns the error.
func (l *Logger) ErrorN(template string, args ...interface{}) error {
	err := fmt.Errorf(template, args...)
	l.wrapped.Error(err)
	return err
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Logger) Debug(args ...interface{}) {
	l.wrapped.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func (l *Logger) Info(args ...interface{}) {
	l.wrapped.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *Logger) Warn(args ...interface{}) {
	l.wrapped.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Error(args ...interface{}) {
	l.wrapped.Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In console mode, the
// logger then panics. (See DPanicLevel for details.)
func (l *Logger) DPanic(args ...interface{}) {
	l.wrapped.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (l *Logger) Panic(args ...interface{}) {
	l.wrapped.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (l *Logger) Fatal(args ...interface{}) {
	l.wrapped.Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.wrapped.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.wrapped.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.wrapped.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.wrapped.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In console mode, the
// logger then panics. (See DPanicLevel for details.)
func (l *Logger) DPanicf(template string, args ...interface{}) {
	l.wrapped.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.wrapped.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.wrapped.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.wrapped.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.wrapped.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.wrapped.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.wrapped.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In console mode, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (l *Logger) DPanicw(msg string, keysAndValues ...interface{}) {
	l.wrapped.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The variadic
// key-value pairs are treated as they are in With.
func (l *Logger) Panicw(msg string, keysAndValues ...interface{}) {
	l.wrapped.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (l *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.wrapped.Fatalw(msg, keysAndValues...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.wrapped.Sync()
}
