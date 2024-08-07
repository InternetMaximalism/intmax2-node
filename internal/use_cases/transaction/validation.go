package transaction

import (
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prodadidb/go-validation"
)

// const emptyETHAddr = "0x0000000000000000000000000000000000000000"

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

// ErrFailToUnmarshalTransfersHash error: failed to unmarshal transfers hash.
var ErrFailToUnmarshalTransfersHash = errors.New("failed to unmarshal transfers hash")

// ErrFailToDecodeTransfersHash error: failed to decode transfers hash.
var ErrFailToDecodeTransfersHash = errors.New("failed to decode transfers hash")

// ErrFailToVerifyPoWNonce error: failed to verify PoW nonce.
var ErrFailToVerifyPoWNonce = errors.New("failed to verify PoW nonce")

// ErrMoreThenZero error: must be more then 0.
var ErrMoreThenZero = errors.New("must be more then 0")

func (input *UCTransactionInput) Valid(cfg *configs.Config, pow PoWNonce) error {
	// var (
	// 	iTxData int
	// )

	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required, input.isSender(func() *intMaxAcc.PublicKey {
			if input.DecodeSender == nil {
				input.DecodeSender = &intMaxAcc.PublicKey{}
			}
			return input.DecodeSender
		}())),
		validation.Field(&input.TransfersHash, validation.Required, input.isHexDecode()),
		validation.Field(&input.PowNonce, validation.Required, input.isPoW(pow)),
		// NOTE: `TransferData` does not need to be sent in the request
		// validation.Field(&input.TransferData, validation.Required, input.transferDataLength(cfg), validation.Each(
		// 	validation.Required, input.isTransferData(func() *TransferDataTransaction {
		// 		iTxData++
		// 		if input.TransferData == nil {
		// 			return nil
		// 		}
		// 		return input.TransferData[iTxData-1]
		// 	}()),
		// 	input.calculateHashData(func() *TransferDataTransaction {
		// 		if input.TransferData == nil {
		// 			return nil
		// 		}
		// 		return input.TransferData[iTxData-1]
		// 	}()),
		// ), input.checkHashWithData(&input.TransfersHash)),
		validation.Field(&input.Nonce, validation.Required, input.nonceMaxLength(cfg)),
		validation.Field(&input.Expiration, validation.Required, validation.By(func(value interface{}) error {
			v, ok := value.(time.Time)
			if !ok {
				return ErrValueInvalid
			}

			if time.Now().UTC().UnixNano() > v.UnixNano() {
				return ErrValueInvalid
			}

			return nil
		})),
		validation.Field(&input.Signature, validation.Required, validation.By(func(value interface{}) error {
			v, ok := value.(string)
			if !ok {
				return ErrValueInvalid
			}

			transfersHash, err := hexutil.Decode(input.TransfersHash)
			if err != nil {
				return ErrValueInvalid
			}
			sender, err := intMaxAcc.NewAddressFromHex(input.Sender)
			if err != nil {
				return ErrValueInvalid
			}
			message, err := MakeMessage(transfersHash, input.Nonce, input.PowNonce, sender, input.Expiration)
			if err != nil {
				return ErrValueInvalid
			}

			var publicKey *intMaxAcc.PublicKey
			publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(input.Sender)
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

func (input *UCTransactionInput) isHexDecode() validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		_, err := hexutil.Decode(v)
		if err != nil {
			return ErrValueInvalid
		}

		return nil
	})
}

// func (input *UCTransactionInput) isConvertStrToBigInt(cv **big.Int) validation.Rule {
// 	const int10Key = 10
// 	return validation.By(func(value interface{}) error {
// 		v, ok := value.(string)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		*cv, ok = new(big.Int).SetString(v, int10Key)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		return nil
// 	})
// }

func (input *UCTransactionInput) isPoW(pow PoWNonce) validation.Rule {
	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		transfersHashBytes, err := hexutil.Decode(input.TransfersHash)
		if err != nil {
			return ErrFailToDecodeTransfersHash
		}

		transfersHash := new(intMaxTypes.PoseidonHashOut)
		err = transfersHash.Unmarshal(transfersHashBytes)
		if err != nil {
			return ErrFailToUnmarshalTransfersHash
		}

		tx, err := intMaxTypes.NewTx(
			transfersHash,
			input.Nonce,
		)
		if err != nil {
			return fmt.Errorf("failed to create new tx: %w", err)
		}

		txHash := tx.Hash()

		messageForPow := txHash.Marshal()
		err = pow.Verify(v, messageForPow)
		if err != nil {
			return ErrFailToVerifyPoWNonce
		}

		return nil
	})
}

// func (input *UCTransactionInput) isEthereumAddress(ga *intMaxTypes.GenericAddress) validation.Rule {
// 	return validation.By(func(value interface{}) (err error) {
// 		v, ok := value.(string)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		addr := common.HexToAddress(v)
// 		if addr.Hex() == emptyETHAddr {
// 			return ErrValueInvalid
// 		}

// 		var gaAddr *intMaxTypes.GenericAddress
// 		gaAddr, err = intMaxTypes.NewEthereumAddress(addr.Bytes())
// 		if err != nil {
// 			return ErrValueInvalid
// 		}

// 		if ga != nil {
// 			*ga = *gaAddr
// 		}

// 		return nil
// 	})
// }

// func (input *UCTransactionInput) isIntMaxAddress(ga *intMaxTypes.GenericAddress) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		v, ok := value.(string)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(v)
// 		if err != nil {
// 			return ErrValueInvalid
// 		}

// 		var gaAddr *intMaxTypes.GenericAddress
// 		gaAddr, err = intMaxTypes.NewINTMAXAddress(publicKey.ToAddress().Bytes())
// 		if err != nil {
// 			return ErrValueInvalid
// 		}

// 		if ga != nil {
// 			*ga = *gaAddr
// 		}

// 		return nil
// 	})
// }

func (input *UCTransactionInput) isSender(pbKey *intMaxAcc.PublicKey) validation.Rule {
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

// func (input *UCTransactionInput) isRecipient(tdTr *TransferDataTransaction) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		if tdTr == nil {
// 			return ErrValueInvalid
// 		}

// 		var isNil bool
// 		value, isNil = validation.Indirect(value)
// 		if isNil || validation.IsEmpty(value) {
// 			return ErrValueInvalid
// 		}

// 		recipient, ok := value.(RecipientTransferDataTransaction)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		return validation.ValidateStruct(&recipient,
// 			validation.Field(&recipient.AddressType, validation.Required, validation.In(
// 				intMaxAccTypes.INTMAXAddressType, intMaxAccTypes.EthereumAddressType,
// 			)),
// 			validation.Field(&recipient.Address,
// 				validation.Required, input.isHexDecode(), validation.By(func(value interface{}) (err error) {
// 					switch recipient.AddressType {
// 					case intMaxAccTypes.INTMAXAddressType:
// 						tdTr.DecodeRecipient = &intMaxTypes.GenericAddress{}
// 						err = input.isIntMaxAddress(tdTr.DecodeRecipient).Validate(value)
// 						if err != nil {
// 							return ErrValueInvalid
// 						}
// 					case intMaxAccTypes.EthereumAddressType:
// 						tdTr.DecodeRecipient = &intMaxTypes.GenericAddress{}
// 						err = input.isEthereumAddress(tdTr.DecodeRecipient).Validate(value)
// 						if err != nil {
// 							return ErrValueInvalid
// 						}
// 					}

// 					return nil
// 				}),
// 			))
// 	})
// }

func (input *UCTransactionInput) nonceMaxLength(*configs.Config) validation.Rule {
	return validation.By(func(value interface{}) error {
		_, ok := value.(uint64)
		if !ok {
			return ErrValueInvalid
		}

		return nil
	})
}

// func (input *UCTransactionInput) transferDataLength(cfg *configs.Config) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		var isNil bool
// 		value, isNil = validation.Indirect(value)
// 		if isNil || validation.IsEmpty(value) {
// 			return ErrValueInvalid
// 		}

// 		data, ok := value.([]*TransferDataTransaction)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		if len(data) > cfg.Blockchain.MaxCounterOfTransaction {
// 			return ErrValueInvalid
// 		}

// 		return nil
// 	})
// }

// func (input *UCTransactionInput) isSalt(tdTr *TransferDataTransaction) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		if tdTr == nil {
// 			return ErrValueInvalid
// 		}

// 		v, ok := value.(string)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		d, err := hexutil.Decode(v)
// 		if err != nil {
// 			return ErrValueInvalid
// 		}

// 		var ph intMaxTypes.PoseidonHashOut
// 		err = ph.Unmarshal(d)
// 		if err != nil {
// 			return ErrValueInvalid
// 		}

// 		tdTr.DecodeSalt = &ph

// 		return nil
// 	})
// }

// func (input *UCTransactionInput) isTransferData(tdTr *TransferDataTransaction) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		if tdTr == nil {
// 			return ErrValueInvalid
// 		}

// 		var isNil bool
// 		value, isNil = validation.Indirect(value)
// 		if isNil || validation.IsEmpty(value) {
// 			return ErrValueInvalid
// 		}

// 		data, ok := value.(TransferDataTransaction)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		return validation.ValidateStruct(&data,
// 			validation.Field(&data.Recipient, validation.Required, input.isRecipient(tdTr)),
// 			validation.Field(&data.TokenIndex, validation.Required,
// 				input.isConvertStrToBigInt(&tdTr.DecodeTokenIndex)),
// 			validation.Field(&data.Amount, validation.Required,
// 				input.isConvertStrToBigInt(&tdTr.DecodeAmount),
// 				validation.By(func(_ interface{}) error {
// 					if tdTr.DecodeAmount.Cmp(new(big.Int).SetInt64(0)) == 0 {
// 						return ErrMoreThenZero
// 					}

// 					return nil
// 				})),
// 			validation.Field(&data.Salt, validation.Required, input.isSalt(tdTr)))
// 	})
// }

// func (input *UCTransactionInput) calculateHashData(tdTr *TransferDataTransaction) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		if tdTr == nil {
// 			return ErrValueInvalid
// 		}

// 		tr := intMaxTypes.Transfer{
// 			Recipient:  tdTr.DecodeRecipient,
// 			TokenIndex: uint32(tdTr.DecodeTokenIndex.Uint64()),
// 			Amount:     tdTr.DecodeAmount,
// 			Salt:       tdTr.DecodeSalt,
// 		}

// 		tdTr.DecodeHash = tr.Hash()

// 		return nil
// 	})
// }

// func (input *UCTransactionInput) checkHashWithData(hash *string) validation.Rule {
// 	return validation.By(func(value interface{}) error {
// 		if hash == nil {
// 			return ErrValueInvalid
// 		}

// 		var isNil bool
// 		value, isNil = validation.Indirect(value)
// 		if isNil || validation.IsEmpty(value) {
// 			return ErrValueInvalid
// 		}

// 		data, ok := value.([]*TransferDataTransaction)
// 		if !ok {
// 			return ErrValueInvalid
// 		}

// 		hashTrList := make([][]byte, len(data))
// 		for key := range data {
// 			hashTrList[key] = data[key].DecodeHash.Marshal()
// 		}

// 		trHash := hexutil.Encode(keccak256.Hash(hashTrList...))

// 		if !strings.EqualFold(*hash, trHash) {
// 			return ErrValueInvalid
// 		}

// 		return nil
// 	})
// }
