package transaction

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	w   Worker
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	w Worker,
) transaction.UseCaseTransaction {
	return &uc{
		cfg: cfg,
		log: log,
		w:   w,
	}
}

func (u *uc) Do(ctx context.Context, input *transaction.UCTransactionInput) (err error) {
	const (
		hName           = "UseCase Transaction"
		senderKey       = "sender"
		transferHashKey = "transfer_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCInputEmpty)
		return ErrUCInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.String(transferHashKey, input.TransfersHash),
	)

	// TODO: check 0.1 ETH with Rollup contract

	/**
	 * // NOTE: `TransferData` does not need to be sent in the request
	 * transferData := make([]*intMaxTypes.Transfer, len(input.TransferData))
	 * for key := range input.TransferData {
	 * 	transferData[key] = &intMaxTypes.Transfer{
	 *		Recipient:  input.TransferData[key].DecodeRecipient,
	 *		TokenIndex: uint32(input.TransferData[key].DecodeTokenIndex.Uint64()),
	 *		Amount:     input.TransferData[key].DecodeAmount,
	 *		Salt:       input.TransferData[key].DecodeSalt,
	 *	}
	 * }
	 */

	err = u.w.Receiver(&worker.ReceiverWorker{
		Sender:        input.DecodeSender.ToAddress().String(),
		Nonce:         input.Nonce,
		TransfersHash: input.TransfersHash,
	})
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrTransferWorkerReceiverFail, err)
	}

	return nil
}
