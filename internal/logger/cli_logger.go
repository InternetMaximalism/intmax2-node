package logger

import "fmt"

type commandLineLogger struct{}

func NewCommandLineLogger() Logger {
	return &commandLineLogger{}
}

func (l *commandLineLogger) Debugf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *commandLineLogger) Infof(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *commandLineLogger) Warnf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *commandLineLogger) Errorf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *commandLineLogger) Fatalf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

func (l *commandLineLogger) Panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

func (l *commandLineLogger) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *commandLineLogger) WithFields(keyValues Fields) Logger {
	return l
}

func (l *commandLineLogger) WithError(err error) Logger {
	return l
}

func (l *commandLineLogger) Logf(level Level, format string, args ...any) {
	switch level {
	case DebugLevel:
		l.Debugf(format, args...)
	case InfoLevel:
		l.Infof(format, args...)
	case WarnLevel:
		l.Warnf(format, args...)
	case ErrorLevel:
		l.Errorf(format, args...)
	case FatalLevel:
		l.Fatalf(format, args...)
	case PanicLevel:
		l.Panicf(format, args...)
	}
}
