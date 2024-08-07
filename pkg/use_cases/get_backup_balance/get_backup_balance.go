package get_backup_balance

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_balance"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_balance.UseCaseGetBackupBalance {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_balance.UCGetBackupBalanceInput,
) (*node.GetBackupBalanceResponse_Data, error) {
	const (
		hName          = "UseCase GetBackupBalance"
		userKey        = "user"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupBalanceInputEmpty)
		return nil, ErrUCGetBackupBalanceInputEmpty
	}

	span.SetAttributes(
	// attribute.String(userKey, input.DecodeUser.ToAddress().String()),
	// attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
	)

	// TODO: Implement backup balance get logic here.
	service.GetBackupBalance(ctx, u.cfg, u.log, u.db, input)

	data := node.GetBackupBalanceResponse_Data{
		Transactions: genTransaction(),
		Meta: &node.GetBackupBalanceResponse_Meta{
			StartBlockNumber: 0,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func genTransaction() []*node.GetBackupBalanceResponse_Transaction {
	result := make([]*node.GetBackupBalanceResponse_Transaction, 1)
	result[0] = &node.GetBackupBalanceResponse_Transaction{
		EncryptedTx: "0x123",
		BlockNumber: 1000,
	}
	return result
}
