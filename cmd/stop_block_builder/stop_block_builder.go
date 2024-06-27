package stop_block_builder

import (
	"context"
	"fmt"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

func NewCmd(
	ctx context.Context,
	log logger.Logger,
	sb ServiceBlockchain,
) *cobra.Command {
	const (
		use   = "stop"
		short = "stop block builder command"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			err := sb.CheckScrollPrivateKey(ctx)
			if err != nil {
				const msg = "check private key error occurred: %v"
				log.Fatalf(msg, err.Error())
			}

			err = sb.StopBlockBuilder(ctx)
			if err != nil {
				const msg = "stop block builder error occurred: %v"
				log.Fatalf(msg, err.Error())
			}

			const msgOK = "The Block builder stopped successfully"
			fmt.Println(msgOK)
		},
	}
}
