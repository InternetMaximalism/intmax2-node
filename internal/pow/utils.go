package pow

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/holiman/uint256"
)

type powNonce struct {
	pow    PoW
	worker Worker
}

func NewPoWNonce(pow PoW, worker Worker) PoWNonce {
	return &powNonce{
		pow:    pow,
		worker: worker,
	}
}

func (p *powNonce) Nonce(ctx context.Context, msg []byte) (string, error) {
	const defNonceKey = ""

	msg = append(msg, make([]byte, p.pow.NonceBytes())...)
	nonce, err := p.worker.Mine(ctx, msg[:len(msg)-p.pow.NonceBytes()])
	if err != nil {
		return defNonceKey, errors.Join(ErrMinePoWNonceFail, err)
	}

	var n uint256.Int
	_ = n.SetUint64(nonce)

	return n.Hex(), nil
}

func (p *powNonce) Verify(nonce string, msg []byte) (err error) {
	var pwNonce uint256.Int
	err = pwNonce.SetFromHex(nonce)
	if err != nil {
		return errors.Join(ErrPoWNonceInvalid, err)
	}

	msg = append(msg, make([]byte, p.pow.NonceBytes())...)
	binary.LittleEndian.PutUint64(msg[len(msg)-p.pow.NonceBytes():], pwNonce.Uint64())

	score, err := p.pow.Score(msg)
	if err != nil {
		return errors.Join(ErrScoreFail, err)
	}

	if score < p.pow.TargetScore() {
		return ErrPoWNonceInvalid
	}

	return nil
}
