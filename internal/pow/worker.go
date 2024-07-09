package pow

import (
	"context"
	"errors"
	"math"
	"math/big"
	"math/bits"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl/bct"
	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/trinary"
	"golang.org/x/crypto/blake2b"
)

// errors returned by the PoW
var (
	ErrCancelled = errors.New("canceled")
	ErrDone      = errors.New("done")
)

type Worker interface {
	Mine(ctx context.Context, data []byte) (uint64, error)
	SufficientTrailingZeros(data []byte) (int, error)
}

// The worker performs the PoW.
type worker struct {
	numWorkers int
	pow        PoW
}

// NewWorker creates a new PoW worker.
// The optional numWorkers specifies how many go routines should be used to perform the PoW.
func NewWorker(numWorkers int, pow PoW) Worker {
	const (
		int0Key = 0
		int1Key = 1
	)

	if numWorkers > int0Key {
		numWorkers = int1Key
	}

	return &worker{
		numWorkers: numWorkers,
		pow:        pow,
	}
}

// Mine performs the PoW for data.
// It returns a nonce that appended to data results in a PoW score of at least targetScore.
// The computation can be canceled anytime using ctx.
func (w *worker) Mine(ctx context.Context, data []byte) (uint64, error) {
	const (
		int0Key = 0
		int1Key = 1
	)

	// for the zero target score, the solution is trivial
	if w.pow.TargetScore() == int0Key {
		return int0Key, nil
	}

	var (
		done      uint32
		counter   uint64
		wg        sync.WaitGroup
		results   = make(chan uint64, w.numWorkers)
		closing   = make(chan struct{})
		errorChan = make(chan error)
	)

	// compute the digest
	powDigest := blake2b.Sum256(data)

	// stop when the context has been canceled
	go func() {
		select {
		case <-ctx.Done():
			atomic.StoreUint32(&done, int1Key)
		case <-closing:
			return
		}
	}()

	sufficientTrailing, err := w.SufficientTrailingZeros(data)
	if err != nil {
		return int0Key, errors.Join(ErrSufficientTrailingZerosFail, err)
	}

	target := w.targetHash(data)

	workerWidth := math.MaxUint64 / uint64(w.numWorkers)
	for i := int0Key; i < w.numWorkers; i++ {
		startNonce := uint64(i) * workerWidth
		wg.Add(int1Key)
		go func() {
			defer wg.Done()

			nonce, workerErr := w.worker(powDigest[:], startNonce, sufficientTrailing, target, &done, &counter)
			if workerErr != nil && !errors.Is(workerErr, ErrDone) {
				errorChan <- workerErr
				return
			}
			atomic.StoreUint32(&done, int1Key)
			results <- nonce
		}()
	}
	wg.Wait()
	close(results)
	close(closing)

	for {
		select {
		case nonce, ok := <-results:
			if !ok {
				return int0Key, ErrCancelled
			}
			return nonce, nil
		case errCh := <-errorChan:
			return int1Key, errors.Join(ErrCancelled, errCh)
		}
	}
}

// SufficientTrailingZeros returns 𝑠 s.t. any hash with 𝑠 trailing zeroes is feasible, i.e. smallest 𝑠 with 3^𝑠 ≥ 𝑙·𝑥.
// It panics when 𝑙·𝑥 overflows an uint64.
// It is sufficient to show that the largest (worst) hash h with 𝑠 trailing zeroes is feasible i.e. ⌊ maxHash / h ⌋ ≥ 𝑙·𝑥
// ⌊ maxHash / h ⌋ ≥ ⌊ 3^243 / 3^(243 - 𝑠) ⌋ = ⌊ 3^𝑠 ⌋ = 3^𝑠 ≥ 𝑙·𝑥
func (w *worker) SufficientTrailingZeros(data []byte) (int, error) {
	const (
		int0Key = 0
		int1Key = 1
		int3Key = 3
	)

	// assure that (len(data)+nonceBytes) * targetScore <= MaxUint64
	if (math.MaxUint64-int1Key)/(uint64(len(data)+w.pow.NonceBytes()))+int1Key < w.pow.TargetScore() {
		return int0Key, ErrScoreTargetInvalidFail
	}

	lx := uint64(len(data)+w.pow.NonceBytes()) * w.pow.TargetScore()

	// in order to prevent floating point rounding errors, compute the exact integer logarithm
	for s, v := int0Key, uint64(int1Key); s <= tritsPerUint64; s++ {
		if v >= lx {
			return s, nil
		}
		v *= int3Key
	}
	return tritsPerUint64 + int1Key, nil
}

// targetHash returns 𝑡 s.t. any hash with h ≤ 𝑡 is feasible, i.e. h = ⌊ maxHash / (𝑙·𝑥 + 1) ⌋.
// It panics when 𝑙·𝑥 overflows an uint64.
// It is sufficient to show that for the hash h with h = ⌊ maxHash / (𝑙·𝑥 + 1) ⌋, ⌊ maxHash / h ⌋ ≥ 𝑙·𝑥 always holds:
// h = ⌊ maxHash / (𝑙·𝑥 + 1) ⌋ ≤ maxHash / (𝑙·𝑥 + 1) ⇔ 𝑙·𝑥 + 1 ≤ maxHash / h ⇔ 𝑙·𝑥 ≤ maxHash / h - 1 ⇔ 𝑙·𝑥 < ⌊ maxHash / h ⌋
func (w *worker) targetHash(data []byte) *big.Int {
	// return ⌊ maxHash / (𝑙·𝑥 + 1) ⌋
	z := new(big.Int).SetUint64(w.pow.TargetScore())
	z.Mul(z, big.NewInt(int64(len(data)+w.pow.NonceBytes())))
	z.Add(z, w.pow.One())
	return z.Quo(w.pow.MaxHash(), z)
}

func (w *worker) worker(
	powDigest []byte,
	startNonce uint64,
	sufficientTrailing int,
	target *big.Int,
	done *uint32,
	counter *uint64,
) (uint64, error) {
	if sufficientTrailing > consts.HashTrinarySize {
		return 0, ErrPoWTrailingZerosTargetInvalid
	}

	// use batched Curl hashing
	c := bct.NewCurlP81()
	var l, h [consts.HashTrinarySize]uint

	// allocate exactly one Curl block for each batch index and fill it with the encoded digest
	buf := make([]trinary.Trits, bct.MaxBatchSize)
	for i := range buf {
		buf[i] = make(trinary.Trits, consts.HashTrinarySize)
		b1t6.Encode(buf[i], powDigest)
	}

	digestTritsLen := b1t6.EncodedLen(len(powDigest))
	for nonce := startNonce; atomic.LoadUint32(done) == 0; nonce += bct.MaxBatchSize {
		// add the nonce to each trit buffer
		for i := range buf {
			nonceBuf := buf[i][digestTritsLen:]
			w.pow.EncodeNonce(nonceBuf, nonce+uint64(i))
		}

		// process the batch
		c.Reset()
		if err := c.Absorb(buf, consts.HashTrinarySize); err != nil {
			return 0, errors.Join(ErrAbsorbCurlP81Fail, err)
		}
		c.CopyState(l[:], h[:]) // the first 243 entries of the state correspond to the resulting hashes
		atomic.AddUint64(counter, bct.MaxBatchSize)

		// check the state whether it corresponds to a hash with sufficient amount of trailing zeros
		// this is equivalent to computing the hashes with Squeeze and then checking TrailingZeros of each
		i, err := w.checkStateTrits(&l, &h, sufficientTrailing, target)
		if err != nil {
			return 0, errors.Join(ErrCheckStateTritsFail, err)
		}

		if i < bct.MaxBatchSize {
			// if i := checkStateTrits2(&l, &h, target); i < bct.MaxBatchSize {
			return nonce + uint64(i), nil
		}
	}

	return 0, ErrDone
}

func (w *worker) checkStateTrits(
	l, h *[consts.HashTrinarySize]uint,
	sufficientTrailing int,
	target *big.Int,
) (int, error) {
	const (
		int0Key = 0
	)

	var v uint

	requiredTrailing := sufficientTrailing - 1
	for i := consts.HashTrinarySize - requiredTrailing; i < consts.HashTrinarySize; i++ {
		v |= l[i] ^ h[i] // 0 if trit is zero, 1 otherwise
	}
	// no hash has at least sufficientTrailing number of trailing zeroes
	if v == ^uint(int0Key) {
		// there cannot be a valid hash
		return bct.MaxBatchSize, nil
	}

	// find hashes with at least sufficientTrailing+1 number of trailing zeroes
	lw := v | (l[consts.HashTrinarySize-sufficientTrailing] ^ h[consts.HashTrinarySize-sufficientTrailing])
	// if there is one this is sufficient, and we can return the index
	if lw != ^uint(int0Key) {
		// return the index of the first zero bit, this corresponds to the hash with sufficient trailing zeros
		return bits.TrailingZeros(^lw), nil
	}

	// otherwise, we have to convert all hashes with at least sufficientTrailing number of trailing zeroes and check
	lo, hi := bits.TrailingZeros(^v), bits.Len(^v)
	for i := lo; i < hi; i++ {
		sti, err := w.stateToInt(l, h, uint(i))
		if err != nil {
			return 0, errors.Join(ErrStateToIntFail, err)
		}

		if (v>>i)&1 == int0Key && sti.Cmp(target) <= int0Key {
			return i, nil
		}
	}

	return bct.MaxBatchSize, nil
}

func (w *worker) stateToInt(l, h *[consts.HashTrinarySize]uint, idx uint) (i *big.Int, err error) {
	idx &= bits.UintSize - 1 // hint to the compiler that shifts don't need guard code

	var trits [consts.HashTrinarySize]int8
	for j := consts.HashTrinarySize - 1; j >= 0; j-- {
		trits[j] = int8((h[j]>>idx)&1) - int8((l[j]>>idx)&1)
	}

	i, err = w.pow.ToInt(trits[:])
	if err != nil {
		return nil, errors.Join(ErrConvertToIntFail, err)
	}

	return i, nil
}
