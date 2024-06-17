package logger

const (
	DefaultTimeFormat = "2006-01-02T15:04:05.000Z0700"
)

// Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]any

type Level string

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = "debug"

	// InfoLevel is the default logging priority.
	InfoLevel Level = "info"

	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel Level = "warn"

	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel Level = "error"

	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = "fatal"

	// PanicLevel logs a message
	PanicLevel Level = "panic"
)

// Logger is our contract for the logger
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)
	Printf(format string, args ...any)
	WithFields(keyValues Fields) Logger
	WithError(err error) Logger
	Logf(level Level, format string, args ...any)
}
