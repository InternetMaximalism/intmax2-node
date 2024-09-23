package block_validity_prover_server

import (
	getVersion "intmax2-node/internal/use_cases/get_version"
	ucGetVersion "intmax2-node/pkg/use_cases/get_version"
)

//go:generate mockgen -destination=mock_commands_test.go -package=block_validity_prover_server_test -source=commands.go

type Commands interface {
	GetVersion(version, buildTime string) getVersion.UseCaseGetVersion
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) GetVersion(version, buildTime string) getVersion.UseCaseGetVersion {
	return ucGetVersion.New(version, buildTime)
}
