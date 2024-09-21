package errors

import (
	"intmax2-node/internal/logger"
	"strings"
)

func ErrScrollProcessing(err error, log logger.Logger, msg string, args ...any) bool {
	const (
		emptyKey = ""
		mKey     = "the Scroll processing error occurred"
	)

	switch {
	case
		strings.Contains(err.Error(), Err520ScrollWebServerStr),
		strings.Contains(err.Error(), Err502ScrollWebServerStr),
		strings.Contains(err.Error(), ErrInvalidSequenceStr):
		msg = strings.TrimSpace(msg)
		if msg == emptyKey {
			msg = mKey
		}
		log.WithError(err).Warnf(msg, args...)
		return true
	default:
		return false
	}
}

func ErrEthereumProcessing(err error, log logger.Logger, msg string, args ...any) bool {
	const (
		emptyKey = ""
		mKey     = "the Ethereum processing error occurred"
	)

	switch {
	case
		strings.Contains(err.Error(), Err502EthereumWevServerStr),
		strings.Contains(err.Error(), Err503EthereumWebServerStr):
		msg = strings.TrimSpace(msg)
		if msg == emptyKey {
			msg = mKey
		}
		log.WithError(err).Warnf(msg, args...)
		return true
	default:
		return false
	}
}
