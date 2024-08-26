package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func (p *pgx) CreateCtrlProcessingJobs(name string) error {
	const (
		q = ` INSERT INTO ctrl_processing_jobs
              (processing_job_name) VALUES ($1)
              ON CONFLICT (processing_job_name)
              DO nothing `
	)

	_, err := p.exec(p.ctx, q, name)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) CtrlProcessingJobs(name string) (*mDBApp.CtrlProcessingJobs, error) {
	const (
		q = ` SELECT processing_job_name, created_at FROM ctrl_processing_jobs
              WHERE processing_job_name = $1 FOR UPDATE SKIP LOCKED LIMIT 1 `
	)

	var ctrlJobs models.CtrlProcessingJobs
	err := errPgx.Err(p.queryRow(p.ctx, q, name).
		Scan(
			&ctrlJobs.ProcessingJobName,
			&ctrlJobs.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	ctrlJobsDBApp := p.ctrlProcessingJobsToDBApp(&ctrlJobs)

	return &ctrlJobsDBApp, nil
}

func (p *pgx) ctrlProcessingJobsToDBApp(ctrlJobs *models.CtrlProcessingJobs) mDBApp.CtrlProcessingJobs {
	return mDBApp.CtrlProcessingJobs{
		ProcessingJobName: ctrlJobs.ProcessingJobName,
		CreatedAt:         ctrlJobs.CreatedAt,
	}
}
