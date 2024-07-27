package block_signature

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/block_signature"
	"intmax2-node/internal/worker"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct {
	cfg *configs.Config
	w   Worker
}

func New(
	cfg *configs.Config,
	w Worker,
) block_signature.UseCaseBlockSignature {
	return &uc{
		cfg: cfg,
		w:   w,
	}
}

func (u *uc) Do(
	ctx context.Context, input *block_signature.UCBlockSignatureInput,
) (err error) {
	const (
		hName        = "UseCase BlockSignature"
		senderKey    = "sender"
		signatureKey = "signature"
		txHashKey    = "tx_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(senderKey, input.Sender),
			attribute.String(signatureKey, input.Signature),
			attribute.String(txHashKey, input.TxHash),
		))
	defer span.End()

	// TODO Verify signature.

	// TODO: Verify enough balance proof by using Balance Validity Prover.

	err = u.w.SignTxTreeByAvailableFile(input.Signature, &worker.TransactionHashesWithSenderAndFile{
		Sender: input.TxInfo.Sender,
		File:   input.TxInfo.File,
	})
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrSignTxTreeByAvailableFileFail, err)
	}

	return nil
}
