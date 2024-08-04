package messenger

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	messengerRelayer "intmax2-node/internal/use_cases/messenger_relayer"
	ucMessengerRelayer "intmax2-node/pkg/use_cases/messenger_relayer"

	messengerWithdrawalRelayer "intmax2-node/internal/use_cases/messenger_withdrawal_relayer"
	messengerWithdrawalRelayerMock "intmax2-node/internal/use_cases/messenger_withdrawal_relayer_mock"
	ucMessengerWithdrawalRelayer "intmax2-node/pkg/use_cases/messenger_withdrawal_relayer"
	ucMessengerWithdrawalRelayerMock "intmax2-node/pkg/use_cases/messenger_withdrawal_relayer_mock"
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
	MessengerWithdrawalRelayer(
		ctx context.Context,
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) messengerWithdrawalRelayer.UseCaseMessengerWithdrawalRelayer
	MessengerWithdrawalRelayerMock(
		ctx context.Context,
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
		sb ServiceBlockchain,
	) messengerWithdrawalRelayerMock.UseCaseMessengerWithdrawalRelayerMock
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

func (c *commands) MessengerWithdrawalRelayer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) messengerWithdrawalRelayer.UseCaseMessengerWithdrawalRelayer {
	return ucMessengerWithdrawalRelayer.New(ctx, cfg, log, sb)
}

func (c *commands) MessengerWithdrawalRelayerMock(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) messengerWithdrawalRelayerMock.UseCaseMessengerWithdrawalRelayerMock {
	return ucMessengerWithdrawalRelayerMock.New(ctx, cfg, log, db, sb)
}
