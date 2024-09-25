package types

import (
	"encoding/json"
	"errors"
	"intmax2-node/internal/finite_field"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type SenderLeaf struct {
	Sender  *big.Int
	IsValid bool
}

func (leaf *SenderLeaf) ToFieldElementSlice() []ffg.Element {
	buf := finite_field.NewBuffer(make([]ffg.Element, 0))
	sender := intMaxTypes.BigIntToBytes32BeArray(leaf.Sender)
	finite_field.WriteFixedSizeBytes(buf, sender[:], intMaxTypes.NumPublicKeyBytes)
	if leaf.IsValid {
		finite_field.WriteUint32(buf, 1)
	} else {
		finite_field.WriteUint32(buf, 0)
	}

	return buf.Inner()
}

func (leaf *SenderLeaf) Hash() *intMaxGP.PoseidonHashOut {
	return intMaxGP.HashNoPad(leaf.ToFieldElementSlice())
}

func (leaf *SenderLeaf) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableSenderLeaf{
		Sender:  leaf.Sender.String(),
		IsValid: leaf.IsValid,
	})
}

func (leaf *SenderLeaf) UnmarshalJSON(data []byte) error {
	var v SerializableSenderLeaf
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	sender, ok := new(big.Int).SetString(v.Sender, Base10)
	if !ok {
		return errors.New("invalid sender")
	}

	leaf.Sender = sender
	leaf.IsValid = v.IsValid

	return nil
}

func GetSenderLeaves(publicKeys []intMaxTypes.Uint256, senderFlag intMaxTypes.Bytes16) []SenderLeaf {
	senderLeaves := make([]SenderLeaf, 0)
	for i, publicKey := range publicKeys {
		senderLeaf := SenderLeaf{
			Sender:  publicKey.BigInt(),
			IsValid: GetBitFromUint32Slice(senderFlag[:], i),
		}
		senderLeaves = append(senderLeaves, senderLeaf)
	}

	return senderLeaves
}
