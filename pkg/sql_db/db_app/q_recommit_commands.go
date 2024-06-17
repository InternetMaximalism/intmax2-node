package db_app

import "errors"

func (p *dbApp) ErrIsRetryable(err error) bool {
	code := p.errCode(err)
	return code == "CR000" || code == "40001"
}

// errWithSQLState is implemented by pgx (pgconn.PgError) and lib/pq
type errWithSQLState interface {
	SQLState() string
}

func (p *dbApp) errCode(err error) string {
	var sqlErr errWithSQLState
	if errors.As(err, &sqlErr) {
		return sqlErr.SQLState()
	}

	return ""
}
