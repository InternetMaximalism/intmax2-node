package post_backup_transfer

import (
	"context"
	"encoding/binary"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/store_vault_service"
	intMaxTypes "intmax2-node/internal/types"
	postBackupTransfer "intmax2-node/internal/use_cases/post_backup_transfer"
	"io"

	"go.opentelemetry.io/otel/attribute"
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
) postBackupTransfer.UseCasePostBackupTransfer {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *postBackupTransfer.UCPostBackupTransferInput,
) error {
	const (
		hName           = "UseCase PostBackupTransfer"
		transferHashKey = "transfer_hash"
		recipientKey    = "recipient"
		blockNumberKey  = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupTransferInputEmpty)
		return ErrUCPostBackupTransferInputEmpty
	}

	span.SetAttributes(
		attribute.String(transferHashKey, input.TransferHash),
		attribute.String(recipientKey, input.Recipient),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
	)

	err := service.PostBackupTransfer(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return fmt.Errorf("failed to post backup transfer: %w", err)
	}

	return nil
}

func WriteTransfer(buf io.Writer, transfer *intMaxTypes.Transfer) error {
	_, err := buf.Write(transfer.Recipient.Marshal())
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, transfer.TokenIndex)
	if err != nil {
		return err
	}
	_, err = buf.Write(transfer.Amount.Bytes())
	if err != nil {
		return err
	}
	_, err = buf.Write(transfer.Salt.Marshal())
	if err != nil {
		return err
	}

	return nil
}
