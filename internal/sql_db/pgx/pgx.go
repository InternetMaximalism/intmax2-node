package pgx

import (
	"context"
	"database/sql"
	"intmax2-node/internal/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	findU0    = `\u0000`
	replaceU0 = `\\*u*0*0*0*0`
)

type pgx struct {
	db                   *sql.DB
	tx                   *sql.Tx
	ctx                  context.Context
	commitAttemptsNumber int
	commitTimeout        time.Duration
}

func New(ctx context.Context, log logger.Logger, cfg *Config) (PGX, error) {
	db, err := sql.Open("pgx", cfg.DNSConnection)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.OpenLimit)
	db.SetMaxIdleConns(cfg.IdleLimit)
	db.SetConnMaxIdleTime(cfg.ReconnectTimeout)
	db.SetConnMaxLifetime(cfg.ConnLife)

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return &pgx{
		db:                   db,
		ctx:                  ctx,
		commitAttemptsNumber: cfg.ReCommitAttemptsNumber,
		commitTimeout:        cfg.ReCommitTimeout,
	}, nil
}

func (p *pgx) queryRow(ctx context.Context, q string, args ...any) *sql.Row { // nolint:unused
	if p.tx != nil {
		return p.tx.QueryRowContext(ctx, q, args...)
	}

	return p.db.QueryRowContext(ctx, q, args...)
}

func (p *pgx) query(ctx context.Context, q string, args ...any) (*sql.Rows, error) { // nolint:unused
	if p.tx != nil {
		return p.tx.QueryContext(ctx, q, args...)
	}

	return p.db.QueryContext(ctx, q, args...)
}

func (p *pgx) exec( // nolint:unused
	ctx context.Context,
	q string,
	arguments ...any,
) (sql.Result, error) {
	if p.tx != nil {
		return p.tx.ExecContext(ctx, q, arguments...)
	}

	return p.db.ExecContext(ctx, q, arguments...)
}
