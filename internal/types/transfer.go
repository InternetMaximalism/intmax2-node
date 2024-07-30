package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

// GenericAddress struct to hold address and its type
//
// TODO: Implement MarshalJSON and UnmarshalJSON methods for GenericAddress
type GenericAddress struct {
	// TypeOfAddress can be "ETHEREUM" or "INTMAX"
	TypeOfAddress string
	// If TypeOfAddress is ETHEREUM, then the address should be a 20-byte value.
	// If TypeOfAddress is INTMAX, then the address should be a 32-byte value.
	Address []byte
}

func (ga *GenericAddress) Set(genericAddress *GenericAddress) *GenericAddress {
	ga.TypeOfAddress = genericAddress.TypeOfAddress
	copy(ga.Address, genericAddress.Address)
	return ga
}

func (ga *GenericAddress) Marshal() []byte {
	return ga.Address
}

func (ga *GenericAddress) Unmarshal(data []byte) error {
	const int20Key = 20
	const int32Key = 32

	switch len(data) {
	case int20Key:
		ga.TypeOfAddress = intMaxAccTypes.EthereumAddressType
	case int32Key:
		ga.TypeOfAddress = intMaxAccTypes.INTMAXAddressType
	default:
		return errors.New("address invalid")
	}

	ga.Address = make([]byte, len(data))
	copy(ga.Address, data)
	return nil
}

func (ga *GenericAddress) String() string {
	return "0x" + hex.EncodeToString(ga.Marshal())
}

func (ga *GenericAddress) AddressType() string {
	return ga.TypeOfAddress
}

func (ga *GenericAddress) Equal(other *GenericAddress) bool {
	if ga.TypeOfAddress != other.TypeOfAddress {
		return false
	}
	if len(ga.Address) != len(other.Address) {
		return false
	}
	for i := range ga.Address {
		if ga.Address[i] != other.Address[i] {
			return false
		}
	}
	return true
}

func NewDefaultGenericAddress() *GenericAddress {
	const int20Key = 20
	defaultAddress := [int20Key]byte{}

	return &GenericAddress{
		TypeOfAddress: intMaxAccTypes.EthereumAddressType,
		Address:       defaultAddress[:],
	}
}

func NewEthereumAddress(address []byte) (*GenericAddress, error) {
	const int20Key = 20
	if len(address) != int20Key {
		return nil, ErrETHAddressInvalid
	}

	return &GenericAddress{
		TypeOfAddress: intMaxAccTypes.EthereumAddressType,
		Address:       address,
	}, nil
}

func NewINTMAXAddress(address []byte) (*GenericAddress, error) {
	const int32Key = 32
	if len(address) != int32Key {
		return nil, ErrINTMAXAddressInvalid
	}

	return &GenericAddress{
		TypeOfAddress: intMaxAccTypes.INTMAXAddressType,
		Address:       address,
	}, nil
}

type Transfer struct {
	Recipient  *GenericAddress
	TokenIndex uint32
	Amount     *big.Int
	Salt       *PoseidonHashOut
}

func NewTransfer(recipient *GenericAddress, tokenIndex uint32, amount *big.Int, salt *PoseidonHashOut) *Transfer {
	return &Transfer{
		Recipient:  recipient,
		TokenIndex: tokenIndex,
		Amount:     amount,
	}
}

func NewTransferWithRandomSalt(recipient *GenericAddress, tokenIndex uint32, amount *big.Int) *Transfer {
	salt, err := new(PoseidonHashOut).SetRandom()
	if err != nil {
		panic(err)
	}

	return &Transfer{
		Recipient:  recipient,
		TokenIndex: tokenIndex,
		Amount:     amount,
		Salt:       salt,
	}
}

func (td *Transfer) Set(transferData *Transfer) *Transfer {
	td.Recipient = new(GenericAddress).Set(transferData.Recipient)
	td.TokenIndex = transferData.TokenIndex
	td.Amount = new(big.Int).Set(transferData.Amount)
	td.Salt = new(PoseidonHashOut).Set(transferData.Salt)
	return td
}

func (td *Transfer) SetZero() *Transfer {
	td.Recipient = NewDefaultGenericAddress()
	td.TokenIndex = 0
	td.Amount = big.NewInt(0)
	td.Salt = new(PoseidonHashOut).SetZero()
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

	buf := bytes.NewBuffer(make([]byte, 0))
	recipientBytes := td.Recipient.Marshal()
	err := binary.Write(buf, binary.BigEndian, uint32(len(recipientBytes)))
	if err != nil {
		panic(err)
	}
	_, err = buf.Write(recipientBytes)
	if err != nil {
		panic(err)
	}
	_, err = buf.Write(tokenIndexBytes)
	if err != nil {
		panic(err)
	}
	_, err = buf.Write(amountBytes)
	if err != nil {
		panic(err)
	}
	_, err = buf.Write(td.Salt.Marshal())
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (td *Transfer) Write(buf *bytes.Buffer) error {
	_, err := buf.Write(td.Marshal())

	return err
}

func (td *Transfer) Read(buf *bytes.Buffer) error {
	const int32Key = 32

	recipientBytes := make([]byte, int32Key)
	if _, err := buf.Read(recipientBytes); err != nil {
		return err
	}
	td.Recipient = new(GenericAddress)
	if err := td.Recipient.Unmarshal(recipientBytes); err != nil {
		return err
	}

	tokenIndexBytes := make([]byte, int32Key)
	if _, err := buf.Read(tokenIndexBytes); err != nil {
		return err
	}
	td.TokenIndex = binary.BigEndian.Uint32(tokenIndexBytes)

	amountBytes := make([]byte, int32Key)
	if _, err := buf.Read(amountBytes); err != nil {
		return err
	}
	td.Amount = new(big.Int)
	td.Amount.SetBytes(amountBytes)

	saltBytes := make([]byte, int32Key)
	if _, err := buf.Read(saltBytes); err != nil {
		return err
	}
	td.Salt = new(PoseidonHashOut)
	if err := td.Salt.Unmarshal(saltBytes); err != nil {
		return err
	}

	return nil
}

func (td *Transfer) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)
	return td.Read(buf)
}

func (td *Transfer) ToFieldElementSlice() []ffg.Element {
	return finite_field.BytesToFieldElementSlice(td.Marshal())
}

func (td *Transfer) Hash() *PoseidonHashOut {
	return goldenposeidon.HashNoPad(td.ToFieldElementSlice())
}

func (td *Transfer) Equal(other *Transfer) bool {
	switch {
	case !td.Recipient.Equal(other.Recipient),
		td.TokenIndex != other.TokenIndex,
		td.Amount.Cmp(other.Amount) != 0,
		!td.Salt.Equal(other.Salt):
		return false
	default:
		return true
	}
}
