package get_backup_transactions

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	getBackupTransaction "intmax2-node/internal/use_cases/get_backup_transactions"
	"intmax2-node/pkg/sql_db/db_app/models"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) getBackupTransaction.UseCaseGetBackupTransactions {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *getBackupTransaction.UCGetBackupTransactionsInput,
) (*node.GetBackupTransactionsResponse_Data, error) {
	const (
		hName               = "UseCase GetBackupTransactions"
		senderKey           = "sender"
		startBlockNumberKey = "start_block_number"
		limitKey            = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransactionsInputEmpty)
		return nil, ErrUCGetBackupTransactionsInputEmpty
	}

	span.SetAttributes(
		attribute.Int64(startBlockNumberKey, int64(input.StartBlockNumber)),
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	transactions, err := service.GetBackupTransactions(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupTransactionsResponse_Data{
		Transactions: generateBackupTransaction(transactions),
		Meta: &node.GetBackupTransactionsResponse_Meta{
			StartBlockNumber: input.StartBlockNumber,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func generateBackupTransaction(transactions []*models.BackupTransaction) []*node.GetBackupTransactionsResponse_Transaction {
	results := make([]*node.GetBackupTransactionsResponse_Transaction, 0, len(transactions))
	for key := range transactions {
		backupTransaction := &node.GetBackupTransactionsResponse_Transaction{
			Id:          transactions[key].ID,
			Sender:      transactions[key].Sender,
			Signature:   transactions[key].Signature,
			BlockNumber: uint64(transactions[key].BlockNumber),
			EncryptedTx: transactions[key].EncryptedTx,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transactions[key].CreatedAt.Unix(),
				Nanos:   int32(transactions[key].CreatedAt.Nanosecond()),
			},
		}
		results = append(results, backupTransaction)
	}
	return results
}
