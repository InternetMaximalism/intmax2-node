package pow_test

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"intmax2-node/configs"
	"intmax2-node/internal/pow"
	"math"
	"math/big"
	mrd "math/rand"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorker_Mine(t *testing.T) {
	const int2Key = 2
	assert.NoError(t, configs.LoadDotEnv(int2Key))

	cfg := configs.New()
	p := pow.New(cfg.PoW.Difficulty)
	testWorker := pow.NewWorker(cfg.PoW.Workers, p)
	msg := append([]byte("Hello, World!"), make([]byte, p.NonceBytes())...)
	nonce, err := testWorker.Mine(context.Background(), msg[:len(msg)-p.NonceBytes()])
	require.NoError(t, err)

	// add nonce to message and check the resulting PoW score
	binary.LittleEndian.PutUint64(msg[len(msg)-p.NonceBytes():], nonce)
	score, err := p.Score(msg)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, score, cfg.PoW.Difficulty)
	t.Log(nonce, score)
}

func TestToInt(t *testing.T) {
	const int2Key = 2
	assert.NoError(t, configs.LoadDotEnv(int2Key))

	cfg := configs.New()
	p := pow.New(cfg.PoW.Difficulty)

	v, err := p.ToInt(make(trinary.Trits, consts.HashTrinarySize))
	assert.NoError(t, err)
	require.Zero(t, v.Cmp(p.One()))

	v, err = p.ToInt(largest(0))
	assert.NoError(t, err)
	require.Zero(t, v.Cmp(p.MaxHash()))
}

func TestSufficientTrailingZeros(t *testing.T) {
	const (
		dataLen = 9
		int2Key = 2
	)
	assert.NoError(t, configs.LoadDotEnv(int2Key))

	cfg := configs.New()
	for score := uint64(1); score <= math.MaxUint64/dataLen; score *= 3 {
		cfg.PoW.Difficulty = score
		p := pow.New(cfg.PoW.Difficulty)
		testWorker := pow.NewWorker(cfg.PoW.Workers, p)
		// the largest possible hash should be feasible
		s, err := testWorker.SufficientTrailingZeros(make([]byte, dataLen-p.NonceBytes()))
		assert.NoError(t, err)

		var vs *big.Int
		vs, err = p.ToInt(largest(s))
		assert.NoError(t, err)
		largestDifficulty := new(big.Int).Quo(p.MaxHash(), vs).Uint64()
		require.GreaterOrEqual(t, largestDifficulty, dataLen*cfg.PoW.Difficulty)

		// the smallest possible hash should be infeasible
		r := s - 1
		var vr *big.Int
		vr, err = p.ToInt(smallest(r))
		assert.NoError(t, err)
		smallestDifficulty := new(big.Int).Quo(p.MaxHash(), vr).Uint64()
		require.Less(t, smallestDifficulty, dataLen*cfg.PoW.Difficulty)
	}
}

func smallest(trailing int) trinary.Trits {
	trits := make(trinary.Trits, consts.HashTrinarySize)
	trits[consts.HashTrinarySize-(trailing+1)] = 1
	return trits
}

func largest(trailing int) trinary.Trits {
	trits := make(trinary.Trits, consts.HashTrinarySize)
	for i := 0; i < consts.HashTrinarySize-trailing; i++ {
		trits[i] = -1
	}
	return trits
}

func BenchmarkTritsToInt(b *testing.B) {
	src := make([]trinary.Trits, b.N)
	for i := range src {
		src[i] = randomTrits(consts.HashTrinarySize)
	}
	b.ResetTimer()

	const int2Key = 2
	assert.NoError(b, configs.LoadDotEnv(int2Key))

	cfg := configs.New()
	p := pow.New(cfg.PoW.Difficulty)
	for i := range src {
		_, err := p.ToInt(src[i])
		assert.NoError(b, err)
	}
}

const benchBytesLen = 1600

func BenchmarkScore(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = make([]byte, benchBytesLen)
		if _, err := rand.Read(data[i]); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()

	const int2Key = 2
	assert.NoError(b, configs.LoadDotEnv(int2Key))

	cfg := configs.New()
	p := pow.New(cfg.PoW.Difficulty)
	for i := range data {
		_, err := p.Score(data[i])
		assert.NoError(b, err)
	}
}

func randomTrits(n int) trinary.Trits {
	trytes := randomTrytes((n + 2) / 3)
	return trinary.MustTrytesToTrits(trytes)[:n]
}

func randomTrytes(n int) trinary.Trytes {
	var result strings.Builder
	result.Grow(n)
	for i := 0; i < n; i++ {
		result.WriteByte(consts.TryteAlphabet[mrd.Intn(len(consts.TryteAlphabet))])
	}
	return result.String()
}
