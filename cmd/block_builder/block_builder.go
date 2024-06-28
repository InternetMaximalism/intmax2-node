package block_builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/internal/bindings"
	errorsB "intmax2-node/internal/blockchain/errors"
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

func infoCmd(
	ctx context.Context,
	log logger.Logger,
	sb ServiceBlockchain,
	bbr BlockBuilderRegistryService,
) *cobra.Command {
	const (
		use   = "info"
		short = "Returns the block builder info"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := log.WithFields(logger.Fields{"module": "info-block-builder"})

		err := sb.CheckScrollPrivateKey(ctx)
		if err != nil && !errors.Is(err, errorsB.ErrWalletInsufficientFundsForNodeStart) {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		var info *bindings.IBlockBuilderRegistryBlockBuilderInfo
		info, err = bbr.GetBlockBuilder(ctx)
		if err != nil {
			const msg = "get the block builder info error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		var bytesI []byte
		bytesI, err = json.Marshal(&info)
		if err != nil {
			const msg = "marshal the block builder info error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		const msgOK = "The Block builder info: %s"
		fmt.Println(fmt.Sprintf(msgOK, string(bytesI)))
	}

	return &cmd
}

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
