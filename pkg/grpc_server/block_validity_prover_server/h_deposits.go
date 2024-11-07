package block_validity_prover_server

import (
	"context"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	blockValidityProverDeposits "intmax2-node/internal/use_cases/block_validity_prover_deposits"
	"intmax2-node/pkg/grpc_server/utils"

	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) Deposits(
	ctx context.Context,
	req *node.DepositsRequest,
) (*node.DepositsResponse, error) {
	resp := node.DepositsResponse{}

	const (
		hName      = "Handler Deposits"
		requestKey = "request"
		actionKey  = "action"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := blockValidityProverDeposits.UCBlockValidityProverDepositsInput{
		DepositHashes: req.DepositHashes,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info []*blockValidityProverDeposits.UCBlockValidityProverDeposits
	info, err = s.commands.BlockValidityProverDeposits(s.config, s.log, s.bvs).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get validity public inputs by block number"
		s.log.WithFields(logger.Fields{
			actionKey:  hName,
			requestKey: req.String(),
		}).WithError(err).Warnf(msg)

		return &resp, utils.OK(spanCtx)
	}

	resp.Success = true
	resp.Data = &node.DepositsResponse_Data{
		Deposits: make([]*node.DepositsResponse_Data_Deposit, len(info)),
	}

	for key := range info {
		resp.Data.Deposits[key] = &node.DepositsResponse_Data_Deposit{
			DepositHash:    info[key].DepositHash.String(),
			DepositId:      info[key].DepositId,
			IsSynchronized: info[key].IsSynchronized,
			From:           info[key].Sender,
		}
		if info[key].DepositIndex != nil {
			resp.Data.Deposits[key].DepositIndex = *info[key].DepositIndex
		}
		if info[key].BlockNumber != nil {
			resp.Data.Deposits[key].BlockNumber = *info[key].BlockNumber
		}
		if info[key].DepositLeaf != nil {
			resp.Data.Deposits[key].DepositLeaf = &node.DepositsResponse_Data_DepositLeaf{
				RecipientSaltHash: common.Hash(info[key].DepositLeaf.RecipientSaltHash).String(),
				TokenIndex:        info[key].DepositLeaf.TokenIndex,
				Amount:            info[key].DepositLeaf.Amount.String(),
			}
		}
	}

	return &resp, utils.OK(spanCtx)
}
