package post_backup_balance

import (
	"context"

	"intmax2-node/internal/sql_db/pgx/models"
)

type SQLDriverApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
	BackApp
}

type BackApp interface {
	BackupUserBalance(input *models.BalanceBackup) error
}
