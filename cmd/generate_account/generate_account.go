package generate_account

import (
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/mnemonic_wallet/models"
	"strconv"

	"github.com/spf13/cobra"
)

func NewCmd(log logger.Logger) *cobra.Command {
	const (
		use   = "generate_account"
		short = "Generate new Ethereum and IntMax accounts"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	const (
		emptyKey                    = ""
		mnemonicPasswordKey         = "mnemonic_password"                                           // nolint:gosec
		mnemonicPasswordDescription = "mnemonic_password flag. use as --mnemonic_password \"pass\"" // nolint:gosec
		derivationDef               = "m/44'/60'/0'/0/"
		derivationKey               = "derivation_path"
		derivationDesc              = "derivation_path flag. use as --derivation_path \"m/44'/60'/0'/0/\""
		keyNumberKey                = "key_number"
		keyNumberDescription        = "key_number flag. use as --key_number \"0\" (0 - parent account, 1...n - child accounts numbers)"
		int10Key                    = 10
	)

	var mnemonicPassword string
	cmd.PersistentFlags().StringVar(&mnemonicPassword, mnemonicPasswordKey, emptyKey, mnemonicPasswordDescription)

	var derivationPath string
	cmd.PersistentFlags().StringVar(&derivationPath, derivationKey, derivationDef, derivationDesc)

	var keyNumber string
	cmd.PersistentFlags().StringVar(&keyNumber, keyNumberKey, emptyKey, keyNumberDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := log.WithFields(logger.Fields{"module": "generate_account"})

		var (
			err    error
			number int
		)
		if keyNumber != emptyKey {
			number, err = strconv.Atoi(keyNumber)
			if err != nil {
				const msg = "failed to convert key_number string to int: %+v"
				l.Fatalf(msg, err)
			}
		}

		var w *models.Wallet
		w, err = mnemonic_wallet.New().WalletGenerator(
			derivationPath+strconv.FormatInt(int64(number), int10Key),
			mnemonicPassword,
		)
		if err != nil {
			const msg = "failed to apply wallet generator: %+v"
			l.Fatalf(msg, err)
		}

		var wb []byte
		wb, err = w.Marshal()
		if err != nil {
			const msg = "failed to marshal wallet info: %+v"
			l.Fatalf(msg, err)
		}

		print(string(wb))
	}

	return &cmd
}
