package deposit_status_by_hash

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucDepositStatusByHash "intmax2-node/internal/use_cases/deposit_status_by_hash"

	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// type poseidonHashOut = goldenposeidon.PoseidonHashOut

// const base10 = 10

type uc struct {
	cfg                 *configs.Config
	log                 logger.Logger
	db                  SQLDriverApp
	blockValidityProver *block_validity_prover.BlockValidityProver
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	blockValidityProver *block_validity_prover.BlockValidityProver,
) ucDepositStatusByHash.UseCaseDepositStatusByHash {
	return &uc{
		cfg:                 cfg,
		log:                 log,
		db:                  db,
		blockValidityProver: blockValidityProver,
	}
}

func (u *uc) Do(
	ctx context.Context, input *ucDepositStatusByHash.UCDepositStatusByHashInput,
) (status *ucDepositStatusByHash.UCDepositStatusByHash, err error) {
	const (
		hName         = "UseCase BlockStatus"
		txTreeRootKey = "tx_tree_root"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(txTreeRootKey, input.DepositHash),
		))
	defer span.End()

	// DepositID         uint32
	// DepositIndex      *uint32
	// DepositHash       common.Hash
	// RecipientSaltHash [int32Key]byte
	// TokenIndex        uint32
	// Amount            *big.Int // uint256
	// CreatedAt         time.Time

	depositHash := common.HexToHash(input.DepositHash)
	_, err = u.db.DepositByDepositHash(depositHash)
	if err != nil {
		return nil, err
	}

	blockNumber := uint32(0)
	// if deposit.BlockNumber != nil {
	// 	blockNumber = deposit.BlockNumber
	// }

	// merkleProof := []poseidonHashOut{}

	status = &ucDepositStatusByHash.UCDepositStatusByHash{
		BlockNumber: blockNumber,
	}

	return status, nil
}
