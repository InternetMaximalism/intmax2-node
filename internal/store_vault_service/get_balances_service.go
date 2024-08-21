package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
)

func GetBalances(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupBalance.UCGetBalancesInput,
) (*backupBalance.UCGetBalances, error) {
	// TODO: get these data concurrently
	deposits, err := db.GetBackupDeposits("recipient", input.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create get backup depsits from db: %w", err)
	}

	transactions, err := db.GetBackupTransactions("sender", input.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create get backup transactions from db: %w", err)
	}

	transfers, err := db.GetBackupTransfers("recipient", input.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create get backup transfers from db: %w", err)
	}

	resDeposits := make([]*backupBalance.BackupDeposit, len(deposits))
	for i, deposit := range deposits {
		resDeposits[i] = &backupBalance.BackupDeposit{
			Recipient:        deposit.Recipient,
			EncryptedDeposit: deposit.EncryptedDeposit,
			BlockNumber:      deposit.BlockNumber,
			CreatedAt:        deposit.CreatedAt,
		}
	}

	resTransfers := make([]*backupBalance.BackupTransfer, len(transfers))
	for i, transfer := range transfers {
		resTransfers[i] = &backupBalance.BackupTransfer{
			EncryptedTransfer: transfer.EncryptedTransfer,
			Recipient:         transfer.Recipient,
			BlockNumber:       transfer.BlockNumber,
			CreatedAt:         transfer.CreatedAt,
		}
	}
	resTransactions := make([]*backupBalance.BackupTransaction, len(transactions))
	for i, transaction := range transactions {
		resTransactions[i] = &backupBalance.BackupTransaction{
			Sender:      transaction.Sender,
			EncryptedTx: transaction.EncryptedTx,
			BlockNumber: uint64(transaction.BlockNumber),
			CreatedAt:   transaction.CreatedAt,
		}
	}

	return &backupBalance.UCGetBalances{
		Deposits:     resDeposits,
		Transactions: resTransactions,
		Transfers:    resTransfers,
	}, nil
}
