package types

import (
	"encoding/binary"
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const numHashBytes = 32

type DepositLeaf struct {
	RecipientSaltHash [numHashBytes]byte
	TokenIndex        uint32
	Amount            *big.Int
}

func (dd *DepositLeaf) Set(deposit *DepositLeaf) *DepositLeaf {
	dd.RecipientSaltHash = deposit.RecipientSaltHash
	dd.TokenIndex = deposit.TokenIndex
	dd.Amount = deposit.Amount
	return dd
}

func (dd *DepositLeaf) SetZero() *DepositLeaf {
	dd.RecipientSaltHash = [numHashBytes]byte{}
	dd.TokenIndex = 0
	dd.Amount = big.NewInt(0)
	return dd
}

func (dd *DepositLeaf) Marshal() []byte {
	const (
		int4Key  = 4
		int31Key = 31
		int32Key = 32
	)

	tokenIndexBytes := make([]byte, int4Key)
	binary.BigEndian.PutUint32(tokenIndexBytes, dd.TokenIndex)
	amountBytes := make([]byte, int32Key)
	for i, v := range dd.Amount.Bytes() {
		amountBytes[int31Key-i] = v
	}

	return append(
		append(dd.RecipientSaltHash[:], tokenIndexBytes...),
		amountBytes...,
	)
}

func (dd *DepositLeaf) Hash() common.Hash {
	return crypto.Keccak256Hash(dd.Marshal())
}

func (dd *DepositLeaf) Equal(other *DepositLeaf) bool {
	switch {
	case dd.RecipientSaltHash != other.RecipientSaltHash,
		dd.TokenIndex != other.TokenIndex,
		dd.Amount.Cmp(other.Amount) != 0:
		return false
	default:
		return true
	}
}

type Deposit struct {
	Recipient  *intMaxAcc.PublicKey
	TokenIndex uint32
	Amount     *big.Int
	Salt       *goldenposeidon.PoseidonHashOut
}

func (d *Deposit) Set(deposit *Deposit) *Deposit {
	d.Recipient = new(intMaxAcc.PublicKey).Set(deposit.Recipient)
	d.TokenIndex = deposit.TokenIndex
	d.Amount = new(big.Int).Set(deposit.Amount)
	d.Salt = new(goldenposeidon.PoseidonHashOut).Set(deposit.Salt)
	return d
}

func (d *Deposit) Equal(other *Deposit) bool {
	return d.Recipient.Equal(other.Recipient) &&
		d.TokenIndex == other.TokenIndex &&
		d.Amount.Cmp(other.Amount) == 0 &&
		d.Salt.Equal(other.Salt)
}

func SplitBigIntTo32BitChunks(value *big.Int) []uint32 {
	const chunkSize = 32
	mask := new(big.Int).Lsh(big.NewInt(1), chunkSize)
	mask.Sub(mask, big.NewInt(1))
	chunks := make([]uint32, 0)
	for value.Cmp(big.NewInt(0)) > 0 {
		chunk := new(big.Int).And(value, mask)
		chunks = append([]uint32{uint32(chunk.Uint64())}, chunks...)
		value.Rsh(value, chunkSize)
	}
	return chunks
}

func GetPublicKeySaltHash(publicKey intMaxAcc.PublicKey, salt goldenposeidon.PoseidonHashOut) *goldenposeidon.PoseidonHashOut {
	const (
		int8Key = 8
		int4Key = 4
	)

	pubkeyBytes := SplitBigIntTo32BitChunks(publicKey.Pk.X.BigInt(new(big.Int)))
	if len(pubkeyBytes) < int8Key {
		pubkeyBytes = append(make([]uint32, int8Key-len(pubkeyBytes)), pubkeyBytes...)
	}
	if len(pubkeyBytes) > int8Key {
		panic("public key is too large")
	}

	buf := make([]ffg.Element, int8Key+int4Key)
	for i := 0; i < len(pubkeyBytes); i++ {
		buf[i].SetUint64(uint64(pubkeyBytes[i]))
	}
	for i := 0; i < len(salt.Elements); i++ {
		buf = append(buf, salt.Elements[i])
	}

	return goldenposeidon.HashNoPad(buf)
}

func (d *Deposit) Marshal() []byte {
	const (
		int4Key  = 4
		int32Key = 32
	)

	tokenIndexBytes := make([]byte, int4Key)
	binary.BigEndian.PutUint32(tokenIndexBytes, d.TokenIndex)
	amountBytes := make([]byte, int32Key)
	for i, v := range d.Amount.Bytes() {
		amountBytes[int32Key-1-i] = v
	}

	return append(
		append(append(d.Recipient.ToAddress().Bytes(), tokenIndexBytes...), amountBytes...),
		d.Salt.Marshal()...,
	)
}

func (d *Deposit) Unmarshal(data []byte) error {
	const (
		int4Key  = 4
		int32Key = 32
	)

	if len(data) < int4Key+int32Key+int32Key+int32Key {
		var ErrInvalidDepositData = errors.New("invalid deposit data")
		return ErrInvalidDepositData
	}

	recipientAddress, err := intMaxAcc.NewAddressFromBytes(data[:int32Key])
	if err != nil {
		ErrorInvalidRecipient := errors.New("failed to unmarshal recipient address")
		return errors.Join(ErrorInvalidRecipient, err)
	}
	d.Recipient, err = recipientAddress.Public()
	if err != nil {
		ErrorInvalidRecipient := errors.New("failed to unmarshal recipient public key")
		return errors.Join(ErrorInvalidRecipient, err)
	}
	d.TokenIndex = binary.BigEndian.Uint32(data[int32Key : int32Key+int4Key])
	d.Amount = new(big.Int).SetBytes(data[int32Key+int4Key : int32Key+int4Key+int32Key])
	d.Salt = new(goldenposeidon.PoseidonHashOut)
	if err = d.Salt.Unmarshal(data[int32Key+int4Key+int32Key : int32Key+int4Key+int32Key+int32Key]); err != nil {
		ErrorInvalidSalt := errors.New("failed to unmarshal salt")
		return errors.Join(ErrorInvalidSalt, err)
	}

	return nil
}
