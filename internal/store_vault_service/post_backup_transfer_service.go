package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupTransfer "intmax2-node/internal/use_cases/post_backup_transfer"
)

func PostBackupTransfer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupTransfer.UCPostBackupTransferInput,
) (err error) {
	// var senderLastBalanceProofBody []byte
	// senderLastBalanceProofBody, err = base64.StdEncoding.DecodeString(input.SenderLastBalanceProofBody)
	// if err != nil {
	// 	return fmt.Errorf("failed to decode sender balance proof body: %w", err)
	// }

	// var senderBalanceTransitionProofBody []byte
	// senderBalanceTransitionProofBody, err = base64.StdEncoding.DecodeString(input.SenderTransitionProofBody)
	// if err != nil {
	// 	return fmt.Errorf("failed to decode sender balance proof body: %w", err)
	// }

	_, err = db.CreateBackupTransfer(
		input.Recipient, input.TransferHash, input.EncryptedTransfer,
		int64(input.BlockNumber),
	)
	if err != nil {
		return fmt.Errorf("failed to create backup transfer to db: %w", err)
	}
	return nil
}
