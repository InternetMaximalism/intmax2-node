package block_builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/internal/block_builder_registry_service"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

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

		err = sb.SetupScrollNetworkChainID(ctx)
		if err != nil {
			const msg = "Fatal: %v"
			l.Fatalf(msg, err.Error())
		}

		var info *block_builder_registry_service.IBlockBuilderRegistryBlockBuilderInfo
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
