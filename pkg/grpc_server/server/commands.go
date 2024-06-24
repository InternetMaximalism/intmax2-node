package server

import (
	"github.com/dimiro1/health"
	getVersion "intmax2-node/internal/use_cases/get_version"
	healthCheck "intmax2-node/internal/use_cases/health_check"
	ucGetVersion "intmax2-node/pkg/use_cases/get_version"
	ucHealthCheck "intmax2-node/pkg/use_cases/health_check"
)

//go:generate mockgen -destination=mock_commands_test.go -package=server_test -source=commands.go

type Commands interface {
	GetVersion(version, buildTime string) getVersion.UseCaseGetVersion
	HealthCheck(hc *health.Handler) healthCheck.UseCaseHealthCheck
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
