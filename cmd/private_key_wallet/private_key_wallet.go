package private_key_wallet

import (
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/mnemonic_wallet/models"

	"github.com/spf13/cobra"
)

func NewCmd(log logger.Logger) *cobra.Command {
	const (
		use   = "private_key_wallet"
		short = "Generate Ethereum wallet from private key"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	const (
		emptyKey                   = ""
		privateKeyInHexKey         = "private_key"
		privateKeyInHexDescription = "private_key flag. use as --private_key \"__PRIVATE_KEY_IN_HEX_WITHOUT_0x__\""
	)

	var privateKeyInHex string
	cmd.PersistentFlags().StringVar(&privateKeyInHex, privateKeyInHexKey, emptyKey, privateKeyInHexDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := log.WithFields(logger.Fields{"module": use})

		if privateKeyInHex == emptyKey {
			const msg = "private_key flag is not set"
			l.Fatalf("%s", msg)
		}

		var (
			err error
			w   *models.Wallet
		)
		w, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(
			privateKeyInHex,
		)
		if err != nil {
			const msg = "failed to get wallet from private key: %+v"
			l.Fatalf(msg, err)
		}

		var wb []byte
		wb, err = w.Marshal()
		if err != nil {
			const msg = "failed to marshal wallet: %+v"
			l.Fatalf(msg, err)
		}

		print(string(wb))
	}

	return &cmd
}
