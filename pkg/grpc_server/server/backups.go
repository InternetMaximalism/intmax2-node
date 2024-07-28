package server

import (
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	backupBalanceService "intmax2-node/pkg/use_cases/post_backup_balance"
)

//go:generate mockgen -destination=mock_backups_test.go -package=server_test -source=backups.go

type Backups interface {
	BackupBalance(dbApp SQLDriverApp) backupBalance.UseCasePostBackupBalance
}

type backups struct{}

func NewBackups() Backups {
	return &backups{}
}

func (b *backups) BackupBalance(dbApp SQLDriverApp) backupBalance.UseCasePostBackupBalance {
	return backupBalanceService.New(dbApp)
}
