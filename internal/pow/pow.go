/*
see: https://raw.githubusercontent.com/Wollac/iota-crypto-demo/master/pkg/pow/v2/pow.go
*/
package pow

import (
	"encoding/binary"
	"errors"
	"math"
	"math/big"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/trinary"
	"golang.org/x/crypto/blake2b" // nolint:stylecheck

	// BLAKE2b_256 is the default hash function for the PoW digest
	_ "golang.org/x/crypto/blake2b" // nolint:stylecheck
)

const (
	nonceBytes     = 8  // len(uint64)
	tritsPerUint64 = 40 // largest x s.t. 3^x <= maxUint64
	maxHashKey     = "2367b879df2fe073dfc27f021fbc70343b4661546d9dcaa0de38db00a48d9b295613405167b19e1ddba02c679e2d7385b"
	uint64RadixKey = uint64(12157665459056928801)
	oneKey         = uint64(1)
	int16Key       = 16
)

type PoW interface {
	Score(msg []byte) (uint64, error)
	Difficulty(powDigest []byte, nonce uint64) (*big.Int, error)
	One() *big.Int
	MaxHash() *big.Int
	NonceBytes() int
	TargetScore() uint64
	EncodeNonce(dst trinary.Trits, nonce uint64)
	ToInt(trits trinary.Trits) (*big.Int, error)
}

type pow struct {
	targetScore uint64
	nonceBytes  int
	maxHash     *big.Int
	uint64Radix *big.Int
	one         *big.Int
}

func New(targetScore uint64) PoW {
	return &pow{
		targetScore: targetScore,
		nonceBytes:  nonceBytes,
		// largest possible integer representation of a Curl hash, i.e. 3^243
		maxHash: func() *big.Int {
			b, _ := new(big.Int).SetString(maxHashKey, int16Key)
			return b
		}(),
		// largest power of 3 fitting in an uint64, i.e. 3^tritsPerUint64 = 3^40
		uint64Radix: new(big.Int).SetUint64(uint64RadixKey),
		one:         new(big.Int).SetUint64(oneKey),
	}
}

// Score returns the PoW score of msg.
func (p *pow) Score(msg []byte) (uint64, error) {
	if len(msg) < p.nonceBytes {
		return 0, ErrMessageLengthInvalid
	}

	dataLen := len(msg) - p.nonceBytes
	// the PoW digest is the hash of msg without the nonce
	powDigest := blake2b.Sum256(msg[:dataLen])

	// extract the nonce from msg and compute the number of trailing zeros
	nonce := binary.LittleEndian.Uint64(msg[dataLen:])

	d, err := p.Difficulty(powDigest[:], nonce)
	if err != nil {
		return 0, errors.Join(ErrDifficultyFail, err)
	}

	// the score is the Difficulty per bytes, so we need to divide by the message length
	if d.IsUint64() {
		return d.Uint64() / uint64(len(msg)), nil
	}
	// try big.Int division
	d.Quo(d, big.NewInt(int64(len(msg))))
	if d.IsUint64() {
		return d.Uint64(), nil
	}
	// otherwise return the largest possible score
	return math.MaxUint64, nil
}

func (p *pow) Difficulty(powDigest []byte, nonce uint64) (*big.Int, error) {
	// allocate exactly one Curl block
	buf := make(trinary.Trits, consts.HashTrinarySize)
	n := b1t6.Encode(buf, powDigest)
	// add the nonce to the trit buffer
	p.EncodeNonce(buf[n:], nonce)

	c := curl.NewCurlP81()

	err := c.Absorb(buf)
	if err != nil {
		return nil, errors.Join(ErrAbsorbCurlP81Fail, err)
	}

	var digest trinary.Trits
	digest, err = c.Squeeze(consts.HashTrinarySize)
	if err != nil {
		return nil, errors.Join(ErrDigestFail, err)
	}

	var h *big.Int
	h, err = p.ToInt(digest)
	if err != nil {
		return nil, errors.Join(ErrConvertDigestFail, err)
	}

	return h.Quo(p.maxHash, h), nil
}

func (p *pow) One() *big.Int {
	return p.one
}

func (p *pow) MaxHash() *big.Int {
	return p.maxHash
}

func (p *pow) NonceBytes() int {
	return p.nonceBytes
}

func (p *pow) TargetScore() uint64 {
	return p.targetScore
}

// EncodeNonce encodes nonce as 48 trits using the b1t6 encoding.
func (p *pow) EncodeNonce(dst trinary.Trits, nonce uint64) {
	var nonceBuf [nonceBytes]byte
	binary.LittleEndian.PutUint64(nonceBuf[:], nonce)
	b1t6.Encode(dst, nonceBuf[:])
}

// ToInt converts the little-endian trinary hash into a positive integer.
// It returns t[242]*3^242 + ... + t[0]*3^0 + 1, where t[i] = { 2 if trits[i] = -1, trits[i] otherwise }.
func (p *pow) ToInt(trits trinary.Trits) (*big.Int, error) {
	if len(trits) != consts.HashTrinarySize {
		return nil, ErrPoWHashInvalid
	}

	const (
		int0Key = 0
		int1Key = 1
		int2Keu = 2
		int3Key = 3
		int9Key = 9

		n = consts.HashTrinarySize
	)

	b := new(big.Int).SetUint64(
		p.tritToUint(trits[n-int1Key])*9 +
			p.tritToUint(trits[n-2])*3 +
			p.tritToUint(trits[n-3]))

	// process as uint64 chunks to avoid costly bigint multiplication
	tmp := new(big.Int)
	for i := consts.HashTrinarySize/tritsPerUint64 - int1Key; i >= int0Key; i-- {
		chunk := trits[i*tritsPerUint64 : i*tritsPerUint64+tritsPerUint64]

		var v uint64
		for j := len(chunk) - int1Key; j >= int0Key; j-- {
			v = v*int3Key + p.tritToUint(chunk[j])
		}
		if i == int0Key {
			v++
		}
		_ = b.Add(b.Mul(b, p.uint64Radix), tmp.SetUint64(v))
	}
	return b, nil
}

func (p *pow) tritToUint(t int8) uint64 {
	const (
		int1Key = -1
		int2Key = 2
	)
	if t == int1Key {
		return int2Key
	}
	return uint64(t)
}
