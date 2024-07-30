package types

import (
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
