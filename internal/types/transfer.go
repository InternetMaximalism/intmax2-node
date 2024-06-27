package types

import (
	"encoding/binary"
	"encoding/hex"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	// EthereumAddressType represents an Ethereum address type
	EthereumAddressType = "ETHEREUM"
	// INTMAXAddressType represents an INTMAX address type
	INTMAXAddressType = "INTMAX"
)

// GenericAddress struct to hold address and its type
type GenericAddress struct {
	// AddressType can be "ETHEREUM" or "INTMAX"
	addressType string
	// If AddressType is ETHEREUM, then the address should be a 20-byte value.
	// If AddressType is INTMAX, then the address should be a 32-byte value.
	address []byte
}

func (ga *GenericAddress) Marshal() []byte {
	return ga.address
}

func (ga *GenericAddress) String() string {
	return "0x" + hex.EncodeToString(ga.Marshal())
}

func (ga *GenericAddress) AddressType() string {
	return ga.addressType
}

func (ga *GenericAddress) Equal(other *GenericAddress) bool {
	if ga.addressType != other.addressType {
		return false
	}
	if len(ga.address) != len(other.address) {
		return false
	}
	for i := range ga.address {
		if ga.address[i] != other.address[i] {
			return false
		}
	}
	return true
}

func NewEthereumAddress(address []byte) (GenericAddress, error) {
	const int20Key = 20
	if len(address) != int20Key {
		return GenericAddress{}, ErrETHAddressInvalid
	}

	return GenericAddress{
		addressType: EthereumAddressType,
		address:     address,
	}, nil
}

func NewINTMAXAddress(address []byte) (GenericAddress, error) {
	const int32Key = 32
	if len(address) != int32Key {
		return GenericAddress{}, ErrINTMAXAddressInvalid
	}

	return GenericAddress{
		addressType: INTMAXAddressType,
		address:     address,
	}, nil
}

type Transfer struct {
	Recipient  GenericAddress
	TokenIndex uint32
	Amount     *big.Int
	Salt       *PoseidonHashOut
}

func (td *Transfer) Set(transferData *Transfer) *Transfer {
	td.Recipient = transferData.Recipient
	td.TokenIndex = transferData.TokenIndex
	td.Amount = transferData.Amount
	td.Salt = transferData.Salt
	return td
}

func (td *Transfer) Marshal() []byte {
	const (
		int4Key  = 4
		int31Key = 31
		int32Key = 32
	)

	tokenIndexBytes := make([]byte, int4Key)
	binary.BigEndian.PutUint32(tokenIndexBytes, td.TokenIndex)
	amountBytes := make([]byte, int32Key)
	for i, v := range td.Amount.Bytes() {
		amountBytes[int31Key-i] = v
	}

	return append(append(
		append(td.Recipient.Marshal(), tokenIndexBytes...),
		amountBytes...,
	), td.Salt.Marshal()...)
}

func (td *Transfer) ToFieldElementSlice() []*ffg.Element {
	return finite_field.BytesToFieldElementSlice(td.Marshal())
}

func (td *Transfer) Hash() *PoseidonHashOut {
	return goldenposeidon.HashNoPad(td.ToFieldElementSlice())
}

func (td *Transfer) Equal(other *Transfer) bool {
	switch {
	case !td.Recipient.Equal(&other.Recipient),
		td.TokenIndex != other.TokenIndex,
		td.Amount.Cmp(other.Amount) != 0,
		!td.Salt.Equal(other.Salt):
		return false
	default:
		return true
	}
}
