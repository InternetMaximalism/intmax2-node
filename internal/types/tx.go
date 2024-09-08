package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type Tx struct {
	TransferTreeRoot *PoseidonHashOut
	Nonce            uint64
}

func NewTx(transferTreeRoot *PoseidonHashOut, nonce uint64) (*Tx, error) {
	if nonce > ffg.Modulus().Uint64() {
		return nil, ErrNonceTooLarge
	}

	t := new(Tx)
	t.Nonce = nonce
	t.TransferTreeRoot = new(PoseidonHashOut).Set(transferTreeRoot)

	return t, nil
}

func (t *Tx) Set(tx *Tx) *Tx {
	if t == nil {
		t = new(Tx)
	}

	t.Nonce = tx.Nonce
	t.TransferTreeRoot = new(PoseidonHashOut).Set(tx.TransferTreeRoot)

	return t
}

func (t *Tx) SetZero() *Tx {
	t.Nonce = 0
	t.TransferTreeRoot = new(PoseidonHashOut).SetZero()

	return t
}

func (t *Tx) Equal(tx *Tx) bool {
	return t.Nonce == tx.Nonce && t.TransferTreeRoot.Equal(tx.TransferTreeRoot)
}

// // SetRandom return Tx
// // Testing purposes only
// func (t *Tx) SetRandom() (*Tx, error) {
// 	var err error

// 	t.Transfers, err = new(PoseidonHashOut).SetRandom()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return t, nil
// }

func (t *Tx) ToFieldElementSlice() []ffg.Element {
	const (
		int0Key = 0
		int4Key = 4
	)
	result := make([]ffg.Element, int4Key+1)
	for i := int0Key; i < goldenposeidon.NUM_HASH_OUT_ELTS; i++ {
		result[i].Set(&t.TransferTreeRoot.Elements[i])
	}
	result[int4Key].SetUint64(t.Nonce)

	return result
}

func (t *Tx) Hash() *PoseidonHashOut {
	input := t.ToFieldElementSlice()
	return goldenposeidon.HashNoPad(input)
}

func (t *Tx) Marshal() []byte {
	buf := bytes.NewBuffer(make([]byte, 0))

	if err := binary.Write(buf, binary.BigEndian, t.Nonce); err != nil {
		panic(err)
	}
	if _, err := buf.Write(t.TransferTreeRoot.Marshal()); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (t *Tx) Write(buf *bytes.Buffer) error {
	_, err := buf.Write(t.Marshal())

	return err
}

func (t *Tx) Read(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &t.Nonce); err != nil {
		return err
	}

	t.TransferTreeRoot = new(PoseidonHashOut)
	return t.TransferTreeRoot.Unmarshal(buf.Next(int32Key))
}

func (t *Tx) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	return t.Read(buf)
}

type TxDetails struct {
	Tx
	TxTreeRoot    *goldenposeidon.PoseidonHashOut
	TxMerkleProof []*goldenposeidon.PoseidonHashOut
	Transfers     []*Transfer
}

func (td *TxDetails) Marshal() []byte {
	buf := bytes.NewBuffer(make([]byte, 0))

	if _, err := buf.Write(td.TransferTreeRoot.Marshal()); err != nil {
		panic(err)
	}
	if err := binary.Write(buf, binary.BigEndian, td.Nonce); err != nil {
		panic(err)
	}
	if err := binary.Write(buf, binary.BigEndian, uint32(len(td.Transfers))); err != nil {
		panic(err)
	}

	for _, transfer := range td.Transfers {
		if _, err := buf.Write(transfer.Marshal()); err != nil {
			panic(err)
		}
	}

	if _, err := buf.Write(td.TxTreeRoot.Marshal()); err != nil {
		panic(err)
	}

	if err := binary.Write(buf, binary.BigEndian, uint32(len(td.TxMerkleProof))); err != nil {
		panic(err)
	}

	for _, proof := range td.TxMerkleProof {
		if _, err := buf.Write(proof.Marshal()); err != nil {
			panic(err)
		}
	}

	return buf.Bytes()
}

func (td *TxDetails) Write(buf *bytes.Buffer) error {
	_, err := buf.Write(td.Marshal())

	return err
}

func (td *TxDetails) Read(buf *bytes.Buffer) error {
	const int32Key = 32

	transferTreeRoot := new(PoseidonHashOut)
	if err := transferTreeRoot.Unmarshal(buf.Next(int32Key)); err != nil {
		var ErrUnmarshalTransferTreeRoot = fmt.Errorf("failed to unmarshal transfer tree root: %w", err)
		return errors.Join(ErrUnmarshalTransferTreeRoot, err)
	}
	td.TransferTreeRoot = new(PoseidonHashOut).Set(transferTreeRoot)

	if err := binary.Read(buf, binary.BigEndian, &td.Nonce); err != nil {
		return err
	}
	var numTransfers uint32
	if err := binary.Read(buf, binary.BigEndian, &numTransfers); err != nil {
		return err
	}

	td.Transfers = make([]*Transfer, numTransfers)
	for i := 0; i < int(numTransfers); i++ {
		transfer := new(Transfer)
		if err := transfer.Read(buf); err != nil {
			return err
		}
		td.Transfers[i] = transfer
	}

	if len(buf.Bytes()) == 0 {
		return nil
	}

	txTreeRoot := new(PoseidonHashOut)
	if err := txTreeRoot.Unmarshal(buf.Next(int32Key)); err != nil {
		return err
	}

	td.TxTreeRoot = new(PoseidonHashOut).Set(txTreeRoot)

	var numTxMerkleProof uint32
	if err := binary.Read(buf, binary.BigEndian, &numTxMerkleProof); err != nil {
		return err
	}

	td.TxMerkleProof = make([]*PoseidonHashOut, numTxMerkleProof)
	for i := 0; i < int(numTxMerkleProof); i++ {
		proof := new(PoseidonHashOut)
		if err := proof.Unmarshal(buf.Next(int32Key)); err != nil {
			return err
		}
		td.TxMerkleProof[i] = proof
	}

	return nil
}

func (td *TxDetails) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)
	return td.Read(buf)
}
