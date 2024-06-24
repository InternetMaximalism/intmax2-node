package tree

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	// EthereumAddress represents an Ethereum address type
	EthereumAddressType = "ETHEREUM"
	// INTMAXAddress represents an INTMAX address type
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

func NewEthereumAddress(address []byte) (*GenericAddress, error) {
	if len(address) != 20 {
		return nil, errors.New("the Ethereum address should be 20 bytes")
	}

	return &GenericAddress{
		addressType: EthereumAddressType,
		address:     address,
	}, nil
}

func NewINTMAXAddress(address []byte) (*GenericAddress, error) {
	if len(address) != 32 {
		return nil, errors.New("the INTMAX address should be 32 bytes")
	}

	return &GenericAddress{
		addressType: EthereumAddressType,
		address:     address,
	}, nil
}

type Transfer struct {
	Recipient  GenericAddress
	TokenIndex uint32
	Amount     *big.Int
	Salt       *poseidonHashOut
}

func (td *Transfer) Set(transferData *Transfer) *Transfer {
	td.Recipient = transferData.Recipient
	td.TokenIndex = transferData.TokenIndex
	td.Amount = transferData.Amount
	td.Salt = transferData.Salt
	return td
}

func (td *Transfer) Marshal() []byte {
	tokenIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tokenIndexBytes, td.TokenIndex)

	return append(append(append(td.Recipient.Marshal(), tokenIndexBytes...), td.Amount.Bytes()...), td.Salt.Marshal()...)
}

func (td *Transfer) ToFieldElementSlice() []*ffg.Element {
	return finite_field.BytesToFieldElementSlice(td.Marshal())
}

func (td *Transfer) Hash() *poseidonHashOut {
	return goldenposeidon.HashNoPad(td.ToFieldElementSlice())
}

func (td *Transfer) Equal(other *Transfer) bool {
	if !td.Recipient.Equal(&other.Recipient) {
		return false
	}
	if td.TokenIndex != other.TokenIndex {
		return false
	}
	if td.Amount.Cmp(other.Amount) != 0 {
		return false
	}
	if !td.Salt.Equal(other.Salt) {
		return false
	}
	return true
}

type TransferTree struct {
	Leaves []*Transfer
	inner  *PoseidonMerkleTree
}

func NewTransferTree(height uint8, initialLeaves []*Transfer, zeroHash *poseidonHashOut) (*TransferTree, error) {
	initialLeafHashes := make([]*poseidonHashOut, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	t, err := NewPoseidonMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, err
	}

	leaves := make([]*Transfer, len(initialLeaves))
	for i, leaf := range initialLeaves {
		leaves[i] = new(Transfer).Set(leaf)
	}

	return &TransferTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *TransferTree) BuildMerkleRoot(leaves []*poseidonHashOut) (*poseidonHashOut, error) {
	return t.inner.BuildMerkleRoot(leaves)
}

func (t *TransferTree) GetCurrentRootCountAndSiblings() (poseidonHashOut, uint64, []*poseidonHashOut) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *TransferTree) AddLeaf(index uint64, leaf *Transfer) (root *poseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return nil, err
	}

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of leaves")
	}
	t.Leaves = append(t.Leaves, new(Transfer).Set(leaf))

	return root, nil
}

func (t *TransferTree) ComputeMerkleProof(index uint64, leaves []*poseidonHashOut) (siblings []*poseidonHashOut, root poseidonHashOut, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
