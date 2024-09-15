package balance_prover_service

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
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

type syncValidityProver struct {
	blockValidityService block_validity_prover.BlockValidityService
	blockSynchronizer    block_validity_prover.BlockSynchronizer
}

// func NewSyncValidityProver(
// 	ctx context.Context,
// 	cfg *configs.Config,
// 	log logger.Logger,
// 	sb block_validity_prover.ServiceBlockchain,
// 	db block_validity_prover.SQLDriverApp,
// ) (*syncValidityProver, error) {
// 	synchronizer, err := block_synchronizer.NewBlockSynchronizer(
// 		ctx, cfg, log,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	validityProver, err := block_validity_prover.NewBlockValidityProver(ctx, cfg, log, sb, db)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &syncValidityProver{
// 		ValidityProver:    validityProver,
// 		blockSynchronizer: synchronizer,
// 	}, nil
// }

// type BlockValidityService interface {
// 	FetchLastDepositIndex() (uint32, error)
// 	LastSeenBlockPostedEventBlockNumber() (uint64, error)
// 	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
// 	LatestIntMaxBlockNumber() uint32
// 	BlockContentByTxRoot(txRoot string) (*block_post_service.PostedBlock, error)
// 	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error)
// 	BlockNumberByDepositIndex(depositIndex uint32) (uint32, error)
// 	LatestSynchronizedBlockNumber() (uint32, error)
// 	IsSynchronizedDepositIndex(depositIndex uint32) (bool, error)
// 	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error)
// 	DepositTreeProof(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
// 	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
// 	RollupContractDeployedBlockNumber() uint64
// 	PostBlock(isRegistrationBlock bool, txs []*MockTxRequest) (*ValidityWitness, error) // XXX
// }

type externalBlockValidityService struct {
	ctx context.Context
	cfg *configs.Config
}

func NewExternalBlockValidityProver(ctx context.Context, cfg *configs.Config) block_validity_prover.BlockValidityService {
	return &externalBlockValidityService{
		ctx: ctx,
		cfg: cfg,
	}
}

func (s *externalBlockValidityService) FetchLastDepositIndex() (uint32, error) {
	return 0, errors.New("not implemented")
}

// func (s *externalBlockValidityService) LastSeenBlockPostedEventBlockNumber() (uint64, error) {
// 	return 0, errors.New("not implemented")
// }

func (s *externalBlockValidityService) LatestIntMaxBlockNumber() (uint32, error) {
	return 0, nil
}

func (s *externalBlockValidityService) BlockContentByTxRoot(txRoot string) (*block_post_service.PostedBlock, error) {
	return nil, errors.New("not implemented")
}

func (s *externalBlockValidityService) GetDepositLeafAndIndexByHash(depositHash common.Hash) (*block_validity_prover.DepositInfo, error) {
	return nil, errors.New("not implemented")
}

func (s *externalBlockValidityService) LatestSynchronizedBlockNumber() (uint32, error) {
	return 0, errors.New("not implemented")
}

func (s *externalBlockValidityService) IsSynchronizedDepositIndex(depositIndex uint32) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *externalBlockValidityService) FetchUpdateWitness(
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber *uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*block_validity_prover.UpdateWitness, error) {
	return nil, errors.New("not implemented")
}

func (s *externalBlockValidityService) DepositTreeProof(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	return nil, common.Hash{}, errors.New("not implemented")
}

func (s *externalBlockValidityService) BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error) {
	return nil, errors.New("not implemented")
}

func (s *externalBlockValidityService) PostBlock(isRegistrationBlock bool, txs []*block_validity_prover.MockTxRequest) (*block_validity_prover.ValidityWitness, error) {
	return nil, errors.New("not implemented")
}

// func (s *externalBlockValidityService) RollupContractDeployedBlockNumber() (uint64, error) {
// 	return 0, errors.New("not implemented")
// }
