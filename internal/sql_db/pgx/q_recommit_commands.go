package pgx

import "errors"

func (p *pgx) ErrIsRetryable(err error) bool {
	code := p.errCode(err)
	return code == "CR000" || code == "40001"
}

// errWithSQLState is implemented by pgx (pgconn.PgError) and lib/pq
type errWithSQLState interface {
	SQLState() string
}

func (p *pgx) errCode(err error) string {
	var sqlErr errWithSQLState
	if errors.As(err, &sqlErr) {
		return sqlErr.SQLState()
	}

	return ""
}
