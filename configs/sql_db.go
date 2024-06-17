package configs

import (
	"time"
)

type SQLDbReCommit struct {
	AttemptsNumber int           `env:"SQL_DB_RECOMMIT_ATTEMPTS_NUMBER" envDefault:"50"`
	Timeout        time.Duration `env:"SQL_DB_RECOMMIT_TIMEOUT" envDefault:"1s"`
}

type SQLDb struct {
	DriverName       string        `env:"SQL_DB_APP_DRIVER_NAME" envDefault:"pgx"`
	DNSConnection    string        `env:"SQL_DB_APP_DNS_CONNECTION,required"`
	ReconnectTimeout time.Duration `env:"SQL_DB_APP_RECONNECT_TIMEOUT" envDefault:"1s"`
	OpenLimit        int           `env:"SQL_DB_APP_OPEN_LIMIT" envDefault:"20"`
	IdleLimit        int           `env:"SQL_DB_APP_IDLE_LIMIT" envDefault:"5"`
	ConnLife         time.Duration `env:"SQL_DB_APP_CONN_LIFE" envDefault:"5m"`

	ReCommit SQLDbReCommit
}
