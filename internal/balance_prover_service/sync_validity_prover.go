package balance_prover_service

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	bbsTypes "intmax2-node/internal/block_builder_storage/types"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

var ErrNotImplemented = errors.New("not implemented")

type externalBlockValidityService struct {
	ctx context.Context
	cfg *configs.Config
}

func NewExternalBlockValidityProver(
	ctx context.Context,
	cfg *configs.Config,
) block_validity_prover.BlockValidityService {
	return &externalBlockValidityService{
		ctx: ctx,
		cfg: cfg,
	}
}

func (s *externalBlockValidityService) BlockContentByTxRoot(
	db block_validity_prover.SQLDriverApp,
	txRoot common.Hash,
) (*block_post_service.PostedBlock, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) GetDepositInfoByHash(
	db block_validity_prover.SQLDriverApp,
	depositHash common.Hash,
) (*bbsTypes.DepositInfo, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) FetchValidityProverInfo(
	db block_validity_prover.SQLDriverApp,
) (*bbsTypes.ValidityProverInfo, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) FetchUpdateWitness(
	db block_validity_prover.SQLDriverApp,
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber *uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*bbsTypes.UpdateWitness, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) DepositTreeProof(
	db block_validity_prover.SQLDriverApp,
	depositIndex uint32,
) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	return nil, common.Hash{}, ErrNotImplemented
}

func (s *externalBlockValidityService) BlockTreeProof(
	rootBlockNumber, leafBlockNumber uint32,
) (*intMaxTree.PoseidonMerkleProof, error) {
	return nil, ErrNotImplemented
}

func (s *externalBlockValidityService) ValidityPublicInputs(
	db block_validity_prover.SQLDriverApp,
	txRoot common.Hash,
) (*bbsTypes.ValidityPublicInputs, []bbsTypes.SenderLeaf, error) {
	return nil, nil, ErrNotImplemented
}
