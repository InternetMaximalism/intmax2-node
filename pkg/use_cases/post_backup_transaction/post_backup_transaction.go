package post_backup_transaction

import (
	"context"
	"encoding/binary"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/store_vault_service"
	intMaxTypes "intmax2-node/internal/types"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	"intmax2-node/pkg/use_cases/post_backup_transfer"
	"io"

	"github.com/iden3/go-iden3-crypto/ffg"
	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCasePostBackupTransaction {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backupTransaction.UCPostBackupTransactionInput,
) error {
	const (
		hName          = "UseCase PostBackupTransaction"
		senderKey      = "sender"
		blockNumberKey = "block_number"
		encryptedTxKey = "encrypted_tx"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupTransactionInputEmpty)
		return ErrUCPostBackupTransactionInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.Sender),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		attribute.String(encryptedTxKey, input.EncryptedTx),
	)

	err := service.PostBackupTransaction(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return fmt.Errorf("failed to post backup transfer: %w", err)
	}

	return nil
}

func WriteTransfers(buf io.Writer, transfers []*intMaxTypes.Transfer) error {
	if err := binary.Write(buf, binary.LittleEndian, int64(len(transfers))); err != nil {
		return err
	}

	for _, transfer := range transfers {
		if err := post_backup_transfer.WriteTransfer(buf, transfer); err != nil {
			return err
		}
	}

	return nil
}

func MakeMessage(senderAddress intMaxAcc.Address, blockNumber uint32, encryptedTx []byte) []ffg.Element {
	const (
		int4Key  = 4
		int8Key  = 8
		int32Key = 32
	)
	bufferSize := int8Key + 1 + len(encryptedTx)/int4Key + 1
	buf := finite_field.NewBuffer(make([]ffg.Element, bufferSize))
	finite_field.WriteFixedSizeBytes(buf, senderAddress.Bytes(), int32Key)
	err := finite_field.WriteUint64(buf, uint64(blockNumber))
	// blockNumber is uint32, so it should be safe to cast to uint64
	if err != nil {
		panic(err)
	}
	finite_field.WriteFixedSizeBytes(buf, encryptedTx, len(encryptedTx))

	return buf.Inner()
}
