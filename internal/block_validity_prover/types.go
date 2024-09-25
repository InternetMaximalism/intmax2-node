package block_validity_prover

import (
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/crypto"
)

func GetPublicKeysHash(publicKeys []intMaxTypes.Uint256) intMaxTypes.Bytes32 {
	publicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range publicKeys {
		publicKeyBytes := sender.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[int32Key*i:int32Key*(i+1)], publicKeyBytes)
	}
	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < intMaxTypes.NumOfSenders; i++ {
		publicKeyBytes := dummyPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[int32Key*i:int32Key*(i+1)], publicKeyBytes[:])
	}

	publicKeysHash := crypto.Keccak256(publicKeysBytes) // TODO: Is this correct hash?

	var result intMaxTypes.Bytes32
	result.FromBytes(publicKeysHash)

	return result
}

type MockTxRequest struct {
	Sender              *intMaxAcc.PrivateKey
	AccountID           uint64
	Tx                  *intMaxTypes.Tx
	WillReturnSignature bool
}
