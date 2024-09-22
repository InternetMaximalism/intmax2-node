package block_synchronizer

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/iden3/go-iden3-crypto/ffg"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	int4Key = 4
	int8Key = 8
)

type poseidonHashOut = intMaxGP.PoseidonHashOut

type BalanceData struct {
	BalanceProofPublicInputs []ffg.Element
	NullifierLeaves          []intMaxTypes.Bytes32
	AssetLeafEntries         []*intMaxTree.AssetLeafEntry
	Nonce                    uint32
	Salt                     poseidonHashOut
	PublicState              *block_validity_prover.PublicState
}

func (bd *BalanceData) Set(other *BalanceData) *BalanceData {
	bd.BalanceProofPublicInputs = make([]ffg.Element, len(other.BalanceProofPublicInputs))
	copy(bd.BalanceProofPublicInputs, other.BalanceProofPublicInputs)

	bd.NullifierLeaves = make([]intMaxTypes.Bytes32, len(other.NullifierLeaves))
	copy(bd.NullifierLeaves, other.NullifierLeaves)

	bd.AssetLeafEntries = make([]*intMaxTree.AssetLeafEntry, len(other.AssetLeafEntries))
	for i := range other.AssetLeafEntries {
		bd.AssetLeafEntries[i] = new(intMaxTree.AssetLeafEntry).Set(other.AssetLeafEntries[i])
	}

	bd.Nonce = other.Nonce
	bd.Salt.Set(&other.Salt)
	bd.PublicState = new(block_validity_prover.PublicState).Set(other.PublicState)

	return bd
}

func (bd *BalanceData) Marshal() ([]byte, error) {
	bufSize := int4Key + int32Key + int4Key + int8Key*len(bd.BalanceProofPublicInputs) +
		int4Key + int32Key*len(bd.NullifierLeaves) + int4Key + (int32Key+1)*len(bd.AssetLeafEntries) +
		block_validity_prover.NumPublicStateBytes
	buf := make([]byte, bufSize)
	offset := 0

	binary.BigEndian.PutUint32(buf[offset:offset+int4Key], bd.Nonce)
	offset += int4Key

	b := bd.Salt.Marshal()
	copy(buf[offset:offset+int32Key], b)
	offset += int32Key

	binary.BigEndian.PutUint32(buf[offset:offset+int4Key], uint32(len(bd.BalanceProofPublicInputs)))
	offset += int4Key
	for _, publicInput := range bd.BalanceProofPublicInputs {
		b := publicInput.ToUint64Regular()
		binary.BigEndian.PutUint64(buf[offset:offset+int8Key], b)
		offset += int8Key
	}

	binary.BigEndian.PutUint32(buf[offset:offset+int4Key], uint32(len(bd.NullifierLeaves)))
	offset += int4Key
	for _, nullifierLeaf := range bd.NullifierLeaves {
		b := nullifierLeaf.Bytes()
		copy(buf[offset:offset+int32Key], b)
		offset += int32Key
	}

	binary.BigEndian.PutUint32(buf[offset:offset+int4Key], uint32(len(bd.AssetLeafEntries)))
	offset += int4Key
	for _, assetLeafEntry := range bd.AssetLeafEntries {
		b := assetLeafEntry.Marshal()
		copy(buf[offset:offset+int32Key+1], b)
		offset += int32Key + 1
	}

	copy(buf[offset:offset+block_validity_prover.NumPublicStateBytes], bd.PublicState.Marshal())

	return buf, nil
}

func (bd *BalanceData) Unmarshal(data []byte) error {
	if len(data) < int4Key {
		return errors.New("invalid data length")
	}

	offset := 0
	bd.Nonce = binary.BigEndian.Uint32(data[offset : offset+int4Key])
	offset += int4Key

	b := new(poseidonHashOut)
	err := b.Unmarshal(data[offset : offset+int32Key])
	if err != nil {
		return err
	}

	bd.Salt = *b
	offset += int32Key

	if len(data) < offset+int4Key {
		return errors.New("invalid data length")
	}

	numBalanceProofPublicInputs := binary.BigEndian.Uint32(data[offset : offset+int4Key])
	offset += int4Key

	bd.BalanceProofPublicInputs = make([]ffg.Element, numBalanceProofPublicInputs)
	for i := 0; i < int(numBalanceProofPublicInputs); i++ {
		if len(data) < offset+int8Key {
			return errors.New("invalid data length")
		}

		bd.BalanceProofPublicInputs[i].SetUint64(binary.BigEndian.Uint64(data[offset : offset+int8Key]))
		offset += int8Key
	}

	if len(data) < offset+int4Key {
		return errors.New("invalid data length")
	}

	numNullifierLeaves := binary.BigEndian.Uint32(data[offset : offset+int4Key])
	offset += int4Key

	bd.NullifierLeaves = make([]intMaxTypes.Bytes32, numNullifierLeaves)
	for i := 0; i < int(numNullifierLeaves); i++ {
		if len(data) < offset+int32Key {
			return errors.New("invalid data length")
		}

		bd.NullifierLeaves[i] = intMaxTypes.Bytes32{}
		bd.NullifierLeaves[i].FromBytes(data[offset : offset+int32Key])
		offset += int32Key
	}

	if len(data) < offset+int4Key {
		return errors.New("invalid data length")
	}

	numAssetLeaves := binary.BigEndian.Uint32(data[offset : offset+int4Key])
	offset += int4Key

	bd.AssetLeafEntries = make([]*intMaxTree.AssetLeafEntry, numAssetLeaves)
	for i := 0; i < int(numAssetLeaves); i++ {
		if len(data) < offset+int32Key+1 {
			return errors.New("invalid data length")
		}

		assetLeafEntry, err := new(intMaxTree.AssetLeafEntry).Unmarshal(data[offset : offset+int32Key+1])
		if err != nil {
			return err
		}

		bd.AssetLeafEntries[i] = assetLeafEntry
		offset += int32Key + 1
	}

	if len(data) < offset+block_validity_prover.NumPublicStateBytes {
		return errors.New("invalid data length")
	}

	bd.PublicState = new(block_validity_prover.PublicState)
	err = bd.PublicState.Unmarshal(data[offset : offset+block_validity_prover.NumPublicStateBytes])
	if err != nil {
		return err
	}

	return nil
}

func (bd *BalanceData) Encrypt(intMaxPublicKey *intMaxAcc.PublicKey) (string, error) {
	encodedBalanceData, err := bd.Marshal()
	if err != nil {
		return "", err
	}

	encryptedBalanceData, err := intMaxAcc.EncryptECIES(rand.Reader, intMaxPublicKey, encodedBalanceData)
	if err != nil {
		return "", err
	}

	result := base64.StdEncoding.EncodeToString(encryptedBalanceData)

	return result, nil
}

func (bd *BalanceData) Decrypt(intMaxPrivateKey *intMaxAcc.PrivateKey, encryptedBalanceData string) error {
	if encryptedBalanceData == "" {
		var ErrEmptyEncryptedBalanceData = errors.New("empty encrypted balance data")
		return ErrEmptyEncryptedBalanceData
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBalanceData)

	if err != nil {
		var ErrFailedToDecodeFromBase64 = errors.New("failed to decode from base64")
		return errors.Join(ErrFailedToDecodeFromBase64, err)
	}

	var message []byte
	message, err = intMaxPrivateKey.DecryptECIES(ciphertext)
	if err != nil {
		var ErrFailedToDecrypt = errors.New("failed to decrypt")
		return errors.Join(ErrFailedToDecrypt, err)
	}

	err = bd.Unmarshal(message)
	if err != nil {
		var ErrFailedToUnmarshal = errors.New("failed to unmarshal")
		return errors.Join(ErrFailedToUnmarshal, err)
	}

	return nil
}

type BackupBalanceData struct {
	ID                   string
	BalanceProofBody     string
	EncryptedBalanceData string
	EncryptedTxs         []string
	EncryptedTransfers   []string
	EncryptedDeposits    []string
	BlockNumber          uint64
	CreatedAt            *timestamppb.Timestamp
}

type GetBackupBalanceError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetBackupBalanceResponse struct {
	Success bool                   `json:"success"`
	Data    *BackupBalanceData     `json:"data,omitempty"`
	Error   *GetBackupBalanceError `json:"error,omitempty"`
}
