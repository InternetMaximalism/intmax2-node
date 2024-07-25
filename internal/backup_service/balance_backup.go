package backup_service

import (
	"encoding/json"

	"intmax2-node/internal/sql_db/pgx/models"
	"intmax2-node/internal/use_cases/backup_balance"
)

func BackupUserBalance(
	db SQLDriverApp,
	input *backup_balance.UCPostBackupBalanceInput,
) error {
	var (
		eTx, eTns, eDep []byte
		err             error
	)

	eTx, err = json.Marshal(input.EncryptedTxs)
	if err != nil {
		return err
	}
	eTns, err = json.Marshal(input.EncryptedTransfers)
	if err != nil {
		return err
	}
	eDep, err = json.Marshal(input.EncryptedDeposits)
	if err != nil {
		return err
	}

	return db.BackupUserBalance(&models.BalanceBackup{
		UserAddress:           input.DecodeUser.ToAddress().String(),
		BlockNumber:           input.BlockNumber,
		EncryptedBalanceProof: input.EncryptedBalanceProof.Proof,
		EncryptedPublicInputs: input.EncryptedBalanceProof.EncryptedPublicInputs,
		EncryptedTxs:          eTx,
		EncryptedTransfers:    eTns,
		EncryptedDeposits:     eDep,
		Signature:             input.Signature,
	})
}
