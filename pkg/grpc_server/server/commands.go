package server

import (
	"intmax2-node/configs"
	blockProposed "intmax2-node/internal/use_cases/block_proposed"
	getVersion "intmax2-node/internal/use_cases/get_version"
	healthCheck "intmax2-node/internal/use_cases/health_check"
	"intmax2-node/internal/use_cases/transaction"
	ucBlockProposed "intmax2-node/pkg/use_cases/block_proposed"
	ucGetVersion "intmax2-node/pkg/use_cases/get_version"
	ucHealthCheck "intmax2-node/pkg/use_cases/health_check"
	ucTransaction "intmax2-node/pkg/use_cases/transaction"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_commands_test.go -package=server_test -source=commands.go

type Commands interface {
	GetVersion(version, buildTime string) getVersion.UseCaseGetVersion
	HealthCheck(hc *health.Handler) healthCheck.UseCaseHealthCheck
	Transaction(
		cfg *configs.Config,
		dbApp SQLDriverApp,
		worker Worker,
	) transaction.UseCaseTransaction
	BlockProposed() blockProposed.UseCaseBlockProposed
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) GetVersion(version, buildTime string) getVersion.UseCaseGetVersion {
	return ucGetVersion.New(version, buildTime)
}

func (c *commands) HealthCheck(hc *health.Handler) healthCheck.UseCaseHealthCheck {
	return ucHealthCheck.New(hc)
}

func (c *commands) Transaction(
	cfg *configs.Config,
	dbApp SQLDriverApp,
	worker Worker,
) transaction.UseCaseTransaction {
	return ucTransaction.New(cfg, dbApp, worker)
}

func (c *commands) BlockProposed() blockProposed.UseCaseBlockProposed {
	return ucBlockProposed.New()
}
