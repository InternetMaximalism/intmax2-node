package block_validity_prover_tx_root_status

import (
	"errors"
	"intmax2-node/configs"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prodadidb/go-validation"
)

// ErrTxRootNotExisting error: txRoot not existing.
var ErrTxRootNotExisting = errors.New("txRoot not existing")

func (input *UCBlockValidityProverTxRootStatusInput) Valid(cfg *configs.Config) error {
	const int1Key = 1

	return validation.ValidateStruct(input,
		validation.Field(&input.TxRoot,
			validation.Required,
			validation.Length(int1Key, cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest),
			validation.Each(input.IsTxRoot(cfg)),
		),
	)
}

func (input *UCBlockValidityProverTxRootStatusInput) IsTxRoot(cfg *configs.Config) validation.Rule {
	const zeroTxRootKey = "0x0000000000000000000000000000000000000000000000000000000000000000"
	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(string)
		if !ok {
			return ErrTxRootNotExisting
		}

		invalidTxsRoot := map[string]string{
			zeroTxRootKey: zeroTxRootKey,
		}
		for key := range cfg.BlockValidityProver.BlockValidityProverInvalidTxRootInRequest {
			invalidTxsRoot[cfg.BlockValidityProver.BlockValidityProverInvalidTxRootInRequest[key]] =
				cfg.BlockValidityProver.BlockValidityProverInvalidTxRootInRequest[key]
		}

		var t common.Hash
		err = t.Scan(common.FromHex(v))
		if err != nil || func() bool {
			_, isInvalidTxRoot := invalidTxsRoot[t.String()]
			return isInvalidTxRoot
		}() {
			if input.TxRootErrors == nil {
				input.TxRootErrors = make(map[string]*TxRootError)
			}

			input.TxRootErrors[v] = &TxRootError{
				Message: ErrTxRootNotExisting.Error(),
			}

			return nil
		}

		input.ConvertTxRoot = append(input.ConvertTxRoot, t)

		return nil
	})
}
