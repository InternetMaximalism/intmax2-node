package transaction

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"time"

	"github.com/holiman/uint256"
	"github.com/iden3/go-iden3-crypto/ffg"
)

func MakeMessage(
	transfersHash []byte,
	nonce uint64,
	powNonce string,
	sender intMaxAcc.Address,
	expiration time.Time,
) ([]ffg.Element, error) {
	const (
		int1Key         = 1
		int32Key        = 32
		numMessageBytes = int32Key + int1Key + int32Key + int32Key + int1Key
	)

	message := finite_field.NewBuffer(make([]ffg.Element, numMessageBytes))
	finite_field.WriteFixedSizeBytes(message, transfersHash, int32Key)

	err := finite_field.WriteUint64(message, nonce)
	if err != nil {
		return nil, err
	}

	var pwN uint256.Int
	err = pwN.SetFromHex(powNonce)
	if err != nil {
		return nil, err
	}
	finite_field.WriteFixedSizeBytes(message, pwN.Bytes(), int32Key)
	finite_field.WriteFixedSizeBytes(message, sender.Bytes(), int32Key)

	expirationInt := expiration.Unix()
	if expirationInt < 0 {
		return nil, ErrValueInvalid
	}
	err = finite_field.WriteUint64(message, uint64(expirationInt))
	if err != nil {
		return nil, err
	}

	return message.Inner(), nil
}
