package block_builder

import (
	"context"
	"errors"
	"fmt"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

func unStakeCmd(
	ctx context.Context,
	log logger.Logger,
	sb ServiceBlockchain,
	bbr BlockBuilderRegistryService,
) *cobra.Command {
	const (
		use   = "unstake"
		short = "Unstake block builder"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := log.WithFields(logger.Fields{"module": "unstake-block-builder"})

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

		err = bbr.UnStakeBlockBuilder(ctx)
		if err != nil {
			const msg = "unstake block builder error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		const msgOK = "The Block builder unstaked successfully"
		fmt.Println(msgOK)
	}

	return &cmd
}
