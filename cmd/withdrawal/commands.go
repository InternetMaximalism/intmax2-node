// nolint:unused
package withdrawal

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	withdrawalRelayer "intmax2-node/internal/use_cases/withdrawal_relayer"
	ucWithdrawalRelayer "intmax2-node/pkg/use_cases/withdrawal_relayer"
)

//go:generate mockgen -destination=mock_command.go -package=withdrawal -source=commands.go

type Commands interface {
	WithdrawalRelayer(
		ctx context.Context,
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) withdrawalRelayer.UseCaseWithdrawalRelayer
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) WithdrawalRelayer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) withdrawalRelayer.UseCaseWithdrawalRelayer {
	return ucWithdrawalRelayer.New(ctx, cfg, log, sb)
}
