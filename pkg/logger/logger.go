package logger

import (
	"intmax2-node/internal/logger"
	"intmax2-node/internal/logger/logrus"
)

func New(logLevel, timeFormat string, logJSON, logLines bool) logger.Logger {
	return logrus.New(logLevel, timeFormat, logJSON, logLines)
}
