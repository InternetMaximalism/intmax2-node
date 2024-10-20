package pgx

import (
	"encoding/json"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"
)

func (p *pgx) CreateCtrlProcessingJobs(name string, options json.RawMessage) (err error) {
	const (
		qOptNull = ` INSERT INTO ctrl_processing_jobs
              (processing_job_name) VALUES ($1)
              ON CONFLICT (processing_job_name)
              DO nothing `

		qOptNotNull = ` INSERT INTO ctrl_processing_jobs
              (processing_job_name, options) VALUES ($1, $2)
              ON CONFLICT (processing_job_name)
              DO nothing `
	)

	if options == nil {
		_, err = p.exec(p.ctx, qOptNull, name)
	} else {
		_, err = p.exec(p.ctx, qOptNotNull, name, options)
	}
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) CtrlProcessingJobs(name string) (*mDBApp.CtrlProcessingJobs, error) {
	const (
		q = ` SELECT processing_job_name, options, created_at, updated_at
              FROM ctrl_processing_jobs
              WHERE processing_job_name = $1 FOR UPDATE SKIP LOCKED LIMIT 1 `
	)

	var ctrlJobs models.CtrlProcessingJobs
	err := errPgx.Err(p.queryRow(p.ctx, q, name).
		Scan(
			&ctrlJobs.ProcessingJobName,
			&ctrlJobs.Options,
			&ctrlJobs.CreatedAt,
			&ctrlJobs.UpdatedAt,
		))
	if err != nil {
		return nil, err
	}

	ctrlJobsDBApp := p.ctrlProcessingJobsToDBApp(&ctrlJobs)

	return &ctrlJobsDBApp, nil
}

func (p *pgx) CtrlProcessingJobsByMaskName(mask string) (*mDBApp.CtrlProcessingJobs, error) {
	const (
		q = ` SELECT processing_job_name, options, created_at, updated_at
              FROM ctrl_processing_jobs
              WHERE processing_job_name ILIKE $1 AND updated_at < NOW()
              ORDER BY updated_at ASC
              FOR UPDATE SKIP LOCKED LIMIT 1 `
	)

	var ctrlJobs models.CtrlProcessingJobs
	err := errPgx.Err(p.queryRow(p.ctx, q, `%`+fmt.Sprintf("%s", mask)+`%`).
		Scan(
			&ctrlJobs.ProcessingJobName,
			&ctrlJobs.Options,
			&ctrlJobs.CreatedAt,
			&ctrlJobs.UpdatedAt,
		))
	if err != nil {
		return nil, err
	}

	ctrlJobsDBApp := p.ctrlProcessingJobsToDBApp(&ctrlJobs)

	return &ctrlJobsDBApp, nil
}

func (p *pgx) UpdatedAtOfCtrlProcessingJobByName(name string, updatedAt time.Time) (err error) {
	const (
		q = ` UPDATE ctrl_processing_jobs SET updated_at = $1 WHERE processing_job_name = $2 `
	)

	_, err = p.exec(p.ctx, q, updatedAt, name)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgx) DeleteCtrlProcessingJobByName(name string) (err error) {
	const (
		q = ` DELETE FROM ctrl_processing_jobs WHERE processing_job_name = $1 `
	)

	_, err = p.exec(p.ctx, q, name)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgx) ctrlProcessingJobsToDBApp(ctrlJobs *models.CtrlProcessingJobs) mDBApp.CtrlProcessingJobs {
	m := mDBApp.CtrlProcessingJobs{
		ProcessingJobName: ctrlJobs.ProcessingJobName,
		Options:           ctrlJobs.Options,
		CreatedAt:         ctrlJobs.CreatedAt,
		UpdatedAt:         ctrlJobs.UpdatedAt,
	}
	if m.Options == nil {
		m.Options = json.RawMessage(`{}`)
	}

	return m
}
