-- +migrate Up

CREATE TABLE ctrl_processing_jobs (
    processing_job_name varchar(255) not null unique,
    created_at timestamptz not null default now(),
    PRIMARY KEY (processing_job_name)
);

-- +migrate Down

DROP TABLE ctrl_processing_jobs;
