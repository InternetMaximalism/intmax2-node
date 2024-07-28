package backup_balance

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prodadidb/go-validation"

	intMaxAcc "intmax2-node/internal/accounts"
)

var (
	ErrValueInvalid = errors.New("must be a valid value")
)

func (input *UCPostBackupBalanceInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.User, validation.Required, input.isUser(func() *intMaxAcc.PublicKey {
			if input.DecodeUser == nil {
				input.DecodeUser = &intMaxAcc.PublicKey{}
			}
			return input.DecodeUser
		}())),
		validation.Field(&input.BlockNumber, validation.Required),
		validation.Field(&input.EncryptedBalanceProof.Proof, validation.Required),                 // TODO serhii: how to check?
		validation.Field(&input.EncryptedBalanceProof.EncryptedPublicInputs, validation.Required), // TODO serhii: how to check?
		validation.Field(&input.EncryptedBalanceData, validation.Required),
		validation.Field(&input.EncryptedTxs, validation.Each()),       // TODO serhii: how to check?
		validation.Field(&input.EncryptedTransfers, validation.Each()), // TODO serhii: how to check?
		validation.Field(&input.EncryptedDeposits, validation.Each()),  // TODO serhii: how to check?
		validation.Field(&input.Signature, validation.Required, validation.By(func(value interface{}) error {
			v, ok := value.(string)
			if !ok {
				return ErrValueInvalid
			}

			message := MakeMessage(
				input.DecodeUser.ToAddress(),
				input.BlockNumber,
				[]byte(input.EncryptedBalanceProof.Proof),
				[]byte(input.EncryptedBalanceProof.EncryptedPublicInputs),
				[]byte(input.EncryptedBalanceData),
				stringArrayToBytes(input.EncryptedTxs),
				stringArrayToBytes(input.EncryptedTransfers),
				stringArrayToBytes(input.EncryptedDeposits))

			var publicKey *intMaxAcc.PublicKey
			publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(input.User)
			if err != nil {
				return ErrValueInvalid
			}

			var sb []byte
			sb, err = hexutil.Decode(v)
			if err != nil {
				return ErrValueInvalid
			}

			sign := bn254.G2Affine{}
			err = sign.Unmarshal(sb)
			if err != nil {
				return ErrValueInvalid
			}

			err = intMaxAcc.VerifySignature(&sign, publicKey, message)
			if err != nil {
				return ErrValueInvalid
			}

			return nil
		})),
	)
}

func stringArrayToBytes(data []string) [][]byte {
	res := make([][]byte, len(data))
	for idx := range data {
		res[idx] = []byte(data[idx])
	}

	return res
}

func (input *UCPostBackupBalanceInput) isUser(pbKey *intMaxAcc.PublicKey) validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(v)
		if err != nil {
			return ErrValueInvalid
		}

		if pbKey != nil {
			*pbKey = *publicKey
		}

		return nil
	})
}
