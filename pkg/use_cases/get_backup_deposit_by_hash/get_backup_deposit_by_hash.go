package get_backup_deposit_by_hash

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	getBackupDepositByHash "intmax2-node/internal/use_cases/get_backup_deposit_by_hash"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupDepositByHash.UseCaseGetBackupDepositByHash {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *getBackupDepositByHash.UCGetBackupDepositByHashInput,
) (*node.GetBackupDepositByHashResponse_Data, error) {
	const (
		hName     = "UseCase GetBackupTransactionByHash"
		senderKey = "sender"
		txHashKey = "tx_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransactionByHashInputEmpty)
		return nil, ErrUCGetBackupTransactionByHashInputEmpty
	}

	span.SetAttributes(
		attribute.String(txHashKey, input.DepositHash),
		attribute.String(senderKey, input.Recipient),
	)

	deposit, err := service.GetBackupDepositByHash(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupDepositByHashResponse_Data{
		Deposit: &node.GetBackupDepositByHashResponse_Deposit{
			Id:               deposit.ID,
			Recipient:        deposit.Recipient,
			BlockNumber:      uint64(deposit.BlockNumber),
			EncryptedDeposit: deposit.EncryptedDeposit,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: deposit.CreatedAt.Unix(),
				Nanos:   int32(deposit.CreatedAt.Nanosecond()),
			},
		},
	}

	return &data, nil
}
