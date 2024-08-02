package block_post_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gen/service/node"
	"intmax2-node/internal/use_cases/backup_transaction"
	"intmax2-node/internal/use_cases/backup_transfer"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func (d *blockPostService) BackupTransaction(
	sender intMaxAcc.Address,
	encodedEncryptedTx string,
	signature string,
	blockNumber uint64,
) error {
	err := backupTransactionRawRequest(
		d.ctx,
		d.cfg,
		d.log,
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
	encodedEncryptedTx string,
	signature string,
	sender string,
	blockNumber uint32,
) error {
	ucInput := backup_transaction.UCPostBackupTransactionInput{
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

	apiUrl := fmt.Sprintf("%s/v1/backups/transaction", cfg.HTTP.DataStoreVaultUrl)

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

func (d *blockPostService) BackupTransfer(
	recipient intMaxAcc.Address,
	encodedEncryptedTransfer string,
	blockNumber uint64,
) error {
	err := backupTransferRawRequest(
		d.ctx,
		d.cfg,
		d.log,
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
	encodedEncryptedTransfer string,
	recipient string,
	blockNumber uint32,
) error {
	ucInput := backup_transfer.UCPostBackupTransferInput{
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

	apiUrl := fmt.Sprintf("%s/v1/backups/transfer", cfg.HTTP.DataStoreVaultUrl)

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
