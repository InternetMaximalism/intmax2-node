package post_backup_transaction

import (
	"context"
	"encoding/binary"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/open_telemetry"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_transaction"
	"intmax2-node/pkg/use_cases/post_backup_transfer"
	"io"

	"github.com/iden3/go-iden3-crypto/ffg"
	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_transaction.UseCasePostBackupTransaction {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transaction.UCPostBackupTransactionInput,
) (*backup_transaction.UCPostBackupTransaction, error) {
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
		return nil, ErrUCPostBackupTransactionInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		attribute.String(encryptedTxKey, input.EncryptedTx),
	)

	// TODO: Implement backup transaction logic here.

	resp := backup_transaction.UCPostBackupTransaction{
		Message: "Transaction data backup successful.",
	}

	return &resp, nil
}

func MakeTransfers(buf io.Writer, transfers []*intMaxTypes.Transfer) error {
	err := binary.Write(buf, binary.LittleEndian, int64(len(transfers)))
	if err != nil {
		return err
	}

	for _, transfer := range transfers {
		err := post_backup_transfer.WriteTransfer(buf, transfer)
		if err != nil {
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
	finite_field.WriteUint64(buf, uint64(blockNumber))
	finite_field.WriteFixedSizeBytes(buf, encryptedTx, len(encryptedTx))

	return buf.Inner()
}
