package withdrawal_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	postWithdrawalsByHashes "intmax2-node/internal/use_cases/post_withdrawals_by_hashes"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func PostWithdrawalsByHashes(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *postWithdrawalsByHashes.UCPostWithdrawalsByHashesInput,
) (*[]mDBApp.Withdrawal, error) {
	withdrawals, err := db.WithdrawalsByHashes(input.TransferHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals by hashes: %w", err)
	}
	return withdrawals, nil
}
