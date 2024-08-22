package block_builder

import (
	"context"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

func NewCmd(
	ctx context.Context,
	log logger.Logger,
	sb ServiceBlockchain,
	bbr BlockBuilderRegistryService,
) *cobra.Command {
	const (
		use   = "block_builder"
		short = "Manage block builder"
	)

	blockBuilderCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	blockBuilderCmd.AddCommand(infoCmd(ctx, log, sb, bbr))
	blockBuilderCmd.AddCommand(stopCmd(ctx, log, sb, bbr))
	blockBuilderCmd.AddCommand(unStakeCmd(ctx, log, sb, bbr))

	return blockBuilderCmd
}
