package logrus

import (
	"intmax2-node/internal/logger"

	lr "github.com/sirupsen/logrus"
)

type logrus struct {
	log *lr.Entry
}

func New(logLevel, timeFormat string, logJSON, logLines bool) logger.Logger {
	if timeFormat == "" {
		timeFormat = logger.DefaultTimeFormat
	}

	level, _ := lr.ParseLevel(logLevel)

	l := lr.New()
	if logJSON {
		formatter := &lr.JSONFormatter{}
		formatter.TimestampFormat = timeFormat
		l.SetFormatter(formatter)
	} else {
		formatter := &lr.TextFormatter{}
		formatter.TimestampFormat = timeFormat
		formatter.FullTimestamp = true
		formatter.ForceColors = true
		l.SetFormatter(formatter)
	}
	l.SetLevel(level)
	l.SetReportCaller(logLines)

	return &logrus{
		log: lr.NewEntry(l),
	}
}

func (l *logrus) Debugf(format string, args ...any) {
	l.log.Debugf(format, args...)
}

func (l *logrus) Printf(format string, args ...any) {
	l.log.Debugf(format, args...)
}

func (l *logrus) Infof(format string, args ...any) {
	l.log.Infof(format, args...)
}

func (l *logrus) Warnf(format string, args ...any) {
	l.log.Warnf(format, args...)
}

func (l *logrus) Errorf(format string, args ...any) {
	l.log.Errorf(format, args...)
}

func (l *logrus) Fatalf(format string, args ...any) {
	l.log.Fatalf(format, args...)
}

func (l *logrus) Panicf(format string, args ...any) {
	l.log.Fatalf(format, args...)
}

func (l *logrus) WithFields(fields logger.Fields) logger.Logger {
	newLogger := l.log.WithFields(lr.Fields(fields))
	return &logrus{newLogger}
}

func (l *logrus) WithError(err error) logger.Logger {
	return l.WithFields(logger.Fields{"error": err})
}

func (l *logrus) Logf(level logger.Level, format string, args ...any) {
	switch level {
	case logger.DebugLevel:
		l.Debugf(format, args...)
	case logger.InfoLevel:
		l.Infof(format, args...)
	case logger.WarnLevel:
		l.Warnf(format, args...)
	case logger.ErrorLevel:
		l.Errorf(format, args...)
	case logger.FatalLevel:
		l.Fatalf(format, args...)
	case logger.PanicLevel:
		l.Panicf(format, args...)
	}
}
