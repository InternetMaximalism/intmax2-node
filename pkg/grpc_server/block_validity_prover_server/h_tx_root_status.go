package block_validity_prover_server

import (
	"context"
	"encoding/hex"
	"fmt"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	blockValidityProverTxRootStatus "intmax2-node/internal/use_cases/block_validity_prover_tx_root_status"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) TxRootStatus(
	ctx context.Context,
	req *node.TxRootStatusRequest,
) (*node.TxRootStatusResponse, error) {
	resp := node.TxRootStatusResponse{}

	const (
		hName      = "Handler TxRootStatus"
		requestKey = "request"
		actionKey  = "action"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput{
		TxRoots: req.TxRoots,
	}

	err := input.Valid(s.config)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info map[string]*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus
	if input.ConvertTxRoot != nil {
		info, err = s.commands.BlockValidityProverTxRootStatus(s.config, s.log, s.bvs).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get block status by tx root"
			s.log.WithFields(logger.Fields{
				actionKey:  hName,
				requestKey: req.String(),
			}).WithError(err).Warnf(msg)

			resp.Data = &node.TxRootStatusResponse_Data{
				Blocks: make([]*node.TxRootStatusResponse_Data_Block, 0),
				Errors: make([]*node.TxRootStatusResponse_Data_Error, 0),
			}

			return &resp, utils.OK(spanCtx)
		}
	}

	resp.Success = true
	resp.Data = &node.TxRootStatusResponse_Data{}
	for key := range input.ConvertTxRoot {
		v, ok := info[input.ConvertTxRoot[key].String()]
		if !ok {
			if resp.Data.Errors == nil {
				resp.Data.Errors = make([]*node.TxRootStatusResponse_Data_Error, 0)
			}

			resp.Data.Errors = append(resp.Data.Errors, &node.TxRootStatusResponse_Data_Error{
				TxRoot:  input.ConvertTxRoot[key].String(),
				Message: blockValidityProverTxRootStatus.ErrTxRootNotExisting.Error(),
			})

			continue
		}

		if resp.Data.Blocks == nil {
			resp.Data.Blocks = make([]*node.TxRootStatusResponse_Data_Block, 0)
		}

		block := node.TxRootStatusResponse_Data_Block{
			BlockType:           node.BlockType_NON_REGISTRATION,
			TxRoot:              v.TxTreeRoot.String(),
			PrevBlockHash:       v.PrevBlockHash.String(),
			BlockNumber:         v.BlockNumber,
			DepositRoot:         v.DepositRoot.String(),
			SignatureHash:       v.SignatureHash.String(),
			MessagePoint:        fmt.Sprintf("0x%s", hex.EncodeToString(v.MessagePoint.Marshal())),
			AggregatedPublicKey: fmt.Sprintf("0x%s", hex.EncodeToString(v.AggregatedPublicKey.Marshal())),
			AggregatedSignature: fmt.Sprintf("0x%s", hex.EncodeToString(v.AggregatedSignature.Marshal())),
			Senders:             make([]*node.TxRootStatusResponse_Data_Block_Sender, len(v.Senders)),
		}

		if v.IsRegistrationBlock {
			block.BlockType = node.BlockType_REGISTRATION
		}

		for sIndex := range v.Senders {
			block.Senders[sIndex] = &node.TxRootStatusResponse_Data_Block_Sender{
				PublicKey: v.Senders[sIndex].PublicKey.ToAddress().String(),
				AccountId: uint32(v.Senders[sIndex].AccountID),
				IsSigned:  v.Senders[sIndex].IsSigned,
			}
		}

		resp.Data.Blocks = append(resp.Data.Blocks, &block)
	}

	for key := range input.TxRootErrors {
		if resp.Data.Errors == nil {
			resp.Data.Errors = make([]*node.TxRootStatusResponse_Data_Error, 0)
		}

		resp.Data.Errors = append(resp.Data.Errors, &node.TxRootStatusResponse_Data_Error{
			TxRoot:  key,
			Message: input.TxRootErrors[key].Message,
		})
	}

	return &resp, utils.OK(spanCtx)
}
