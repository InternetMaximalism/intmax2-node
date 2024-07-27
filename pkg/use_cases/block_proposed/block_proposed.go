package block_proposed

import (
	"context"
	"fmt"
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

	// input.TxTree.
	txTreeMerkleProof := make([]string, len(input.TxTree.TxTreeSiblings))
	for i, v := range input.TxTree.TxTreeSiblings {
		txTreeMerkleProof[i] = v.String()
	}

	const numOfSenders = 128
	resp := block_proposed.UCBlockProposed{
		TxRoot:            input.TxTree.TxTreeRootHash.String(),
		TxTreeMerkleProof: txTreeMerkleProof,
		PublicKeys:        make([]string, numOfSenders),
	}
	fmt.Printf("resp: %v\n", resp.TxRoot)
	fmt.Printf("resp: %v\n", resp.TxTreeMerkleProof)
	fmt.Printf("resp: %v\n", resp.PublicKeys)

	return &resp, nil
}
