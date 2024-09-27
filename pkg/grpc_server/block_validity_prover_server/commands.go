package block_validity_prover_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	blockTreeProofByRootAndLeafBlockNumbers "intmax2-node/internal/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	blockValidityProverInfo "intmax2-node/internal/use_cases/block_validity_prover_info"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	getVersion "intmax2-node/internal/use_cases/get_version"
	ucBlockTreeProofByRootAndLeafBlockNumbers "intmax2-node/pkg/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	ucBlockValidityProverInfo "intmax2-node/pkg/use_cases/block_validity_prover_info"
	ucDepositTreeProofByDepositIndex "intmax2-node/pkg/use_cases/deposit_tree_proof_by_deposit_index"
	ucGetVersion "intmax2-node/pkg/use_cases/get_version"
)

//go:generate mockgen -destination=mock_commands_test.go -package=block_validity_prover_server_test -source=commands.go

type Commands interface {
	GetVersion(version, buildTime string) getVersion.UseCaseGetVersion
	DepositTreeProofByDepositIndex(
		cfg *configs.Config,
		log logger.Logger,
		bvs BlockValidityService,
	) depositTreeProofByDepositIndex.UseCaseDepositTreeProofByDepositIndex
	BlockTreeProofByRootAndLeafBlockNumbers(
		cfg *configs.Config,
		log logger.Logger,
		bvs BlockValidityService,
	) blockTreeProofByRootAndLeafBlockNumbers.UseCaseBlockTreeProofByRootAndLeafBlockNumbers
	BlockValidityProverInfo(
		cfg *configs.Config,
		log logger.Logger,
		bvs BlockValidityService,
	) blockValidityProverInfo.UseCaseBlockValidityProverInfo
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) GetVersion(version, buildTime string) getVersion.UseCaseGetVersion {
	return ucGetVersion.New(version, buildTime)
}

func (c *commands) DepositTreeProofByDepositIndex(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) depositTreeProofByDepositIndex.UseCaseDepositTreeProofByDepositIndex {
	return ucDepositTreeProofByDepositIndex.New(cfg, log, bvs)
}

func (c *commands) BlockTreeProofByRootAndLeafBlockNumbers(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) blockTreeProofByRootAndLeafBlockNumbers.UseCaseBlockTreeProofByRootAndLeafBlockNumbers {
	return ucBlockTreeProofByRootAndLeafBlockNumbers.New(cfg, log, bvs)
}

func (c *commands) BlockValidityProverInfo(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) blockValidityProverInfo.UseCaseBlockValidityProverInfo {
	return ucBlockValidityProverInfo.New(cfg, log, bvs)
}
