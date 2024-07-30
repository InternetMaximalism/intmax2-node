package block_signature

import (
	"encoding/json"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/iden3/go-iden3-crypto/keccak256"
)

func MakeMessage(inputTxHash, inputSender string, enoughBalanceProof *EnoughBalanceProofInput) ([]ffg.Element, error) {
	const (
		int4Key  = 4
		int8Key  = 8
		int32Key = 32
	)

	message := finite_field.NewBuffer(make([]ffg.Element, int4Key+int8Key+int32Key))
	txHash := new(goldenposeidon.PoseidonHashOut)
	txHashBytes, err := hexutil.Decode(inputTxHash)
	if err != nil {
		return nil, err
	}
	err = txHash.Unmarshal(txHashBytes)
	if err != nil {
		return nil, err
	}

	var bytesEBP []byte
	bytesEBP, err = json.Marshal(&enoughBalanceProof)
	if err != nil {
		return nil, err
	}

	senderAddress, err := accounts.NewAddressFromHex(inputSender)
	if err != nil {
		return nil, err
	}

	finite_field.WritePoseidonHashOut(message, txHash)

	finite_field.WriteFixedSizeBytes(message, senderAddress.Bytes(), int32Key)

	finite_field.WriteFixedSizeBytes(message, keccak256.Hash(bytesEBP), int32Key)

	return message.Inner(), nil
}
