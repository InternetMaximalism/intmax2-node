package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	int3Key  = 3
	int4Key  = 4
	int16Key = 16
	int20Key = 20
	int24Key = 24
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
	const mask = 0b11000000
	const flag = 0x80
	d := make([]byte, int32Key)
	copy(d[int32Key-len(ga.Address):], ga.Address)

	if ga.TypeOfAddress == intMaxAccTypes.INTMAXAddressType {
		if d[0]&mask != 0 {
			panic("address type is not INTMAX")
		}

		d[0] |= flag
	}

	return d
}

func (ga *GenericAddress) Unmarshal(data []byte) error {
	const (
		filter = 0b10000000
		flag   = 0b11000000
		mask   = 0b00111111
	)
	ga.Address = make([]byte, int32Key)
	copy(ga.Address[int32Key-len(data):], data)

	if ga.Address[0]&flag == filter {
		ga.TypeOfAddress = intMaxAccTypes.INTMAXAddressType
		ga.Address[0] &= mask
	} else {
		ga.TypeOfAddress = intMaxAccTypes.EthereumAddressType
		for i := 0; i < int32Key-int20Key; i++ {
			if ga.Address[i] != 0 {
				return errors.New("address invalid: not an Ethereum address")
			}
		}
	}

	return nil
}

func (ga *GenericAddress) String() string {
	return "0x" + hex.EncodeToString(ga.Marshal())
}

func (ga *GenericAddress) AddressType() string {
	return ga.TypeOfAddress
}

func (ga *GenericAddress) ToINTMAXAddress() (intMaxAcc.Address, error) {
	if ga.TypeOfAddress != intMaxAccTypes.INTMAXAddressType {
		return intMaxAcc.Address{}, errors.New("address is not INTMAX")
	}

	return intMaxAcc.NewAddressFromBytes(ga.Address)
}

func (ga *GenericAddress) ToEthereumAddress() (common.Address, error) {
	if ga.TypeOfAddress != intMaxAccTypes.EthereumAddressType {
		return common.Address{}, errors.New("address is not Ethereum")
	}

	return common.BytesToAddress(ga.Address), nil
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
		Salt:       salt,
	}
}

func NewTransferWithRandomSalt(recipient *GenericAddress, tokenIndex uint32, amount *big.Int) *Transfer {
	salt, err := new(PoseidonHashOut).SetRandom()
	if err != nil {
		panic(err)
	}

	return NewTransfer(recipient, tokenIndex, amount, salt)
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

func (td *Transfer) ToUint64Slice() []uint64 {
	isPubicKey := 0
	if td.Recipient == nil {
		panic("Recipient is nil")
	}
	if td.Recipient.AddressType() == intMaxAccTypes.INTMAXAddressType {
		isPubicKey = 1
	}

	recipientBytes := make([]byte, int32Key)
	copy(recipientBytes[int32Key-len(td.Recipient.Address):], td.Recipient.Address)

	amountBytes := make([]byte, int32Key)
	copy(amountBytes[int32Key-len(td.Amount.Bytes()):], td.Amount.Bytes())

	result := []uint64{uint64(isPubicKey)}
	result = append(result, BytesToUint64Array(recipientBytes)...)
	result = append(result, uint64(td.TokenIndex))
	result = append(result, BytesToUint64Array(amountBytes)...)
	for i := 0; i < int4Key; i++ {
		result = append(result, td.Salt.Elements[i].ToUint64Regular())
	}

	return result
}

func bytesToUint64(b []byte) uint64 {
	return uint64(b[0])<<int24Key | uint64(b[1])<<int16Key | uint64(b[2])<<int8Key | uint64(b[int3Key])
}

func BytesToUint64Array(b []byte) []uint64 {
	resultLength := (len(b) + int4Key - 1) / int4Key
	result := make([]uint64, resultLength)

	for i := 0; i < resultLength; i++ {
		start := i * int4Key
		end := start + int4Key

		if end > len(b) {
			end = len(b)
		}

		chunk := make([]byte, int4Key)
		copy(chunk, b[start:end])

		result[i] = bytesToUint64(chunk)
	}

	return result
}

func (td *Transfer) Marshal() []byte {
	recipientBytes := make([]byte, int32Key)
	copy(recipientBytes, td.Recipient.Marshal())
	tokenIndexBytes := make([]byte, int4Key)
	binary.BigEndian.PutUint32(tokenIndexBytes, td.TokenIndex)
	amountBytes := make([]byte, int32Key)
	tdAmountBytes := td.Amount.Bytes()
	copy(amountBytes[int32Key-len(tdAmountBytes):], tdAmountBytes)
	reversedAmountBytes := make([]byte, int32Key)
	for i := range amountBytes {
		reversedAmountBytes[i] = amountBytes[int32Key-1-i]
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	_, err := buf.Write(recipientBytes)
	if err != nil {
		panic(err)
	}

	_, err = buf.Write(tokenIndexBytes)
	if err != nil {
		panic(err)
	}

	_, err = buf.Write(reversedAmountBytes)
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
	td.Recipient = new(GenericAddress)
	a := buf.Next(int32Key)
	if err := td.Recipient.Unmarshal(a); err != nil {
		return err
	}

	tokenIndexBytes := make([]byte, int4Key)
	if _, err := buf.Read(tokenIndexBytes); err != nil {
		return err
	}
	td.TokenIndex = binary.BigEndian.Uint32(tokenIndexBytes)

	amountBytes := make([]byte, int32Key)
	if _, err := buf.Read(amountBytes); err != nil {
		return err
	}
	reversedAmountBytes := make([]byte, int32Key)
	for i := range amountBytes {
		reversedAmountBytes[i] = amountBytes[int32Key-1-i]
	}
	td.Amount = new(big.Int)
	td.Amount.SetBytes(reversedAmountBytes)

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
	uint64Slice := td.ToUint64Slice()
	inputs := make([]ffg.Element, len(uint64Slice))
	for i, v := range uint64Slice {
		inputs[i].SetUint64(v)
	}

	return goldenposeidon.HashNoPad(inputs)
}

func (td *Transfer) Nullifier() *PoseidonHashOut {
	return td.Hash()
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

func (td *Transfer) Commitment() *PoseidonHashOut {
	flatten := td.ToUint64Slice()

	inputs := make([]ffg.Element, len(flatten))
	for i := 0; i < len(flatten); i++ {
		inputs[i].SetUint64(flatten[i])
	}

	return goldenposeidon.HashNoPad(inputs)
}

func (td *Transfer) GetWithdrawalNullifier() *PoseidonHashOut {
	transferCommitment := td.Commitment()
	input := transferCommitment.Elements[:]
	input = append(input, td.Salt.Elements[:]...)
	return goldenposeidon.HashNoPad(input)
}

type TransferWitness struct {
	Transfer            Transfer
	TransferIndex       uint32
	Tx                  Tx
	TransferMerkleProof []*PoseidonHashOut
}

func (t *TransferWitness) Equal(other *TransferWitness) bool {
	if !t.Transfer.Equal(&other.Transfer) {
		return false
	}

	if t.TransferMerkleProof == nil && other.TransferMerkleProof == nil {
		return true
	}

	if t.TransferIndex != other.TransferIndex {
		return false
	}

	if !t.Tx.Equal(&other.Tx) {
		return false
	}

	if len(t.TransferMerkleProof) != len(other.TransferMerkleProof) {
		return false
	}

	for i := range t.TransferMerkleProof {
		if !t.TransferMerkleProof[i].Equal(other.TransferMerkleProof[i]) {
			return false
		}
	}

	return true
}

func (t *TransferWitness) Marshal() []byte {
	buf := bytes.NewBuffer(make([]byte, 0))

	if _, err := buf.Write(t.Transfer.Marshal()); err != nil {
		panic(err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.TransferIndex); err != nil {
		panic(err)
	}
	if _, err := buf.Write(t.Tx.Marshal()); err != nil {
		panic(err)
	}
	if err := binary.Write(buf, binary.BigEndian, uint32(len(t.TransferMerkleProof))); err != nil {
		panic(err)
	}
	for _, proof := range t.TransferMerkleProof {
		if _, err := buf.Write(proof.Marshal()); err != nil {
			panic(err)
		}
	}
	return buf.Bytes()
}

func (t *TransferWitness) Write(buf *bytes.Buffer) error {
	_, err := buf.Write(t.Marshal())

	return err
}

func (t *TransferWitness) Read(buf *bytes.Buffer) error {
	t.Transfer = *new(Transfer)
	if err := t.Transfer.Read(buf); err != nil {
		fmt.Printf("Reading TransferWitness\n")
		return err
	}

	if len(buf.Bytes()) == 0 {
		fmt.Printf("WARNING: only transfer data is available\n")
		return nil
	}

	transferIndexBytes := make([]byte, int4Key)
	if _, err := buf.Read(transferIndexBytes); err != nil {
		fmt.Println("Reading TransferIndex")
		return err
	}
	t.TransferIndex = binary.BigEndian.Uint32(transferIndexBytes)

	t.Tx = *new(Tx)
	if err := t.Tx.Read(buf); err != nil {
		fmt.Println("Reading Tx")
		return err
	}

	proofsLenBytes := make([]byte, int4Key)
	if _, err := buf.Read(proofsLenBytes); err != nil {
		fmt.Println("Reading TransferMerkleProof")
		return err
	}

	proofsLen := binary.BigEndian.Uint32(proofsLenBytes)
	if int(proofsLen) > buf.Len() {
		fmt.Printf("Reading TransferMerkleProof length: %d\n", proofsLen)
		return fmt.Errorf("invalid TransferMerkleProof length: %d", proofsLen)
	}
	t.TransferMerkleProof = make([]*PoseidonHashOut, proofsLen)
	for i := range t.TransferMerkleProof {
		t.TransferMerkleProof[i] = new(PoseidonHashOut)
		if err := t.TransferMerkleProof[i].Unmarshal(buf.Next(int32Key)); err != nil {
			fmt.Println("Reading TransferMerkleProof")
			return err
		}
	}

	return nil
}

func (t *TransferWitness) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	return t.Read(buf)
}

type TransferDetails struct {
	TransferWitness                     *TransferWitness
	TxTreeRoot                          *PoseidonHashOut
	TxMerkleProof                       []*PoseidonHashOut
	SenderLastBalancePublicInputs       []byte
	SenderBalanceTransitionPublicInputs []byte
	SenderEnoughBalanceProofBodyHash    string
}

func (t *TransferDetails) Equal(other *TransferDetails) bool {
	if !t.TransferWitness.Equal(other.TransferWitness) {
		return false
	}

	if t.TxTreeRoot == nil && other.TxTreeRoot == nil {
		return true
	}

	if !t.TxTreeRoot.Equal(other.TxTreeRoot) {
		return false
	}

	if len(t.TxMerkleProof) != len(other.TxMerkleProof) {
		return false
	}

	for i := range t.TxMerkleProof {
		if !t.TxMerkleProof[i].Equal(other.TxMerkleProof[i]) {
			return false
		}
	}

	if !bytes.Equal(t.SenderLastBalancePublicInputs, other.SenderLastBalancePublicInputs) {
		return false
	}

	if !bytes.Equal(t.SenderBalanceTransitionPublicInputs, other.SenderBalanceTransitionPublicInputs) {
		return false
	}

	return t.SenderEnoughBalanceProofBodyHash == other.SenderEnoughBalanceProofBodyHash
}

func (t *TransferDetails) Marshal() []byte {
	buf := bytes.NewBuffer(make([]byte, 0))

	if _, err := buf.Write(t.TransferWitness.Marshal()); err != nil {
		panic(err)
	}

	if _, err := buf.Write(t.TxTreeRoot.Marshal()); err != nil {
		panic(err)
	}

	proofsLen := uint32(len(t.TxMerkleProof))
	if err := binary.Write(buf, binary.BigEndian, proofsLen); err != nil {
		panic(err)
	}
	for _, proof := range t.TxMerkleProof {
		if _, err := buf.Write(proof.Marshal()); err != nil {
			panic(err)
		}
	}

	if err := binary.Write(buf, binary.BigEndian, uint32(len(t.SenderLastBalancePublicInputs))); err != nil {
		panic(err)
	}
	if _, err := buf.Write(t.SenderLastBalancePublicInputs); err != nil {
		panic(err)
	}

	if err := binary.Write(buf, binary.BigEndian, uint32(len(t.SenderBalanceTransitionPublicInputs))); err != nil {
		panic(err)
	}
	if _, err := buf.Write(t.SenderBalanceTransitionPublicInputs); err != nil {
		panic(err)
	}

	senderEnoughBalanceProofBodyHash, err := hexutil.Decode(t.SenderEnoughBalanceProofBodyHash)
	if err != nil {
		panic(err)
	}
	if _, err := buf.Write(senderEnoughBalanceProofBodyHash); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (t *TransferDetails) Write(buf *bytes.Buffer) error {
	_, err := buf.Write(t.Marshal())

	return err
}

func (t *TransferDetails) Read(buf *bytes.Buffer) error {
	t.TransferWitness = new(TransferWitness)
	if err := t.TransferWitness.Read(buf); err != nil {
		return err
	}

	if len(buf.Bytes()) == 0 {
		fmt.Printf("WARNING: only transfer witness is available\n")
		return nil
	}

	t.TxTreeRoot = new(PoseidonHashOut)
	if err := t.TxTreeRoot.Unmarshal(buf.Next(int32Key)); err != nil {
		return err
	}

	proofsLenBytes := make([]byte, int4Key)
	if _, err := buf.Read(proofsLenBytes); err != nil {
		return err
	}
	proofsLen := binary.BigEndian.Uint32(proofsLenBytes)
	t.TxMerkleProof = make([]*PoseidonHashOut, proofsLen)
	for i := range t.TxMerkleProof {
		t.TxMerkleProof[i] = new(PoseidonHashOut)
		if err := t.TxMerkleProof[i].Unmarshal(buf.Next(int32Key)); err != nil {
			fmt.Println("Error reading TxMerkleProof")
			return err
		}
	}

	senderLastBalancePublicInputsLenBytes := make([]byte, int4Key)
	if _, err := buf.Read(senderLastBalancePublicInputsLenBytes); err != nil {
		return err
	}
	senderLastBalancePublicInputsLen := binary.BigEndian.Uint32(senderLastBalancePublicInputsLenBytes)
	t.SenderLastBalancePublicInputs = make([]byte, senderLastBalancePublicInputsLen)
	if _, err := buf.Read(t.SenderLastBalancePublicInputs); err != nil {
		return err
	}

	senderBalanceTransitionPublicInputsLenBytes := make([]byte, int4Key)
	if _, err := buf.Read(senderBalanceTransitionPublicInputsLenBytes); err != nil {
		return err
	}
	senderBalanceTransitionPublicInputsLen := binary.BigEndian.Uint32(senderBalanceTransitionPublicInputsLenBytes)
	t.SenderBalanceTransitionPublicInputs = make([]byte, senderBalanceTransitionPublicInputsLen)
	if _, err := buf.Read(t.SenderBalanceTransitionPublicInputs); err != nil {
		return err
	}

	if len(buf.Bytes()) == 0 {
		fmt.Printf("WARNING: sender enough balance proof is not available\n")
		return nil
	}

	senderEnoughBalanceProofBodyHash := buf.Next(int32Key)
	t.SenderEnoughBalanceProofBodyHash = hexutil.Encode(senderEnoughBalanceProofBodyHash)

	return nil
}

func (t *TransferDetails) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	return t.Read(buf)
}

type TransferDetailsWithProofBody struct {
	TransferDetails                  *TransferDetails
	SenderLastBalanceProofBody       string
	SenderBalanceTransitionProofBody string
}
