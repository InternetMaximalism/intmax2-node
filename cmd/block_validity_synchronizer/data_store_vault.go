package block_validity_synchronizer

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"time"
)

type DataStoreVault interface {
	// PostBackupTransfer is invoked `POST /backup/transfer` of the Data Store Vault.
	PostBackupTransfer(recipient *intMaxAcc.PublicKey, encryptedTransfer interface{}, blockNumber uint32) (recipientTransfer string, err error)

	// GetBackupTransfer is invoked `GET /backup/transfer` of the Data Store Vault.
	GetBackupTransfer(recipient *intMaxAcc.PublicKey, startBackupTime time.Time, limit uint) (responseData interface{}, err error)
}
