package get_backup_transaction

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_transaction"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_transaction.UseCaseGetBackupTransaction {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transaction.UCGetBackupTransactionInput,
) (*node.GetBackupTransactionResponse_Data, error) {
	const (
		hName               = "UseCase GetBackupTransaction"
		senderKey           = "sender"
		startBlockNumberKey = "start_block_number"
		limitKey            = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransactionInputEmpty)
		return nil, ErrUCGetBackupTransactionInputEmpty
	}

	span.SetAttributes(
		// attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.Int64(startBlockNumberKey, int64(input.StartBlockNumber)),
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	service.GetBackupTransaction(ctx, u.cfg, u.log, u.db, input)

	data := node.GetBackupTransactionResponse_Data{
		Transactions: genTransaction(),
		Meta: &node.GetBackupTransactionResponse_Meta{
			StartBlockNumber: 0,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func genTransaction() []*node.GetBackupTransactionResponse_Transaction {
	result := make([]*node.GetBackupTransactionResponse_Transaction, 1)
	result[0] = &node.GetBackupTransactionResponse_Transaction{
		EncryptedTx: "0x123",
		BlockNumber: 1000,
	}
	return result
}
