package db_app

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/sql_db"
	"intmax2-node/internal/sql_db/pgx"
	driver "intmax2-node/pkg/sql_db/errors"
	"time"
)

type dbApp struct {
	SQLDb
	driver               string
	commitAttemptsNumber int
	commitTimeout        time.Duration
}

func New(ctx context.Context, log logger.Logger, cfg *configs.SQLDb) (SQLDb, error) {
	if cfg.DriverName == sql_db.Pgx {
		db, err := pgx.New(ctx, log, &pgx.Config{
			DNSConnection:          cfg.DNSConnection,
			ReconnectTimeout:       cfg.ReconnectTimeout,
			OpenLimit:              cfg.OpenLimit,
			IdleLimit:              cfg.IdleLimit,
			ConnLife:               cfg.ConnLife,
			ReCommitAttemptsNumber: cfg.ReCommit.AttemptsNumber,
			ReCommitTimeout:        cfg.ReCommit.Timeout,
		})
		if err != nil {
			return nil, errors.Join(driver.ErrSQLDriverLoad, err)
		}
		return &dbApp{
			SQLDb:                db,
			driver:               cfg.DriverName,
			commitAttemptsNumber: cfg.ReCommit.AttemptsNumber,
			commitTimeout:        cfg.ReCommit.Timeout,
		}, nil
	}

	return nil, driver.ErrSQLDriverNameInvalid
}
