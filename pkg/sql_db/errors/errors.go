package errors

import "errors"

// ErrSQLDriverNameInvalid error: sql driver name must be valid.
var ErrSQLDriverNameInvalid = errors.New("sql driver name must be valid")

// ErrSQLDriverLoad error: sql driver load invalid.
var ErrSQLDriverLoad = errors.New("sql driver load invalid")

// ErrSQLDriverDisabled error: sql driver disabled.
var ErrSQLDriverDisabled = errors.New("sql driver disabled")

// ErrNotFound error: not found.
var ErrNotFound = errors.New("not found")

// ErrNotUnique error: not unique.
var ErrNotUnique = errors.New("not unique")
