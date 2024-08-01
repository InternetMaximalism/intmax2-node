package store_vault_service

import (
	"crypto/rand"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
)

func EncryptTransfer(publicKey intMaxAcc.PublicKey, transfer *intMaxTypes.Transfer) ([]byte, error) {
	encodedTransfer := transfer.Marshal()
	return intMaxAcc.EncryptECIES(rand.Reader, &publicKey, encodedTransfer)
}

func DecryptTransfer(privateKey intMaxAcc.PrivateKey, encryptedTransfer []byte) ([]byte, error) {
	encodedTransfer, err := privateKey.DecryptECIES(encryptedTransfer)
	if err != nil {
		return nil, err
	}

	return encodedTransfer, nil
}
