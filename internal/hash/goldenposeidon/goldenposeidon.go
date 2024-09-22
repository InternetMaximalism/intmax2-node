package goldenposeidon

/// The implementation is the same as in the link below, but with the addition of the Permute function.
/// https://github.com/iden3/go-iden3-crypto/blob/e5cf066b8be3da9a3df9544c65818df189fdbebe/goldenposeidon/poseidon.go
///
/// The following implementations were also referred to.
/// - utility functions related to Poseidon hash
///   https://github.com/0xPolygonZero/plonky2/blob/b600142cd454b95eba403fa1f86f582ff8688c79/plonky2/src/hash/hashing.rs

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
	poseidon "github.com/iden3/go-iden3-crypto/goldenposeidon"
)

const (
	NROUNDSF          = poseidon.NROUNDSF
	NROUNDSP          = poseidon.NROUNDSP
	CAPLEN            = poseidon.CAPLEN
	mLen              = 12 // poseidon.mLen
	NUM_HASH_OUT_ELTS = 4
	ALPHA             = 7

	int4Key  = 4
	int8Key  = 8
	int32Key = 32
)

var (
	S = poseidon.S
	C = poseidon.C
	P = poseidon.P
	M = poseidon.M
)

func zero() *ffg.Element {
	return ffg.NewElement()
}

var big7 = big.NewInt(ALPHA)

// exp7 performs x^7 mod p
func exp7(a *ffg.Element) {
	a.Exp(*a, big7)
}

// exp7state perform exp7 for whole state
func exp7state(state []*ffg.Element) {
	for i := 0; i < len(state); i++ {
		exp7(state[i])
	}
}

// ark computes Add-Round Key, from the paper https://eprint.iacr.org/2019/458.pdf
func ark(state []*ffg.Element, it int) {
	for i := 0; i < len(state); i++ {
		state[i].Add(state[i], C[it+i])
	}
}

// mix returns [[matrix]] * [vector]
func mix(state []*ffg.Element, opt bool) []*ffg.Element {
	mul := zero()
	newState := make([]*ffg.Element, mLen)
	for i := 0; i < mLen; i++ {
		newState[i] = zero()
	}
	for i := 0; i < mLen; i++ {
		newState[i].SetUint64(0)
		for j := 0; j < mLen; j++ {
			if opt {
				mul.Mul(P[j][i], state[j])
			} else {
				mul.Mul(M[j][i], state[j])
			}
			newState[i].Add(newState[i], mul)
		}
	}
	return newState
}

// Hash computes the hash for the given inputs
// nolint:gocritic
func Permute(input [mLen]*ffg.Element) [mLen]*ffg.Element {
	state := make([]*ffg.Element, mLen)
	for i := 0; i < mLen; i++ {
		state[i] = input[i]
	}
	for i := 0; i < mLen; i++ {
		state[i].Add(state[i], C[i])
	}

	for r := 0; r < NROUNDSF/2; r++ {
		exp7state(state)
		ark(state, (r+1)*mLen)
		state = mix(state, r == NROUNDSF/2-1)
	}

	for r := 0; r < NROUNDSP; r++ {
		exp7(state[0])
		state[0].Add(state[0], C[(NROUNDSF/2+1)*mLen+r])

		s0 := zero()
		mul := zero()
		mul.Mul(S[(mLen*2-1)*r], state[0])
		s0.Add(s0, mul)
		for i := 1; i < mLen; i++ {
			mul.Mul(S[(mLen*2-1)*r+i], state[i])
			s0.Add(s0, mul)
			mul.Mul(S[(mLen*2-1)*r+mLen+i-1], state[0])
			state[i].Add(state[i], mul)
		}
		state[0] = s0
	}

	for r := 0; r < NROUNDSF/2; r++ {
		exp7state(state)
		if r < NROUNDSF/2-1 {
			ark(state, (NROUNDSF/2+1+r)*mLen+NROUNDSP)
		}

		state = mix(state, false)
	}

	result := [mLen]*ffg.Element{}
	for i := 0; i < mLen; i++ {
		result[i] = state[i]
	}

	return result
}

func Hash(inpBI [NROUNDSF]uint64, capBI [CAPLEN]uint64) [CAPLEN]uint64 {
	input := [mLen]*ffg.Element{
		ffg.NewElement().SetUint64(inpBI[0]),
		ffg.NewElement().SetUint64(inpBI[1]),
		ffg.NewElement().SetUint64(inpBI[2]),
		ffg.NewElement().SetUint64(inpBI[3]),
		ffg.NewElement().SetUint64(inpBI[4]),
		ffg.NewElement().SetUint64(inpBI[5]),
		ffg.NewElement().SetUint64(inpBI[6]),
		ffg.NewElement().SetUint64(inpBI[7]),
		ffg.NewElement().SetUint64(capBI[0]),
		ffg.NewElement().SetUint64(capBI[1]),
		ffg.NewElement().SetUint64(capBI[2]),
		ffg.NewElement().SetUint64(capBI[3]),
	}

	output := Permute(input)

	return [CAPLEN]uint64{
		output[0].ToUint64Regular(),
		output[1].ToUint64Regular(),
		output[2].ToUint64Regular(),
		output[3].ToUint64Regular(),
	}
}

// Hash a message without any padding step. Note that this can enable length-extension attacks.
// However, it is still collision-resistant in cases where the input has a fixed length.
func hashNToMNoPad(
	inputs []ffg.Element,
	numOutputs int,
) []ffg.Element {
	if numOutputs <= 0 {
		panic("numOutputs must be greater than 0")
	}

	perm := [mLen]*ffg.Element{}
	for i := 0; i < mLen; i++ {
		perm[i] = zero()
	}

	// Absorb all input chunks.
	for i := 0; i < len(inputs); i += NROUNDSF {
		for j := 0; j < NROUNDSF; j++ {
			if i+j < len(inputs) {
				perm[j] = &inputs[i+j]
			}
		}
		perm = Permute(perm)
	}

	// Squeeze until we have the desired number of outputs.
	outputs := []ffg.Element{}
	for {
		for _, item := range perm[0:NROUNDSF] {
			outputs = append(outputs, *item)
			if len(outputs) == numOutputs {
				return outputs
			}
		}
		perm = Permute(perm)
	}
}

func HashNoPad(
	inputs []ffg.Element,
) *PoseidonHashOut {
	outputs := hashNToMNoPad(inputs, NUM_HASH_OUT_ELTS)
	result := NewPoseidonHashOut()
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		result.Elements[i].Set(&outputs[i])
	}

	return result
}

type PoseidonHashOut struct {
	Elements [NUM_HASH_OUT_ELTS]ffg.Element
}

func NewPoseidonHashOut() *PoseidonHashOut {
	h := new(PoseidonHashOut)
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		h.Elements[i] = *new(ffg.Element).SetZero()
	}

	return h
}

func (dst *PoseidonHashOut) Set(src *PoseidonHashOut) *PoseidonHashOut {
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		dst.Elements[i] = src.Elements[i]
	}

	return dst
}

func (h *PoseidonHashOut) SetRandom() (*PoseidonHashOut, error) {
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		_, err := h.Elements[i].SetRandom()
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (h *PoseidonHashOut) SetZero() *PoseidonHashOut {
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		h.Elements[i].SetZero()
	}

	return h
}

func (h *PoseidonHashOut) Equal(other *PoseidonHashOut) bool {
	if h == nil || other == nil {
		return false
	}

	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		if !h.Elements[i].Equal(&other.Elements[i]) {
			return false
		}
	}

	return true
}

func (h *PoseidonHashOut) Marshal() []byte {
	if h == nil {
		panic("value is nil")
	}

	result := []byte{}
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		b := h.Elements[i].Marshal() // big-endian
		result = append(result, b...)
	}

	return result
}

func (h *PoseidonHashOut) String() string {
	return "0x" + hex.EncodeToString(h.Marshal())
}

func (h *PoseidonHashOut) Unmarshal(data []byte) error {
	const elementSize = 8
	if len(data) != NUM_HASH_OUT_ELTS*elementSize {
		fmt.Printf("Fail to unmarshal data: %v\n", data)
		return fmt.Errorf("invalid data size: %d", len(data))
	}

	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		r := data[i*elementSize : (i+1)*elementSize] // big-endian
		h.Elements[i] = *new(ffg.Element).SetBytes(r)
	}

	return nil
}

func (h *PoseidonHashOut) FromString(s string) error {
	if has0xPrefix(s) {
		s = s[2:]
	}

	data, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	return h.Unmarshal(data)
}

func (h *PoseidonHashOut) FromPartial(elementsIn []ffg.Element) *PoseidonHashOut {
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		h.Elements[i] = elementsIn[i]
	}

	return h
}

func (h PoseidonHashOut) MarshalJSON() ([]byte, error) {
	hashOutHex := "0x" + hex.EncodeToString(h.Marshal())
	return json.Marshal(hashOutHex)
}

func (h *PoseidonHashOut) UnmarshalJSON(data []byte) error {
	var hexStr string
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return err
	}
	if !has0xPrefix(hexStr) {
		return fmt.Errorf("invalid hex string: %s", hexStr)
	}
	hashOutHex, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return err
	}

	return h.Unmarshal(hashOutHex)
}

func (h *PoseidonHashOut) Uint32Array() [int8Key]uint32 {
	flatten := []uint32{}
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		e := h.Elements[i][0]
		low := uint32(e)
		high := uint32(e >> int32Key)
		flatten = append(flatten, high, low)
	}

	limbs := [int8Key]uint32{}
	for i := 0; i < len(limbs); i++ {
		limbs[i] = flatten[i]
	}

	return limbs
}

func (h *PoseidonHashOut) Uint32Slice() []uint32 {
	b := h.Uint32Array()
	return b[:]
}

func HexToHash(s string) *PoseidonHashOut {
	if has0xPrefix(s) {
		s = s[2:]
	}

	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	result := new(PoseidonHashOut)
	err = result.Unmarshal(data)
	if err != nil {
		panic(err)
	}

	return result
}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func Compress(input1, input2 *PoseidonHashOut) *PoseidonHashOut {
	if input1 == nil || input2 == nil {
		return nil
	}

	input := [mLen]*ffg.Element{}
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		input[i] = new(ffg.Element).Set(&input1.Elements[i])
	}
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		input[i+NUM_HASH_OUT_ELTS] = new(ffg.Element).Set(&input2.Elements[i])
	}
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		input[i+NUM_HASH_OUT_ELTS*2] = zero()
	}
	output := Permute(input)
	result := new(PoseidonHashOut)
	for i := 0; i < NUM_HASH_OUT_ELTS; i++ {
		result.Elements[i] = *output[i]
	}

	return result
}
