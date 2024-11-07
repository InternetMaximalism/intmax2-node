package block_validity_prover_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	blockTreeProofByRootAndLeafBlockNumbers "intmax2-node/internal/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	blockValidityProverAccount "intmax2-node/internal/use_cases/block_validity_prover_account"
	blockValidityProverBalanceUpdateWitness "intmax2-node/internal/use_cases/block_validity_prover_balance_update_witness"
	blockValidityProverBlockStatusByBlockHash "intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_hash"
	blockValidityProverBlockStatusByBlockNumber "intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_number"
	blockValidityProverBlockValidityProof "intmax2-node/internal/use_cases/block_validity_prover_block_validity_proof"
	blockValidityProverBlockValidityPublicInputs "intmax2-node/internal/use_cases/block_validity_prover_block_validity_public_inputs"
	blockValidityProverInfo "intmax2-node/internal/use_cases/block_validity_prover_info"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	getVersion "intmax2-node/internal/use_cases/get_version"
	ucBlockTreeProofByRootAndLeafBlockNumbers "intmax2-node/pkg/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	ucBlockValidityProverAccount "intmax2-node/pkg/use_cases/block_validity_prover_account"
	ucBlockValidityProverBalanceUpdateWitness "intmax2-node/pkg/use_cases/block_validity_prover_balance_update_witness"
	ucBlockValidityProverBlockStatusByBlockHash "intmax2-node/pkg/use_cases/block_validity_prover_block_status_by_block_hash"
	ucBlockValidityProverBlockStatusByBlockNumber "intmax2-node/pkg/use_cases/block_validity_prover_block_status_by_block_number"
	ucBlockValidityProverBlockValidityProof "intmax2-node/pkg/use_cases/block_validity_prover_block_validity_proof"
	ucBlockValidityProverBlockValidityPublicInputs "intmax2-node/pkg/use_cases/block_validity_prover_block_validity_public_inputs"
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
	BlockValidityProverAccount(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) blockValidityProverAccount.UseCaseBlockValidityProverAccount
	BlockValidityProverBalanceUpdateWitness(
		cfg *configs.Config,
		log logger.Logger,
		bvs BlockValidityService,
	) blockValidityProverBalanceUpdateWitness.UseCaseBlockValidityProverBalanceUpdateWitness
	BlockValidityProverBlockValidityPublicInputs(
		cfg *configs.Config,
		log logger.Logger,
		bvs BlockValidityService,
	) blockValidityProverBlockValidityPublicInputs.UseCaseBlockValidityProverBlockValidityPublicInputs
	BlockValidityProverBlockValidityProof(
		cfg *configs.Config,
		log logger.Logger,
		bvs BlockValidityService,
	) blockValidityProverBlockValidityProof.UseCaseBlockValidityProverBlockValidityProof
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

func (c *commands) BlockValidityProverAccount(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) blockValidityProverAccount.UseCaseBlockValidityProverAccount {
	return ucBlockValidityProverAccount.New(cfg, log, db)
}

func (c *commands) BlockValidityProverBalanceUpdateWitness(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) blockValidityProverBalanceUpdateWitness.UseCaseBlockValidityProverBalanceUpdateWitness {
	return ucBlockValidityProverBalanceUpdateWitness.New(cfg, log, bvs)
}

func (c *commands) BlockValidityProverBlockValidityPublicInputs(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) blockValidityProverBlockValidityPublicInputs.UseCaseBlockValidityProverBlockValidityPublicInputs {
	return ucBlockValidityProverBlockValidityPublicInputs.New(cfg, log, bvs)
}

func (c *commands) BlockValidityProverBlockValidityProof(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) blockValidityProverBlockValidityProof.UseCaseBlockValidityProverBlockValidityProof {
	return ucBlockValidityProverBlockValidityProof.New(cfg, log, bvs)
}
