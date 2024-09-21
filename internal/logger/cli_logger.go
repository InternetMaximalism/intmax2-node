package logger

import (
	"fmt"
)

type commandLineLogger struct {
	fields map[string]any
	err    error
}

func NewCommandLineLogger() Logger {
	return &commandLineLogger{
		fields: make(map[string]any),
	}
}

func (l *commandLineLogger) printMetaDataAndReset() {
	for k, v := range l.fields {
		fmt.Printf("%s = %v", k, v)
	}
	if l.err != nil {
		fmt.Printf("error = %v", l.err)
	}

	l.fields = make(map[string]any)
	l.err = nil
}

func (l *commandLineLogger) Debugf(format string, args ...any) {
	fmt.Printf(format, args...)

	l.printMetaDataAndReset()
}

func (l *commandLineLogger) Infof(format string, args ...any) {
	fmt.Printf(format, args...)

	l.printMetaDataAndReset()
}

func (l *commandLineLogger) Warnf(format string, args ...any) {
	fmt.Printf(format, args...)

	l.printMetaDataAndReset()
}

func (l *commandLineLogger) Errorf(format string, args ...any) {
	fmt.Printf(format, args...)

	l.printMetaDataAndReset()
}

func (l *commandLineLogger) Fatalf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

func (l *commandLineLogger) Panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

func (l *commandLineLogger) Printf(format string, args ...any) {
	fmt.Printf(format, args...)

	l.printMetaDataAndReset()
}

func (l *commandLineLogger) WithFields(keyValues Fields) Logger {
	for k, v := range keyValues {
		l.fields[k] = v
	}

	return l
}

func (l *commandLineLogger) WithError(err error) Logger {
	l.err = err

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
