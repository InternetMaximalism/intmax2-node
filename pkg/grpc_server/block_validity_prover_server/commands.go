package block_validity_prover_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	getVersion "intmax2-node/internal/use_cases/get_version"
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
