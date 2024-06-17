package errors

import (
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"strings"

	"database/sql"
)

// Err describes wrapper of gorm errors.
func Err(err error) error {
	if err == nil {
		return nil
	}

	if strings.EqualFold(err.Error(), sql.ErrNoRows.Error()) {
		return errorsDB.ErrNotFound
	}

	const (
		errNotUnique = "SQLSTATE 23505"
	)

	if strings.Contains(err.Error(), errNotUnique) {
		return errorsDB.ErrNotUnique
	}

	return err
}
