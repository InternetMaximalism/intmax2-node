package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txWithdrawalCmd(b *Transaction) *cobra.Command {
	const (
		use   = "withdrawal"
		short = "Manage withdrawal transaction"

		ethTokenType     = "eth"
		erc20TokenType   = "erc20"
		erc721TokenType  = "erc721"
		erc1155TokenType = "erc1155"
	)

	withdrawalCmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	withdrawalCmd.AddCommand(txWithdrawalTransfersListCmd(b))
	withdrawalCmd.AddCommand(txWithdrawalTransferInfoByHashCmd(b))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, ethTokenType))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, erc20TokenType))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, erc721TokenType))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, erc1155TokenType))

	var recipientAddressStr string
	withdrawalCmd.PersistentFlags().StringVar(&recipientAddressStr, recipientKey, emptyKey, recipientDescription)

	var userEthPrivateKey string
	withdrawalCmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	var resume bool
	withdrawalCmd.PersistentFlags().BoolVar(&resume, resumeKey, defaultResume, resumeDescription)

	withdrawalCmd.Run = func(cmd *cobra.Command, args []string) {
		err := b.SB.SetupEthereumNetworkChainID(b.Context)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}

		err = newCommands().SendWithdrawalTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, utils.RemoveZeroX(userEthPrivateKey), resume, b.DbApp)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
	}

	return &withdrawalCmd
}
