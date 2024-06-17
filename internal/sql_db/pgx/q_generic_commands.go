package pgx

import (
	"context"
	"fmt"
	"time"
)

const (
	SQLDbKey = "sql-db-key"
)

func (p *pgx) Begin(ctx context.Context) (interface{}, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &pgx{
		db:                   p.db,
		tx:                   tx,
		ctx:                  ctx,
		commitAttemptsNumber: p.commitAttemptsNumber,
		commitTimeout:        p.commitTimeout,
	}, nil
}

func (p *pgx) Rollback() {
	if p.tx == nil {
		return
	}

	_ = p.tx.Rollback()
}

func (p *pgx) Commit() error {
	if p.tx == nil {
		return nil
	}

	return p.tx.Commit()
}

func (p *pgx) Exec(ctx context.Context, input interface{}, executor func(sqlDriver interface{}, input interface{}) error) error {
	if value := ctx.Value(SQLDbKey); value != nil {
		if err := executor(value, input); err != nil {
			const msg = "failed with apply executor for %s: %w"
			return fmt.Errorf(msg, SQLDbKey, err)
		}

		return nil
	}

	fn := func(ctx context.Context, input interface{}, executor func(sqlDriver interface{}, input interface{}) error) error {
		bx, err := p.Begin(ctx)
		if err != nil {
			const msg = "failed of begin transaction with pgx driver: %w"
			return fmt.Errorf(msg, err)
		}

		tx, _ := bx.(PGX)
		defer tx.Rollback()

		err = executor(bx, input)
		if err != nil {
			const msg = "failed of apply executor with pgx driver: %w"
			return fmt.Errorf(msg, err)
		}

		err = tx.Commit()
		if err != nil {
			const msg = "failed of commit of executor with pgx driver: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	}

	var attemptsNumber int
	for {
		attemptsNumber++
		err := fn(ctx, input, executor)
		if err != nil && !p.ErrIsRetryable(err) {
			const msg = "failed of checks error os retryable with pgx driver: %w"
			return fmt.Errorf(msg, err)
		}
		if err != nil && p.ErrIsRetryable(err) {
			if attemptsNumber > p.commitAttemptsNumber {
				const msg = "the current attemptsNumber value more than the sql-driver commitAttemptsNumber value with pgx driver: %w"
				return fmt.Errorf(msg, err)
			}
			<-time.After(p.commitTimeout)
			continue
		}
		break
	}

	return nil
}
