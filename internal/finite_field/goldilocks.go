package finite_field

import (
	"encoding/binary"

	"github.com/iden3/go-iden3-crypto/ffg"
)

// BytesToFieldElementSlice converts a hash to a slice of ffg.Elements, ensuring the hash is padded
// to a multiple of 4 bytes.
func BytesToFieldElementSlice(bytes []byte) []*ffg.Element {
	const uint32ByteSize = 4
	hashByteSize := len(bytes)
	numLimbs := (hashByteSize + uint32ByteSize - 1) / uint32ByteSize // rounds up the division
	for len(bytes) != numLimbs*uint32ByteSize {
		bytes = append(bytes, 0)
	}
	flattenTxTreeRoot := make([]*ffg.Element, numLimbs)
	for i := 0; i < len(flattenTxTreeRoot); i++ {
		v := binary.BigEndian.Uint32(bytes[uint32ByteSize*i : uint32ByteSize*(i+1)])
		flattenTxTreeRoot[i] = new(ffg.Element).SetUint64(uint64(v))
	}

	return flattenTxTreeRoot
}
