-- +migrate Up

ALTER TABLE ctrl_processing_jobs
    ADD COLUMN options jsonb not null default '{}'::jsonb;

ALTER TABLE ctrl_processing_jobs
    ADD COLUMN updated_at timestamptz not null default now();

-- +migrate Down

ALTER TABLE ctrl_processing_jobs DROP COLUMN options;
ALTER TABLE ctrl_processing_jobs DROP COLUMN updated_at;
