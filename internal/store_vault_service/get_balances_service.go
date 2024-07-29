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
	deposits, err := db.GetBackupDeposits("recipient", input.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create get backup depsits from db: %w", err)
	}
	fmt.Println("deposits", deposits)

	transactions, err := db.GetBackupTransactions("sender", input.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create get backup transactions from db: %w", err)
	}
	fmt.Println("transactions", transactions)

	transfers, err := db.GetBackupTransfers("recipient", input.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create get backup transfers from db: %w", err)
	}
	fmt.Println("transfers", transfers)

	balanceMap := map[int]string{
		1: "100",
		2: "200",
	}

	var balances []*backupBalance.TokenBalance
	for index, amount := range balanceMap {
		balances = append(balances, &backupBalance.TokenBalance{TokenIndex: index, Amount: amount})
	}

	return &backupBalance.UCGetBalances{Balances: balances}, nil
}
