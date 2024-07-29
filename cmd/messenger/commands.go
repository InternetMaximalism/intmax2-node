package messenger

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	messengerRelayer "intmax2-node/internal/use_cases/messenger_relayer"
	ucMessengerRelayer "intmax2-node/pkg/use_cases/messenger_relayer"
)

//go:generate mockgen -destination=mock_command.go -package=messenger -source=commands.go

type Commands interface {
	MessengerRelayer(
		Context context.Context,
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
		sb ServiceBlockchain,
	) messengerRelayer.UseCaseMessengerRelayer
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) MessengerRelayer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) messengerRelayer.UseCaseMessengerRelayer {
	return ucMessengerRelayer.New(ctx, cfg, log, db, sb)
}
