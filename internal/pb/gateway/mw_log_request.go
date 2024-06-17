package gateway

import (
	"intmax2-node/internal/logger"
	"net/http"
	"time"
)

type rw struct {
	http.ResponseWriter
	Code int
	Err  error
}

func (w *rw) WriteHeader(statusCode int) {
	w.Code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *rw) WriteError(err error) {
	w.Err = err
}

func logRequest(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rwCustom := &rw{ResponseWriter: w}

			next.ServeHTTP(rwCustom, r)

			var level logger.Level
			switch {
			case rwCustom.Code >= http.StatusInternalServerError:
				level = logger.ErrorLevel
			case rwCustom.Code >= http.StatusBadRequest:
				level = logger.WarnLevel
			case r.RequestURI == healthPath: // remove healthcheck from logs
				return
			default:
				level = logger.InfoLevel
			}

			const mask = "%d %s %s (%v)"
			log.Logf(
				level,
				mask,
				rwCustom.Code,
				r.Method,
				r.RequestURI,
				time.Since(started),
			)
		})
	}
}
