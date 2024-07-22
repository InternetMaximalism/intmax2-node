package block_signature

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/block_signature"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct{}

func New() block_signature.UseCaseBlockSignature {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *block_signature.UCBlockSignatureInput,
) (*block_signature.UCBlockSignature, error) {
	const (
		hName        = "UseCase BlockSignature"
		senderKey    = "sender"
		signatureKey = "signature"
		txHashKey    = "tx_hash"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(senderKey, input.Sender),
			attribute.String(signatureKey, input.Signature),
			attribute.String(txHashKey, input.TxHash),
		))
	defer span.End()

	// TODO Verify signature.

	// TODO: Verify enough balance proof by using Balance Validity Prover.

	resp := block_signature.UCBlockSignature{
		Message: "Signature accepted.",
	}

	return &resp, nil
}
