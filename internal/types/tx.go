package types

import (
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type Tx struct {
	FeeTransferHash  *poseidonHashOut
	TransferTreeRoot *poseidonHashOut
}

func (t *Tx) Set(tx *Tx) *Tx {
	t.FeeTransferHash = tx.FeeTransferHash
	t.TransferTreeRoot = tx.TransferTreeRoot

	return t
}

// Testing purposes only
func (t *Tx) SetRandom() (*Tx, error) {
	var err error
	t.FeeTransferHash, err = new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}
	t.TransferTreeRoot, err = new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Tx) ToFieldElementSlice() []*ffg.Element {
	result := make([]*ffg.Element, 8)
	for i := 0; i < goldenposeidon.NUM_HASH_OUT_ELTS; i++ {
		result[i] = new(ffg.Element).Set(&t.FeeTransferHash.Elements[i])
	}
	for i := 0; i < goldenposeidon.NUM_HASH_OUT_ELTS; i++ {
		result[i+goldenposeidon.NUM_HASH_OUT_ELTS] = new(ffg.Element).Set(&t.TransferTreeRoot.Elements[i])
	}

	return result
}

func (t *Tx) Hash() *poseidonHashOut {
	input := t.ToFieldElementSlice()
	return goldenposeidon.HashNoPad(input)
}
