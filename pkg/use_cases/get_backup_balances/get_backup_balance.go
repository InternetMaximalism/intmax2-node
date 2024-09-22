package get_backup_balances

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/sql_db/db_app/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_balance.UseCaseGetBackupBalances {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_balance.UCGetBackupBalancesInput,
) (*node.GetBackupBalancesResponse_Data, error) {
	const (
		hName          = "UseCase GetBackupBalances"
		userKey        = "user"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupBalancesInputEmpty)
		return nil, ErrUCGetBackupBalancesInputEmpty
	}

	balances, err := service.GetBackupBalances(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		var ErrGetBackupBalances = errors.New("failed to get backup balances")
		return nil, errors.Join(ErrGetBackupBalances, err)
	}

	for _, balance := range balances {
		fmt.Printf("UseCaseGetBackupBalances balance: %v, %v\n", balance.CreatedAt, balance.EncryptedBalanceData)
	}

	data := node.GetBackupBalancesResponse_Data{
		Balances: generateBackupBalances(balances),
		Meta: &node.GetBackupBalancesResponse_Meta{
			StartBlockNumber: input.StartBlockNumber,
			EndBlockNumber:   input.StartBlockNumber,
		},
	}

	return &data, nil
}

func generateBackupBalances(balances []*models.BackupBalance) []*node.GetBackupBalancesResponse_Balance {
	results := make([]*node.GetBackupBalancesResponse_Balance, 0, len(balances))
	for _, balance := range balances {
		backupBalance := &node.GetBackupBalancesResponse_Balance{
			Id:                    balance.ID,
			UserAddress:           balance.UserAddress,
			EncryptedBalanceProof: balance.EncryptedBalanceProof,
			EncryptedBalanceData:  balance.EncryptedBalanceData,
			EncryptedTxs:          balance.EncryptedTxs,
			EncryptedTransfers:    balance.EncryptedTransfers,
			EncryptedDeposits:     balance.EncryptedDeposits,
			BlockNumber:           balance.BlockNumber,
			Signature:             balance.Signature,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: balance.CreatedAt.Unix(),
				Nanos:   int32(balance.CreatedAt.Nanosecond()),
			},
		}
		fmt.Printf("generateBackupBalances backupBalance: %s, %v\n", backupBalance.Id, backupBalance.EncryptedBalanceData)
		results = append(results, backupBalance)
	}
	return results
}
