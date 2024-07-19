package finite_field

import (
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type Buffer struct {
	buf []ffg.Element
	off int
}

func NewBuffer(buf []ffg.Element) *Buffer {
	return &Buffer{
		buf: buf,
	}
}

func (b *Buffer) Inner() []ffg.Element { return b.buf[0:b.off] }

func Write(buf *Buffer, data interface{}) error {
	switch d := data.(type) {
	case ffg.Element:
		WriteGoldilocksField(buf, &d)
	case *ffg.Element:
		WriteGoldilocksField(buf, d)
	case goldenposeidon.PoseidonHashOut:
		WritePoseidonHashOut(buf, &d)
	case *goldenposeidon.PoseidonHashOut:
		WritePoseidonHashOut(buf, d)
	case uint64:
		return WriteUint64(buf, d)
	case []byte:
		WriteBytes(buf, d)
	default:
		return ErrUnknownType
	}

	return nil
}

func WriteGoldilocksField(buf *Buffer, data *ffg.Element) {
	buf.buf[buf.off].Set(data)
	buf.off++
}

func WritePoseidonHashOut(buf *Buffer, data *goldenposeidon.PoseidonHashOut) {
	for i := 0; i < len(data.Elements); i++ {
		WriteGoldilocksField(buf, &data.Elements[i])
	}
}

func WriteUint64(buf *Buffer, data uint64) error {
	if data >= ffg.Modulus().Uint64() {
		return ErrValueTooLarge
	}
	d := new(ffg.Element).SetUint64(data)
	WriteGoldilocksField(buf, d)

	return nil
}

func WriteFixedSizeBytes(buf *Buffer, data []byte, numDataBytes int) {
	const int4Key = 4
	for len(data) < numDataBytes {
		data = append(data, 0)
	}
	for i := 0; i < numDataBytes; i += int4Key {
		m := min(numDataBytes, i+int4Key)
		d := new(ffg.Element).SetBytes(data[i:m])
		WriteGoldilocksField(buf, d)
	}
}

func WriteBytes(buf *Buffer, data []byte) {
	WriteUint64(buf, uint64(len(data)))
	WriteFixedSizeBytes(buf, data, len(data))
}

// BytesToFieldElementSlice converts a hash to a slice of ffg.Elements, ensuring the hash is padded
// to a multiple of 4 bytes.
func BytesToFieldElementSlice(bytes []byte) []ffg.Element {
	const uint32ByteSize = 4
	hashByteSize := len(bytes)
	numLimbs := (hashByteSize + uint32ByteSize - 1) / uint32ByteSize // rounds up the division
	for len(bytes) != numLimbs*uint32ByteSize {
		bytes = append(bytes, 0)
	}

	buf := NewBuffer(make([]ffg.Element, numLimbs))
	WriteFixedSizeBytes(buf, bytes, numLimbs*uint32ByteSize)

	return buf.Inner()
}
