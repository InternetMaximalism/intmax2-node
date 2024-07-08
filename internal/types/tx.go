package types

import "errors"

// "intmax2-node/internal/tree"

const (
	txTreeHeight    = 7
	numTxTreeLeaves = 1 << txTreeHeight
)

type Tx struct {
	Nonce     uint64
	PowNonce  uint64
	Transfers [numTxTreeLeaves]Transfer
}

func NewTxWithoutTransfers(nonce, powNonce uint64) *Tx {
	t := new(Tx)
	t.Nonce = nonce
	t.PowNonce = powNonce
	t.Transfers = [numTxTreeLeaves]Transfer{}
	for i := 0; i < numTxTreeLeaves; i++ {
		t.Transfers[i].SetZero()
	}

	return t
}

func NewTxWithPartialTransfers(nonce, powNonce uint64, partialTransfers []Transfer) (*Tx, error) {
	if len(partialTransfers) > numTxTreeLeaves {
		var ErrTooManyTransfers = errors.New("too many transfers")
		return nil, ErrTooManyTransfers
	}

	t := new(Tx)
	t.Nonce = nonce
	t.PowNonce = powNonce
	t.Transfers = [numTxTreeLeaves]Transfer{}
	for i := 0; i < len(partialTransfers); i++ {
		t.Transfers[i].Set(new(Transfer).Set(&partialTransfers[i]))
	}
	for i := len(partialTransfers); i < numTxTreeLeaves; i++ {
		t.Transfers[i].SetZero()
	}

	return t, nil
}

func (t *Tx) Set(tx *Tx) *Tx {
	t.Nonce = tx.Nonce
	t.PowNonce = tx.PowNonce
	copy(t.Transfers[:], tx.Transfers[:])

	return t
}

func (t *Tx) SetZero() *Tx {
	t.Nonce = 0
	t.PowNonce = 0
	t.Transfers = [numTxTreeLeaves]Transfer{}
	for i := 0; i < numTxTreeLeaves; i++ {
		t.Transfers[i].SetZero()
	}

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

// func (t *Tx) ToFieldElementSlice() []*ffg.Element {
// 	const (
// 		int0Key = 0
// 		int4Key = 4
// 	)
// 	result := make([]*ffg.Element, int4Key)
// 	for i := int0Key; i < goldenposeidon.NUM_HASH_OUT_ELTS; i++ {
// 		result[i] = new(ffg.Element).Set(&t.TransfersHash.Elements[i])
// 	}

// 	return result
// }

func (t *Tx) Hash() *PoseidonHashOut {
	// var height uint8 = 7
	// initialLeaves := make([]*Transfer, 2)
	// zeroHash := new(Transfer)
	// tt, err := tree.NewTransferTree(height, initialLeaves, zeroHash)
	// if err != nil {
	// 	panic(err)
	// }
	// index := uint64(1)
	// leaf := new(Transfer).Set(&t.Transfers[0])
	// tt.AddLeaf(index, leaf)

	// input := t.ToFieldElementSlice()
	// return goldenposeidon.HashNoPad(input)
	h := new(PoseidonHashOut)
	h.SetZero()
	return h
}
