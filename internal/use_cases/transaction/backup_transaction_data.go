package transaction

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
)

func NewBackupTransactionData(
	userPublicKey *intMaxAcc.PublicKey,
	txDetails intMaxTypes.TxDetails,
	txHash *goldenposeidon.PoseidonHashOut,
	signature string,
) (*BackupTransactionData, error) {
	encodedTx := txDetails.Marshal()
	encryptedTx, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		userPublicKey,
		encodedTx,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt deposit: %w", err)
	}

	encodedEncryptedTx := base64.StdEncoding.EncodeToString(encryptedTx)

	return &BackupTransactionData{
		TxHash:             txHash.String(),
		EncodedEncryptedTx: encodedEncryptedTx,
		Signature:          signature,
	}, nil
}
