package post_backup_user_state

import (
	"errors"
	bps "intmax2-node/internal/balance_prover_service"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCPostBackupUserStateInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.UserAddress, validation.Required),
		validation.Field(&input.BalanceProof, validation.Required, input.isBalanceProof()),
		validation.Field(&input.EncryptedUserState, validation.Required),
		validation.Field(&input.AuthSignature, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
	)
}

func (input *UCPostBackupUserStateInput) isBalanceProof() validation.Rule {
	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		var bp *intMaxTypes.Plonky2Proof
		bp, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(v)
		if err != nil {
			return ErrValueInvalid
		}

		_, err = new(bps.BalancePublicInputs).FromPublicInputs(bp.PublicInputs)
		if err != nil {
			return ErrValueInvalid
		}

		return nil
	})
}
