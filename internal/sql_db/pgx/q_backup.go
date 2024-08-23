package pgx

import (
	"database/sql"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
)

func (p *pgx) createBackupEntry(query string, args ...interface{}) error {
	_, err := p.exec(
		p.ctx,
		query,
		args...,
	)
	return errPgx.Err(err)
}

func (p *pgx) getBackupEntries(query string, value interface{}, scanFunc func(*sql.Rows) error) error {
	rows, err := p.query(p.ctx, query, value)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		err = scanFunc(rows)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}
