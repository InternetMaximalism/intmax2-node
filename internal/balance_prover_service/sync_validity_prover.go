package balance_prover_service

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/intmax_block_content"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

var ErrNotImplemented = errors.New("not implemented")

type externalBlockValidityService struct {
	ctx context.Context
	cfg *configs.Config
}

func NewExternalBlockValidityProver(ctx context.Context, cfg *configs.Config) BlockValidityService {
	return &externalBlockValidityService{
		ctx: ctx,
		cfg: cfg,
	}
}

func (s *externalBlockValidityService) BlockContentByTxRoot(txRoot common.Hash) (*intmax_block_content.PostedBlock, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) AuxInfoListFromBlockContentByTxRoot(txRoot ...common.Hash) (map[common.Hash]*block_validity_prover.AuxInfo, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) GetDepositInfoByHash(depositHash common.Hash) (*block_validity_prover.DepositInfo, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) GetDepositsInfoByHash(depositHash ...common.Hash) (map[uint32]*block_validity_prover.DepositInfo, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) FetchValidityProverInfo() (*block_validity_prover.ValidityProverInfo, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) FetchUpdateWitness(
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*block_validity_prover.UpdateWitness, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	return nil, common.Hash{}, ErrNotImplemented
}

func (s *externalBlockValidityService) BlockTreeProof(
	rootBlockNumber, leafBlockNumber uint32,
) (
	*intMaxTree.PoseidonMerkleProof,
	*intMaxTree.PoseidonHashOut,
	error,
) {
	return nil, nil, ErrNotImplemented
}

func (s *externalBlockValidityService) ValidityPublicInputsByBlockNumber(
	blockNumber uint32,
) (*block_validity_prover.ValidityPublicInputs, []block_validity_prover.SenderLeaf, error) {
	return nil, nil, ErrNotImplemented
}

func (s *externalBlockValidityService) ValidityPublicInputs(txRoot common.Hash) (*block_validity_prover.ValidityPublicInputs, []block_validity_prover.SenderLeaf, error) {
	return nil, nil, ErrNotImplemented
}

func (s *externalBlockValidityService) ValidityProofByBlockNumber(blockNumber uint32) (*block_validity_prover.ValidityProof, error) {
	return nil, ErrNotImplemented
}
