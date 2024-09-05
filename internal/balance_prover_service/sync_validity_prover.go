package balance_prover_service

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type ValidityProcessor interface {
	Prove(prevValidityProof *intMaxTypes.Plonky2Proof, validityWitness *block_validity_prover.ValidityWitness) (*intMaxTypes.Plonky2Proof, error)
}

// type validityProcessor struct{}

// func (p *validityProcessor) Prove(
// 	prevValidityProof *intMaxTypes.Plonky2Proof,
// 	validityWitness *block_validity_prover.ValidityWitness,
// ) (*intMaxTypes.Plonky2Proof, error) {
// 	return nil, errors.New("not implemented")
// }

type ExternalValidityProcessor struct {
}

func NewExternalValidityProcessor() *ExternalValidityProcessor {
	return nil
}

func (p *ExternalValidityProcessor) Prove(prevValidityProof *intMaxTypes.Plonky2Proof, validityWitness *block_validity_prover.ValidityWitness) (*intMaxTypes.Plonky2Proof, error) {
	return nil, nil
}

type syncExternalValidityProver struct {
	ValidityProcessor ValidityProcessor
	LastBlockNumber   uint32
	ValidityProofs    map[uint32]*intMaxTypes.Plonky2Proof
}

// func NewSyncValidityProver() block_validity_prover.SyncValidityProver {
// 	return &syncExternalValidityProver{
// 		ValidityProcessor: NewExternalValidityProcessor(),
// 		LastBlockNumber:   0,
// 		ValidityProofs:    make(map[uint32]*intMaxTypes.Plonky2Proof),
// 	}
// }

type syncValidityProver struct {
	log               logger.Logger
	ValidityProver    block_validity_prover.BlockValidityProver
	blockSynchronizer block_validity_prover.BlockSynchronizer
}

func NewSyncValidityProver(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	db block_validity_prover.SQLDriverApp,
) (*syncValidityProver, error) {
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

	return &syncValidityProver{
		log:               log,
		ValidityProver:    validityProver,
		blockSynchronizer: synchronizer,
	}, nil
}

// check synchronization of INTMAX blocks
func (s *syncValidityProver) Check() (err error) {
	// s.blockSynchronizer.SyncBlockTree(blockProverService)
	startBlock, err := s.ValidityProver.BlockBuilder().LastSeenBlockPostedEventBlockNumber()
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

func (s *syncValidityProver) Sync(validityWitness *block_validity_prover.ValidityWitness) (err error) {
	err = s.ValidityProver.SyncBlockProver()
	if err != nil {
		return err
	}

	return nil
}

func (s *syncValidityProver) FetchUpdateWitness(
	blockBuilder MockBlockBuilder,
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*UpdateWitness, error) {
	s.log.Debugf("FetchUpdateWitness currentBlockNumber: %d\n", currentBlockNumber)
	s.log.Debugf("FetchUpdateWitness targetBlockNumber: %d\n", targetBlockNumber)
	// request validity prover
	latestValidityProof, err := blockBuilder.ValidityProofByBlockNumber(currentBlockNumber)
	if err != nil {
		return nil, err
	}

	// blockMerkleProof := blockBuilder.GetBlockMerkleProof(currentBlockNumber, targetBlockNumber)
	blockMerkleProof, err := blockBuilder.BlockTreeProof(currentBlockNumber, targetBlockNumber)
	if err != nil {
		return nil, err
	}

	var accountMembershipProof *intMaxTree.IndexedMembershipProof
	if isPrevAccountTree {
		s.log.Debugf("is PrevAccountTree %d\n", currentBlockNumber-1)
		accountMembershipProof, err = blockBuilder.GetAccountMembershipProof(currentBlockNumber-1, publicKey.BigInt())
		if err != nil {
			return nil, err
		}
	} else {
		s.log.Debugf("is not PrevAccountTree %d\n", currentBlockNumber)
		accountMembershipProof, err = blockBuilder.GetAccountMembershipProof(currentBlockNumber, publicKey.BigInt())
		if err != nil {
			return nil, err
		}
	}

	return &UpdateWitness{
		ValidityProof:          *latestValidityProof,
		BlockMerkleProof:       *blockMerkleProof,
		AccountMembershipProof: accountMembershipProof,
	}, nil
}
