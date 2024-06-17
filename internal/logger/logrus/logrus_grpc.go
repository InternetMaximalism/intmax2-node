package logrus

import (
	lr "github.com/sirupsen/logrus"
)

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (l *logrus) Info(args ...any) {
	l.log.Info(args...)
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (l *logrus) Infoln(args ...any) {
	l.log.Infoln(args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (l *logrus) Warning(args ...any) {
	l.log.Warn(args...)
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (l *logrus) Warningln(args ...any) {
	l.log.Warnln(args...)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l *logrus) Warningf(format string, args ...any) {
	l.log.Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (l *logrus) Error(args ...any) {
	l.log.Error(args...)
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (l *logrus) Errorln(args ...any) {
	l.log.Errorln(args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *logrus) Fatal(args ...any) {
	l.log.Fatal(args...)
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *logrus) Fatalln(args ...any) {
	l.log.Fatalln(args...)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l *logrus) V(level int) bool {
	const (
		// infoLog indicates Info severity.
		infoLog int = iota
		// warningLog indicates Warning severity.
		warningLog
		// errorLog indicates Error severity.
		errorLog
		// fatalLog indicates Fatal severity.
		fatalLog
	)
	// logrus have levels from info(4) to fatal(1)
	// grpclog have levels from info(0) to fatal(3)
	var lev lr.Level
	switch level {
	case infoLog: // info
		lev = lr.InfoLevel
	case warningLog: // warn
		lev = lr.WarnLevel
	case errorLog: // error
		lev = lr.ErrorLevel
	case fatalLog: // fatal
		lev = lr.FatalLevel
	default:
		return false
	}

	return l.log.Logger.IsLevelEnabled(lev)
}
