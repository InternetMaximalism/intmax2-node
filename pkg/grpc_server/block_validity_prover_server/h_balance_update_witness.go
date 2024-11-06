package block_validity_prover_server

import (
	"context"
	"errors"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	blockValidityProverBalanceUpdateWitness "intmax2-node/internal/use_cases/block_validity_prover_balance_update_witness"
	"intmax2-node/pkg/grpc_server/utils"
	ucBlockValidityProverBalanceUpdateWitness "intmax2-node/pkg/use_cases/block_validity_prover_balance_update_witness"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) BalanceUpdateWitness(
	ctx context.Context,
	req *node.BalanceUpdateWitnessRequest,
) (*node.BalanceUpdateWitnessResponse, error) {
	resp := node.BalanceUpdateWitnessResponse{
		Data: &node.BalanceUpdateWitnessResponse_Data{},
	}

	const (
		hName      = "Handler BalanceUpdateWitness"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
		User:               req.User,
		CurrentBlockNumber: int64(req.CurrentBlockNumber),
		TargetBlockNumber:  int64(req.TargetBlockNumber),
		IsPrevAccountTree:  req.IsPrevAccountTree,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness
	info, err = s.Commands().BlockValidityProverBalanceUpdateWitness(s.config, s.log, s.bvs).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		switch {
		case errors.Is(err, ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberLessThenTargetBlockNumber):
			input.IsTargetBlockNumberMoreThenCurrentBlockNumber = true
			input.IsCurrentBlockNumberLessThenTargetBlockNumber = true
			err = input.Valid()
			return &resp, utils.BadRequest(spanCtx, err)
		case errors.Is(err, ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberInvalid):
			input.IsInvalidCurrentBlockNumber = true
			err = input.Valid()
			return &resp, utils.BadRequest(spanCtx, err)
		case errors.Is(err, ucBlockValidityProverBalanceUpdateWitness.ErrTargetBlockNumberInvalid):
			input.IsInvalidTargetBlockNumber = true
			err = input.Valid()
			return &resp, utils.BadRequest(spanCtx, err)
		case errors.Is(err, ucBlockValidityProverBalanceUpdateWitness.ErrPublicKeyFromIntMaxAccFail):
			input.User = "0x0000000000000000000000000000000000000000000000000000000000000000"
			err = input.Valid()
			return &resp, utils.BadRequest(spanCtx, err)
		default:
			const msg = "failed to fetch balance update witness: %+v"
			return &resp, utils.Internal(spanCtx, s.log, msg, err)
		}
	}

	resp.Success = true
	resp.Data.IsPrevAccountTree = req.IsPrevAccountTree
	resp.Data.ValidityProof = info.ValidityProof

	resp.Data.BlockMerkleProof = make([]string, len(info.BlockMerkleProof.Siblings))
	for key := range info.BlockMerkleProof.Siblings {
		resp.Data.BlockMerkleProof[key] = info.BlockMerkleProof.Siblings[key].String()
	}

	if info.AccountMembershipProof != nil {
		resp.Data.AccountMembershipProof = &node.BalanceUpdateWitnessResponse_AccountMembershipProof{
			IsIncluded: info.AccountMembershipProof.IsIncluded,
			Leaf: &node.BalanceUpdateWitnessResponse_Leaf{
				Key:       info.AccountMembershipProof.Leaf.Key.String(),
				Value:     uint32(info.AccountMembershipProof.Leaf.Value),
				NextIndex: uint32(info.AccountMembershipProof.Leaf.NextIndex),
				NextKey:   info.AccountMembershipProof.Leaf.NextKey.String(),
			},
		}

		resp.Data.AccountMembershipProof.LeafProof = make([]string, len(info.AccountMembershipProof.LeafProof.Siblings))
		for key := range info.AccountMembershipProof.LeafProof.Siblings {
			resp.Data.AccountMembershipProof.LeafProof[key] = info.AccountMembershipProof.LeafProof.Siblings[key].String()
		}
	}

	return &resp, utils.OK(spanCtx)
}
