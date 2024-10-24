package intmax_block

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	getINTMAXBlockInfo "intmax2-node/internal/use_cases/get_intmax_block_info"
	ucGetINTMAXBlockInfo "intmax2-node/pkg/use_cases/get_intmax_block_info"
)

type Commands interface {
	GetINTMAXBlockInfo(
		cfg *configs.Config,
		log logger.Logger,
	) getINTMAXBlockInfo.UseCaseGetINTMAXBlockInfo
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) GetINTMAXBlockInfo(
	cfg *configs.Config,
	log logger.Logger,
) getINTMAXBlockInfo.UseCaseGetINTMAXBlockInfo {
	return ucGetINTMAXBlockInfo.New(cfg, log)
}
