package block_validity_prover_server

import (
	"context"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	blockValidityProverBlockValidityPublicInputs "intmax2-node/internal/use_cases/block_validity_prover_block_validity_public_inputs"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) BlockValidityPublicInputs(
	ctx context.Context,
	req *node.BlockValidityPublicInputsRequest,
) (*node.BlockValidityPublicInputsResponse, error) {
	resp := node.BlockValidityPublicInputsResponse{}

	const (
		hName      = "Handler BlockValidityPublicInputs"
		requestKey = "request"
		actionKey  = "action"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput{
		BlockNumber: int64(req.BlockNumber),
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs
	info, err = s.commands.BlockValidityProverBlockValidityPublicInputs(s.config, s.log, s.bvs).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get validity public inputs by block number"
		s.log.WithFields(logger.Fields{
			actionKey:  hName,
			requestKey: req.String(),
		}).WithError(err).Warnf(msg)
	}

	if info != nil &&
		info.ValidityPublicInputs != nil &&
		info.ValidityPublicInputs.PublicState != nil &&
		info.ValidityPublicInputs.PublicState.BlockTreeRoot != nil &&
		info.ValidityPublicInputs.PublicState.PrevAccountTreeRoot != nil &&
		info.ValidityPublicInputs.PublicState.AccountTreeRoot != nil &&
		info.ValidityPublicInputs.SenderTreeRoot != nil {
		resp.Success = true
		resp.Data = &node.BlockValidityPublicInputsResponse_Data{
			ValidityPublicInputs: &node.BlockValidityPublicInputsResponse_Data_ValidityPublicInputs{
				TxTreeRoot:     info.ValidityPublicInputs.TxTreeRoot.Hex(),
				SenderTreeRoot: info.ValidityPublicInputs.SenderTreeRoot.String(),
				IsValidBlock:   info.ValidityPublicInputs.IsValidBlock,
				PublicState: &node.BlockValidityPublicInputsResponse_Data_PublicState{
					BlockTreeRoot:       info.ValidityPublicInputs.PublicState.BlockTreeRoot.String(),
					PrevAccountTreeRoot: info.ValidityPublicInputs.PublicState.PrevAccountTreeRoot.String(),
					AccountTreeRoot:     info.ValidityPublicInputs.PublicState.AccountTreeRoot.String(),
					DepositTreeRoot:     info.ValidityPublicInputs.PublicState.DepositTreeRoot.String(),
					BlockHash:           info.ValidityPublicInputs.PublicState.BlockHash.String(),
					BlockNumber:         info.ValidityPublicInputs.PublicState.BlockNumber,
				},
			},
			Senders: make([]*node.BlockValidityPublicInputsResponse_Data_Sender, len(info.Sender)),
		}
		for key := range info.Sender {
			resp.Data.Senders[key] = &node.BlockValidityPublicInputsResponse_Data_Sender{
				PublicKey: info.Sender[key].Hash().String(),
				IsValid:   info.Sender[key].IsValid,
			}
		}
	}

	return &resp, utils.OK(spanCtx)
}
