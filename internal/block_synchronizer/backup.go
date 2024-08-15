package block_synchronizer

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	postBackupTransaction "intmax2-node/internal/use_cases/post_backup_transaction"
	"intmax2-node/internal/use_cases/post_backup_transfer"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-resty/resty/v2"
)

func (d *blockSynchronizer) BackupTransaction(
	sender intMaxAcc.Address,
	txHash, encodedEncryptedTx string,
	signature string,
	blockNumber uint64,
) error {
	err := backupTransactionRawRequest(
		d.ctx,
		d.cfg,
		d.log,
		txHash,
		encodedEncryptedTx,
		signature,
		sender.String(),
		uint32(blockNumber),
	)

	if err != nil {
		return fmt.Errorf("failed to backup transaction: %w", err)
	}

	return nil
}

func backupTransactionRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	txHash, encodedEncryptedTx string,
	signature string,
	sender string,
	blockNumber uint32,
) error {
	ucInput := postBackupTransaction.UCPostBackupTransactionInput{
		TxHash:      txHash,
		EncryptedTx: encodedEncryptedTx,
		Sender:      sender,
		BlockNumber: blockNumber,
		Signature:   signature,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/transaction", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return err
	}

	response := new(node.BackupTransactionResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}

func (d *blockSynchronizer) BackupTransfer(
	recipient intMaxAcc.Address,
	encodedEncryptedTransferHash, encodedEncryptedTransfer string,
	blockNumber uint64,
) error {
	err := backupTransferRawRequest(
		d.ctx,
		d.cfg,
		d.log,
		encodedEncryptedTransferHash,
		encodedEncryptedTransfer,
		recipient.String(),
		uint32(blockNumber),
	)

	if err != nil {
		return fmt.Errorf("failed to backup transfer: %w", err)
	}

	return nil
}

func backupTransferRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	encodedEncryptedTransferHash, encodedEncryptedTransfer string,
	recipient string,
	blockNumber uint32,
) error {
	ucInput := post_backup_transfer.UCPostBackupTransferInput{
		TransferHash:      encodedEncryptedTransferHash,
		EncryptedTransfer: encodedEncryptedTransfer,
		Recipient:         recipient,
		BlockNumber:       blockNumber,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/transfer", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return err
	}

	response := new(node.BackupTransferResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}

func (d *blockSynchronizer) BackupWithdrawal(
	recipient common.Address,
	encodedEncryptedTransferHash, encodedEncryptedTransfer string,
	blockNumber uint64,
) error {
	err := backupWithdrawalRawRequest(
		d.ctx,
		d.cfg,
		d.log,
		encodedEncryptedTransferHash,
		encodedEncryptedTransfer,
		recipient.Hex(),
		uint32(blockNumber),
	)

	if err != nil {
		return fmt.Errorf("failed to backup transfer: %w", err)
	}

	return nil
}

func backupWithdrawalRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	encodedEncryptedTransferHash, encodedEncryptedTransfer string,
	recipient string,
	blockNumber uint32,
) error {
	ucInput := post_backup_transfer.UCPostBackupTransferInput{
		TransferHash:      encodedEncryptedTransferHash,
		EncryptedTransfer: encodedEncryptedTransfer,
		Recipient:         recipient,
		BlockNumber:       blockNumber,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/transfer", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return err
	}

	response := new(node.BackupTransferResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}
