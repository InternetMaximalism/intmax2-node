package withdrawal_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	postWithdrawalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	ucPostWithdrawalRequest "intmax2-node/pkg/use_cases/post_withdrawal_request"
)

//go:generate mockgen -destination=mock_commands_test.go -package=withdrawal_server_test -source=commands.go

type Commands interface {
	PostWithdrawalRequest(cfg *configs.Config, log logger.Logger, db SQLDriverApp) postWithdrawalRequest.UseCasePostWithdrawalRequest
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) PostWithdrawalRequest(cfg *configs.Config, log logger.Logger, db SQLDriverApp) postWithdrawalRequest.UseCasePostWithdrawalRequest {
	return ucPostWithdrawalRequest.New(cfg, log, db)
}
