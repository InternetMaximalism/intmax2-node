package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func (p *pgx) CreateCtrlEventBlockNumbersJobs(eventName string) error {
	const (
		q = ` INSERT INTO ctrl_event_block_numbers_jobs
              (event_name) VALUES ($1)
              ON CONFLICT (event_name)
              DO nothing `
	)

	_, err := p.exec(p.ctx, q, eventName)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) CtrlEventBlockNumbersJobs(eventName string) (*mDBApp.CtrlEventBlockNumbersJobs, error) {
	const (
		q = ` SELECT event_name, created_at FROM ctrl_event_block_numbers_jobs
              WHERE event_name = $1 FOR UPDATE SKIP LOCKED LIMIT 1 `
	)

	var ctrlEBnJobs models.CtrlEventBlockNumbersJobs
	err := errPgx.Err(p.queryRow(p.ctx, q, eventName).
		Scan(
			&ctrlEBnJobs.EventName,
			&ctrlEBnJobs.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	ctrlEBnJobsDBApp := p.ctrlEBnJobsToDBApp(&ctrlEBnJobs)

	return &ctrlEBnJobsDBApp, nil
}

func (p *pgx) ctrlEBnJobsToDBApp(ctrlEBnJobs *models.CtrlEventBlockNumbersJobs) mDBApp.CtrlEventBlockNumbersJobs {
	return mDBApp.CtrlEventBlockNumbersJobs{
		EventName: ctrlEBnJobs.EventName,
		CreatedAt: ctrlEBnJobs.CreatedAt,
	}
}
