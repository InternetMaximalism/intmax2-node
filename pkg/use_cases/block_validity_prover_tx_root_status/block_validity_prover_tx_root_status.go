package block_validity_prover_tx_root_status

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverTxRootStatus "intmax2-node/internal/use_cases/block_validity_prover_tx_root_status"

	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel/attribute"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
	bvs BlockValidityService
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) ucBlockValidityProverTxRootStatus.UseCaseBlockValidityProverTxRootStatus {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput,
) (map[string]*ucBlockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus, error) {
	const (
		hName     = "UseCase BlockValidityProverTxRootStatus"
		txRootKey = "tx_root"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCBlockValidityProverTxRootStatusInputEmpty)
		return nil, ErrUCBlockValidityProverTxRootStatusInputEmpty
	}

	span.SetAttributes(
		attribute.StringSlice(txRootKey, input.TxRoot),
	)

	list, err := u.bvs.AuxInfoListFromBlockContentByTxRoot(input.ConvertTxRoot...)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	result := make(map[string]*ucBlockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus)

	for key := range list {
		var txTree common.Hash
		txTree.SetBytes(list[key].BlockContent.TxTreeRoot[:])
		result[txTree.String()] = &ucBlockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus{
			IsRegistrationBlock: list[key].BlockContent.IsRegistrationBlock,
			TxTreeRoot:          txTree,
			PrevBlockHash:       list[key].PostedBlock.PrevBlockHash,
			BlockNumber:         list[key].PostedBlock.BlockNumber,
			DepositRoot:         list[key].PostedBlock.DepositRoot,
			SignatureHash:       list[key].PostedBlock.SignatureHash,
			MessagePoint:        list[key].BlockContent.MessagePoint,
			AggregatedPublicKey: list[key].BlockContent.AggregatedPublicKey,
			AggregatedSignature: list[key].BlockContent.AggregatedSignature,
			Senders:             list[key].BlockContent.Senders,
		}
	}

	return result, nil
}
