package pgx

import (
	"time"
)

type Config struct {
	DNSConnection string

	ReconnectTimeout time.Duration
	OpenLimit        int
	IdleLimit        int
	ConnLife         time.Duration

	ReCommitAttemptsNumber int
	ReCommitTimeout        time.Duration
}
