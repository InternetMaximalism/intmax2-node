package types

import (
	"encoding/binary"
	"encoding/json"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	numAccountIDBytes       = 5
	numUint32Bytes          = 4
	NumAccountIDPackedBytes = intMaxTypes.NumOfSenders * numAccountIDBytes / numUint32Bytes
)

type AccountIdPacked [NumAccountIDPackedBytes]uint32

func (accountIDPacked *AccountIdPacked) Set(other *AccountIdPacked) *AccountIdPacked {
	if other == nil {
		accountIDPacked = nil
		return accountIDPacked
	}

	for i := 0; i < NumAccountIDPackedBytes; i++ {
		accountIDPacked[i] = other[i]
	}

	return accountIDPacked
}

func (accountIDPacked *AccountIdPacked) FromBytes(bytes []byte) {
	if len(bytes) > NumAccountIDPackedBytes*numUint32Bytes {
		panic("invalid bytes length")
	}

	if len(bytes) < NumAccountIDPackedBytes*numUint32Bytes {
		panic("invalid bytes length")
	}

	for i := 0; i < NumAccountIDPackedBytes; i++ {
		accountIDPacked[i] = binary.BigEndian.Uint32(bytes[i*numUint32Bytes : (i+1)*numUint32Bytes])
	}
}

func (accountIDPacked *AccountIdPacked) Bytes() []byte {
	bytes := make([]byte, NumAccountIDPackedBytes*numUint32Bytes)
	for i := 0; i < intMaxTypes.NumOfSenders; i++ {
		binary.BigEndian.PutUint32(bytes[i*numUint32Bytes:(i+1)*numUint32Bytes], accountIDPacked[i])
	}

	return bytes
}

func (accountIDPacked *AccountIdPacked) Hex() string {
	return hexutil.Encode(accountIDPacked.Bytes())
}

func (accountIDPacked *AccountIdPacked) FromHex(s string) error {
	bytes, err := hexutil.Decode(s)
	if err != nil {
		return err
	}

	accountIDPacked.FromBytes(bytes)
	return nil
}

func (accountIDPacked *AccountIdPacked) MarshalJSON() ([]byte, error) {
	return json.Marshal(accountIDPacked.Hex())
}

func (accountIDPacked *AccountIdPacked) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	return accountIDPacked.FromHex(s)
}

func (accountIDPacked *AccountIdPacked) Pack(accountIDs []uint64) *AccountIdPacked {
	accountIDsBytes := make([]byte, numAccountIDBytes*intMaxTypes.NumOfSenders)
	for i, accountID := range accountIDs {
		chunkBytes := make([]byte, Int8Key)
		binary.BigEndian.PutUint64(chunkBytes, accountID)
		copy(accountIDsBytes[i*numAccountIDBytes:(i+1)*numAccountIDBytes], chunkBytes[Int8Key-numAccountIDBytes:])
	}
	const defaultAccountID = uint64(1)
	for i := len(accountIDs); i < intMaxTypes.NumOfSenders; i++ {
		chunkBytes := make([]byte, Int8Key)
		binary.BigEndian.PutUint64(chunkBytes, defaultAccountID)
		copy(accountIDsBytes[i*numAccountIDBytes:(i+1)*numAccountIDBytes], chunkBytes[Int8Key-numAccountIDBytes:])
	}

	accountIDPacked.FromBytes(accountIDsBytes)

	return accountIDPacked
}

func (accountIDPacked *AccountIdPacked) Unpack() []uint64 {
	accountIDsBytes := accountIDPacked.Bytes()
	accountIDs := make([]uint64, 0)
	for i := 0; i < intMaxTypes.NumOfSenders; i++ {
		chunkBytes := make([]byte, Int8Key)
		copy(chunkBytes[Int8Key-numAccountIDBytes:], accountIDsBytes[i*numAccountIDBytes:(i+1)*numAccountIDBytes])

		accountID := binary.BigEndian.Uint64(chunkBytes)
		accountIDs = append(accountIDs, accountID)
	}

	return accountIDs
}

func (accountIDPacked *AccountIdPacked) Hash() intMaxTypes.Bytes32 {
	h := crypto.Keccak256(accountIDPacked.Bytes()) // TODO: Is this correct hash?
	var b intMaxTypes.Bytes32
	b.FromBytes(h)

	return b
}
