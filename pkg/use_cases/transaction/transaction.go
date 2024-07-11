package transaction

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/open_telemetry"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// uc describes use case
type uc struct {
	cfg   *configs.Config
	dbApp SQLDriverApp
	w     Worker
}

func New(
	cfg *configs.Config,
	dbApp SQLDriverApp,
	w Worker,
) transaction.UseCaseTransaction {
	return &uc{
		cfg:   cfg,
		dbApp: dbApp,
		w:     w,
	}
}

func (u *uc) Do(ctx context.Context, input *transaction.UCTransactionInput) (err error) {
	const (
		hName           = "UseCase Transaction"
		senderKey       = "sender"
		transferHashKey = "transfer_hash"
	)

	if input == nil {
		return errors.New("input is nil")
	}

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
			attribute.String(transferHashKey, input.TransfersHash),
		))
	defer span.End()

	transferData := make([]*intMaxTypes.Transfer, len(input.TransferData))
	for key := range input.TransferData {
		transferData[key] = &intMaxTypes.Transfer{
			Recipient:  input.TransferData[key].DecodeRecipient,
			TokenIndex: uint32(input.TransferData[key].DecodeTokenIndex.Uint64()),
			Amount:     input.TransferData[key].DecodeAmount,
			Salt:       input.TransferData[key].DecodeSalt,
		}
	}

	err = u.w.Receiver(&worker.ReceiverWorker{
		Sender:       input.DecodeSender.ToAddress().String(),
		TransferHash: input.TransfersHash,
		TransferData: transferData,
	})
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrTransferWorkerReceiverFail, err)
	}

	return nil
}
