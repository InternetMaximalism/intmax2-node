package types

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type ValidityPublicInputs struct {
	PublicState    *PublicState
	TxTreeRoot     intMaxTypes.Bytes32
	SenderTreeRoot *intMaxGP.PoseidonHashOut
	IsValidBlock   bool
}

func (vpi *ValidityPublicInputs) Genesis() *ValidityPublicInputs {
	txTreeRoot := intMaxTypes.Bytes32{}
	senderTreeRoot := new(intMaxGP.PoseidonHashOut).SetZero()
	isValidBlock := false

	return &ValidityPublicInputs{
		PublicState:    new(PublicState).Genesis(),
		TxTreeRoot:     txTreeRoot,
		SenderTreeRoot: senderTreeRoot,
		IsValidBlock:   isValidBlock,
	}
}

func (vpi *ValidityPublicInputs) FromPublicInputs(publicInputs []ffg.Element) *ValidityPublicInputs {
	const (
		txTreeRootOffset     = PublicStateLimbSize
		senderTreeRootOffset = txTreeRootOffset + Int8Key
		isValidBlockOffset   = senderTreeRootOffset + intMaxGP.NUM_HASH_OUT_ELTS
		end                  = isValidBlockOffset + 1
	)

	vpi.PublicState = new(PublicState).FromFieldElementSlice(publicInputs[:txTreeRootOffset])
	txTreeRoot := intMaxTypes.Bytes32{}
	copy(txTreeRoot[:], FieldElementSliceToUint32Slice(publicInputs[txTreeRootOffset:senderTreeRootOffset]))
	vpi.TxTreeRoot = txTreeRoot
	vpi.SenderTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(publicInputs[senderTreeRootOffset:isValidBlockOffset])
	vpi.IsValidBlock = publicInputs[isValidBlockOffset].ToUint64Regular() == 1

	return vpi
}

func (vpi *ValidityPublicInputs) Equal(other *ValidityPublicInputs) bool {
	if !vpi.PublicState.Equal(other.PublicState) {
		return false
	}
	if !vpi.TxTreeRoot.Equal(&other.TxTreeRoot) {
		return false
	}
	if !vpi.SenderTreeRoot.Equal(other.SenderTreeRoot) {
		return false
	}
	if vpi.IsValidBlock != other.IsValidBlock {
		return false
	}

	return true
}
