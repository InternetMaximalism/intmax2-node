package block_builder

import (
	"context"
	"errors"
	"fmt"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

func stopCmd(
	ctx context.Context,
	log logger.Logger,
	sb ServiceBlockchain,
	bbr BlockBuilderRegistryService,
) *cobra.Command {
	const (
		use   = "stop"
		short = "Stop block builder"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := log.WithFields(logger.Fields{"module": "stop-block-builder"})

		err := sb.CheckScrollPrivateKey(ctx)
		if err != nil && !errors.Is(err, errorsB.ErrWalletInsufficientFundsForNodeStart) {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = sb.SetupScrollNetworkChainID(ctx)
		if err != nil {
			const msg = "Fatal: %v"
			l.Fatalf(msg, err.Error())
		}

		err = bbr.StopBlockBuilder(ctx)
		if err != nil {
			const msg = "stop block builder error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		const msgOK = "The Block builder stopped successfully"
		fmt.Println(msgOK)
	}

	return &cmd
}
