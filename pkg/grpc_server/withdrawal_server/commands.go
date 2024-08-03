package withdrawal_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	postWithdrawalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	postWithdrawalsByHashes "intmax2-node/internal/use_cases/post_withdrawals_by_hashes"
	"intmax2-node/internal/withdrawal_service"
	ucPostWithdrawalRequest "intmax2-node/pkg/use_cases/post_withdrawal_request"
	ucPostWithdrawalsByHashes "intmax2-node/pkg/use_cases/post_withdrawals_by_hashes"
)

//go:generate mockgen -destination=mock_commands_test.go -package=withdrawal_server_test -source=commands.go

type Commands interface {
	PostWithdrawalRequest(cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb withdrawal_service.ServiceBlockchain) postWithdrawalRequest.UseCasePostWithdrawalRequest
	PostWithdrawalsByHashes(cfg *configs.Config, log logger.Logger, db SQLDriverApp) postWithdrawalsByHashes.UseCasePostWithdrawalsByHashes
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) PostWithdrawalRequest(cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb withdrawal_service.ServiceBlockchain) postWithdrawalRequest.UseCasePostWithdrawalRequest {
	return ucPostWithdrawalRequest.New(cfg, log, db, sb)
}

func (c *commands) PostWithdrawalsByHashes(cfg *configs.Config, log logger.Logger, db SQLDriverApp) postWithdrawalsByHashes.UseCasePostWithdrawalsByHashes {
	return ucPostWithdrawalsByHashes.New(cfg, log, db)
}
