package block_proposed

import (
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
)

func MakeMessage(inputTxHash, inputSender string, inputExpiration time.Time) ([]ffg.Element, error) {
	const (
		int1Key  = 1
		int4Key  = 4
		int8Key  = 8
		int32Key = 32
	)

	message := finite_field.NewBuffer(make([]ffg.Element, int4Key+int8Key+int1Key))
	txHash := new(goldenposeidon.PoseidonHashOut)
	txHashBytes, err := hexutil.Decode(inputTxHash)
	if err != nil {
		return nil, err
	}
	err = txHash.Unmarshal(txHashBytes)
	if err != nil {
		return nil, err
	}
	expiration := new(big.Int).SetInt64(inputExpiration.Unix())

	senderAddress, err := accounts.NewAddressFromHex(inputSender)
	if err != nil {
		return nil, err
	}

	finite_field.WritePoseidonHashOut(message, txHash)

	finite_field.WriteFixedSizeBytes(message, senderAddress.Bytes(), int32Key)

	if expiration.Cmp(ffg.Modulus()) >= 0 {
		return nil, ErrValueInvalid
	}
	err = finite_field.WriteUint64(message, expiration.Uint64())
	if err != nil {
		return nil, err
	}

	return message.Inner(), nil
}
