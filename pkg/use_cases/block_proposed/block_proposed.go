package block_proposed

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/block_proposed"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// uc describes use case
type uc struct{}

func New() block_proposed.UseCaseBlockProposed {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *block_proposed.UCBlockProposedInput,
) (*block_proposed.UCBlockProposed, error) {
	const (
		hName     = "UseCase BlockProposed"
		senderKey = "sender"
		txHashKey = "tx_hash"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
			attribute.String(txHashKey, input.TxHash),
		))
	defer span.End()

	resp := block_proposed.UCBlockProposed{
		TxRoot:            input.TxTree.TxTreeHash.String(),
		TxTreeMerkleProof: make([]string, len(input.TxTree.SenderTransfers)),
	}

	for key := range input.TxTree.SenderTransfers {
		resp.TxTreeMerkleProof[key] = input.TxTree.SenderTransfers[key].TxTreeLeaveHash.String()
	}

	return &resp, nil
}
