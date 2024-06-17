package db_app

import (
	"context"
	"fmt"
	"time"
)

const (
	SQLDbKey = "sql-db-key"
)

func (p *dbApp) Begin(ctx context.Context) (interface{}, error) {
	tx, err := p.SQLDb.Begin(ctx)
	if err != nil {
		return nil, err
	}
	db, _ := tx.(SQLDb)

	return &dbApp{
		SQLDb:                db,
		driver:               p.driver,
		commitAttemptsNumber: p.commitAttemptsNumber,
		commitTimeout:        p.commitTimeout,
	}, nil
}

func (p *dbApp) Exec(ctx context.Context, input interface{}, executor func(sqlDriver interface{}, input interface{}) error) error {
	fn := func(ctx context.Context, input interface{}, executor func(sqlDriver interface{}, input interface{}) error) error {
		bx, err := p.Begin(ctx)
		if err != nil {
			const msg = "failed of begin transaction with postgres driver: %w"
			return fmt.Errorf(msg, err)
		}

		tx, _ := bx.(SQLDb)
		defer tx.Rollback()

		err = p.SQLDb.Exec(context.WithValue(ctx, SQLDbKey, tx), input, executor)
		if err != nil {
			const msg = "failed of apply executor with postgres driver: %w"
			return fmt.Errorf(msg, err)
		}

		err = tx.Commit()
		if err != nil {
			const msg = "failed of commit of executor with postgres driver: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	}

	var attemptsNumber int
	for {
		attemptsNumber++
		err := fn(ctx, input, executor)
		if err != nil && !p.ErrIsRetryable(err) {
			const msg = "failed of checks error os retryable with postgres driver: %w"
			return fmt.Errorf(msg, err)
		}
		if err != nil && p.ErrIsRetryable(err) {
			if attemptsNumber > p.commitAttemptsNumber {
				const msg = "the current attemptsNumber value more than the sql-driver commitAttemptsNumber value with postgres driver: %w"
				return fmt.Errorf(msg, err)
			}
			<-time.After(p.commitTimeout)
			continue
		}
		break
	}

	return nil
}
