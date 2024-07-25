package backup_service

import (
	"context"

	"intmax2-node/internal/sql_db/pgx/models"
)

type SQLDriverApp interface {
	GenericCommandsApp
	BalanceBackupApp
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type BalanceBackupApp interface {
	BackupUserBalance(input *models.BalanceBackup) error
}
