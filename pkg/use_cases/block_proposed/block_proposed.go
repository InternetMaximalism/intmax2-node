package block_proposed

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/block_proposed"

	"go.opentelemetry.io/otel/attribute"
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

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCInputEmpty)
		return nil, ErrUCInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.String(txHashKey, input.TxHash),
	)

	txTreeMerkleProof := make([]string, len(input.TxTree.Siblings))
	for i, v := range input.TxTree.Siblings {
		txTreeMerkleProof[i] = v.String()
	}

	// TODO: publicKeys is all sender public keys included in this block.
	senderPublicKeys := make([]string, 1)
	senderPublicKeys[0] = input.DecodeSender.ToAddress().String()

	resp := block_proposed.UCBlockProposed{
		TxRoot:            input.TxTree.RootHash.String(),
		TxTreeMerkleProof: make([]string, len(input.TxTree.Siblings)),
		PublicKeys:        senderPublicKeys,
	}

	for key := range input.TxTree.Siblings {
		resp.TxTreeMerkleProof[key] = input.TxTree.Siblings[key].String()
	}

	return &resp, nil
}
