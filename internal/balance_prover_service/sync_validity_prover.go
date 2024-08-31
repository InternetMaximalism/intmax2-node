package balance_prover_service

import (
	"context"
	"errors"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"

	"intmax2-node/configs"
)

type ValidityProcessor struct{}

func (s *ValidityProcessor) Prove(
	prevValidityProof *intMaxTypes.Plonky2Proof,
	validityWitness *block_validity_prover.ValidityWitness,
) (*intMaxTypes.Plonky2Proof, error) {
	return nil, errors.New("not implemented")
}

type SyncValidityProver struct {
	ValidityProcessor block_validity_prover.BlockValidityProver
	blockSynchronizer block_validity_prover.BlockSynchronizer
}

func NewSyncValidityProver(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	db block_validity_prover.SQLDriverApp,
) (*SyncValidityProver, error) {
	synchronizer, err := block_synchronizer.NewBlockSynchronizer(
		ctx, cfg, log,
	)
	if err != nil {
		return nil, err
	}
	validityProver, err := block_validity_prover.NewBlockValidityProver(ctx, cfg, log, sb, db)
	if err != nil {
		return nil, err
	}

	return &SyncValidityProver{
		ValidityProcessor: validityProver,
		blockSynchronizer: synchronizer,
	}, nil
}

// check synchronization of INTMAX blocks
func (s *SyncValidityProver) Sync() (err error) {
	// s.blockSynchronizer.SyncBlockTree(blockProverService)
	startBlock, err := s.ValidityProcessor.BlockBuilder().LastSeenBlockPostedEventBlockNumber()
	if err != nil {
		var ErrNotFound = errors.New("not found")
		if !errors.Is(err, ErrNotFound) {
			var ErrLastSeenBlockPostedEventBlockNumberFail = errors.New("last seen block posted event block number fail")
			panic(errors.Join(ErrLastSeenBlockPostedEventBlockNumberFail, err))
		}

		startBlock = s.blockSynchronizer.RollupContractDeployedBlockNumber()
	}

	const int5000Key = 5000
	endBlock := startBlock + int5000Key
	events, _, err := s.blockSynchronizer.FetchNewPostedBlocks(startBlock, &endBlock)
	if err != nil {
		var ErrFetchNewPostedBlocksFail = errors.New("fetch new posted blocks fail")
		return errors.Join(ErrFetchNewPostedBlocksFail, err)
	}

	if len(events) != 0 {
		return errors.New("not synchronized")
	}

	return nil
}

func (s *SyncValidityProver) FetchUpdateWitness(
	blockBuilder MockBlockBuilder,
	publicKey *intMaxAcc.PublicKey,
	blockNumber uint32,
	prevBlockNumber uint32,
	shouldProve bool,
) (*UpdateWitness, error) {
	// request validity prover
	return nil, errors.New("not implemented")
}
