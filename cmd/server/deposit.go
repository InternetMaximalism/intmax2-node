package server

import (
	"context"
	"intmax2-node/configs"
	service "intmax2-node/internal/deposit_service"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

var depoitCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Manage deposit",
}

// TODO: Regular execution, Scalability, Tx Management, Error Handling, Timeout
func NewDepositCmd(ctx context.Context, cfg *configs.Config, log logger.Logger) *cobra.Command {
	var depositAnalyzerCmd = &cobra.Command{
		Use:   "analyzer",
		Short: "Run deposit analyzer service",
		Run: func(cmd *cobra.Command, args []string) {
			service.DepositAnalyzer(ctx, cfg, log)
		},
	}

	var depositRelayerCmd = &cobra.Command{
		Use:   "relayer",
		Short: "Run deposit relayer service",
		Run: func(cmd *cobra.Command, args []string) {
			service.DepositRelayer(ctx, cfg, log)
		},
	}

	depoitCmd.AddCommand(depositAnalyzerCmd)
	depoitCmd.AddCommand(depositRelayerCmd)

	return depoitCmd
}
