package block_validity_prover_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	blockTreeProofByRootAndLeafBlockNumbers "intmax2-node/internal/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	blockValidityProverBlockStatusByBlockHash "intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_hash"
	blockValidityProverBlockStatusByBlockNumber "intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_number"
	blockValidityProverInfo "intmax2-node/internal/use_cases/block_validity_prover_info"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	getVersion "intmax2-node/internal/use_cases/get_version"
	ucBlockTreeProofByRootAndLeafBlockNumbers "intmax2-node/pkg/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	ucBlockValidityProverBlockStatusByBlockHash "intmax2-node/pkg/use_cases/block_validity_prover_block_status_by_block_hash"
	ucBlockValidityProverBlockStatusByBlockNumber "intmax2-node/pkg/use_cases/block_validity_prover_block_status_by_block_number"
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
	BlockValidityProverBlockStatusByBlockHash(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) blockValidityProverBlockStatusByBlockHash.UseCaseBlockValidityProverBlockStatusByBlockHash
	BlockValidityProverBlockStatusByBlockNumber(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) blockValidityProverBlockStatusByBlockNumber.UseCaseBlockValidityProverBlockStatusByBlockNumber
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

func (c *commands) BlockValidityProverBlockStatusByBlockHash(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) blockValidityProverBlockStatusByBlockHash.UseCaseBlockValidityProverBlockStatusByBlockHash {
	return ucBlockValidityProverBlockStatusByBlockHash.New(cfg, log, db)
}

func (c *commands) BlockValidityProverBlockStatusByBlockNumber(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) blockValidityProverBlockStatusByBlockNumber.UseCaseBlockValidityProverBlockStatusByBlockNumber {
	return ucBlockValidityProverBlockStatusByBlockNumber.New(cfg, log, db)
}
