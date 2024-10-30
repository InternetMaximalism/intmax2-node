package intmax_block

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

func NewCmd(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) *cobra.Command {
	const (
		use   = "block"
		short = "Manage INTMAX block"
	)

	blockCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	blockCmd.AddCommand(infoCmd(ctx, cfg, log))

	return blockCmd
}
