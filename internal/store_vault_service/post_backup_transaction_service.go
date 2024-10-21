package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/use_cases/block_signature"
	postBackupTransaction "intmax2-node/internal/use_cases/post_backup_transaction"
)

func PostBackupTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *postBackupTransaction.UCPostBackupTransactionInput,
) (senderEnoughBalanceProofBodyHash string, err error) {
	_, err = db.CreateBackupTransaction(
		input.Sender, input.TxHash, input.EncryptedTx, input.Signature, int64(input.BlockNumber),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create backup transaction to db: %w", err)
	}
	senderEnoughBalanceProofBodyInput := block_signature.EnoughBalanceProofBodyInput{
		PrevBalanceProofBody:  input.SenderEnoughBalanceProofBody.PrevBalanceProofBody,
		TransferStepProofBody: input.SenderEnoughBalanceProofBody.TransferStepProofBody,
	}
	senderEnoughBalanceProofBody, err := senderEnoughBalanceProofBodyInput.EnoughBalanceProofBody()
	if err != nil {
		return "", fmt.Errorf("failed to get enough balance proof body: %w", err)
	}

	senderEnoughBalanceProofBodyHash = senderEnoughBalanceProofBody.Hash()
	log.Debugf("(PostBackupTransaction) senderEnoughBalanceProofBodyHash: %s", senderEnoughBalanceProofBodyHash)
	_, err = db.CreateBackupSenderProof(
		senderEnoughBalanceProofBody.PrevBalanceProofBody, senderEnoughBalanceProofBody.TransferStepProofBody, senderEnoughBalanceProofBodyHash,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create backup sender proof to db: %w", err)
	}

	return senderEnoughBalanceProofBodyHash, nil
}
