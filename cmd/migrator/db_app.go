package migrator

import "context"

type SQLDriverApp interface {
	Migrator(ctx context.Context, command string) (step int, err error)
}
