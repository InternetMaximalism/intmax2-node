package block_synchronizer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	"intmax2-node/internal/use_cases/block_signature"
	postBackupTransaction "intmax2-node/internal/use_cases/post_backup_transaction"
	"intmax2-node/internal/use_cases/post_backup_transfer"
	"net/http"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

const ErrFailedToUnmarshalResponse = "failed to unmarshal response: %w"
const unexpectedStatusCode = "unexpected status code"
const ErrFailedToMarshalJSON = "failed to marshal JSON: %w"

func (d *blockSynchronizer) BackupTransaction(
	sender intMaxAcc.Address,
	txHash, encodedEncryptedTx string,
	senderLastBalanceProofBody, senderBalanceTransitionProofBody []byte,
	signature string,
	blockNumber uint64,
) error {
	err := backupTransactionRawRequest(
		d.ctx,
		d.cfg,
		d.log,
		txHash,
		encodedEncryptedTx,
		senderLastBalanceProofBody,
		senderBalanceTransitionProofBody,
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
	senderLastBalanceProofBody, senderBalanceTransitionProofBody []byte,
	signature string,
	sender string,
	blockNumber uint32,
) error {
	senderEnoughBalanceProof := block_signature.EnoughBalanceProofBody{
		PrevBalanceProofBody:  senderLastBalanceProofBody,
		TransferStepProofBody: senderBalanceTransitionProofBody,
	}
	senderEnoughBalanceProofInput := new(block_signature.EnoughBalanceProofBodyInput).FromEnoughBalanceProofBody(&senderEnoughBalanceProof)
	ucInput := postBackupTransaction.UCPostBackupTransactionInput{
		TxHash:                       txHash,
		EncryptedTx:                  encodedEncryptedTx,
		SenderEnoughBalanceProofBody: senderEnoughBalanceProofInput,
		Sender:                       sender,
		BlockNumber:                  blockNumber,
		Signature:                    signature,
		// SenderLastBalanceProofBody:       senderLastBalanceProofBody,
		// SenderBalanceTransitionProofBody: senderBalanceTransitionProofBody,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf(ErrFailedToMarshalJSON, err)
	}

	const (
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
		const msg = "failed to get backup transaction: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return err
	}

	response := new(node.BackupTransactionResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf(ErrFailedToUnmarshalResponse, err)
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
		return fmt.Errorf(ErrFailedToMarshalJSON, err)
	}

	const (
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
		const msg = "failed to get backup transaction: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return err
	}

	response := new(node.BackupTransferResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf(ErrFailedToUnmarshalResponse, err)
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
		return fmt.Errorf(ErrFailedToMarshalJSON, err)
	}

	const (
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
		const msg = "failed to get backup transaction: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return err
	}

	response := new(node.BackupTransferResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf(ErrFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}

func GetBackupSenderBalanceProofs(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	hashes []string,
) (*BackupBalanceProofsData, error) {
	fmt.Printf("(GetBackupSenderBalanceProofs) hashes: %v\n", hashes)
	backupBalanceProofs, err := getBackupSenderBalanceProofsRawRequest(
		ctx,
		cfg,
		log,
		hashes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to backup balance proof: %w", err)
	}

	return backupBalanceProofs, nil
}

type BackupBalanceProofData struct {
	EnoughBalanceProofBodyHash string
	LastBalanceProofBody       string
	BalanceTransitionProofBody string
}

type BackupBalanceProofsData struct {
	Proofs []BackupBalanceProofData
}

// BackupBalanceRequest
func getBackupSenderBalanceProofsRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	hashes []string,
) (*BackupBalanceProofsData, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	u, err := url.Parse(fmt.Sprintf("%s/v1/backups/proof", cfg.API.DataStoreVaultUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for _, hash := range hashes {
		q.Add("hashes", hash)
	}
	u.RawQuery = q.Encode()

	fmt.Printf("url: %s\n", u.String())

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(u.String())
	if err != nil {
		const msg = "failed to send of the sender proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "sender proof request error occurred"
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(node.GetBackupBalanceProofsResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(ErrFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to balance proof: %s", response.Data)
	}

	proofs := make([]BackupBalanceProofData, 0)
	for _, proof := range response.Data.Proofs {
		proofs = append(proofs, BackupBalanceProofData{
			EnoughBalanceProofBodyHash: proof.EnoughBalanceProofBodyHash,
			LastBalanceProofBody:       proof.LastBalanceProofBody,
			BalanceTransitionProofBody: proof.BalanceTransitionProofBody,
		})
	}

	return &BackupBalanceProofsData{
		Proofs: proofs,
	}, nil
}

func BackupBalanceProof(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sender intMaxAcc.Address,
	prevID string,
	balanceProofBody, encryptedBalanceData string,
	encryptedTxs, encryptedTransfers, encryptedDeposits []string,
	signature string,
	blockNumber uint64,
) (*BackupBalanceData, error) {
	newBackupBalance, err := backupBalanceProofRawRequest(
		ctx,
		cfg,
		log,
		prevID,
		balanceProofBody,
		encryptedBalanceData,
		encryptedTxs,
		encryptedTransfers,
		encryptedDeposits,
		signature,
		sender.String(),
		uint32(blockNumber),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to backup balance proof: %w", err)
	}

	return newBackupBalance, nil
}

// BackupBalanceRequest
func backupBalanceProofRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	prevID, balanceProofBody, encryptedBalanceData string,
	encryptedTxs, encryptedTransfers, encryptedDeposits []string,
	signature string,
	user string,
	blockNumber uint32,
) (*BackupBalanceData, error) {
	fmt.Printf("size of balanceProofBody: %d", len(balanceProofBody))
	ucInput := node.BackupBalanceRequest{
		User:                  user,
		EncryptedBalanceProof: balanceProofBody,
		EncryptedBalanceData:  encryptedBalanceData,
		EncryptedTxs:          encryptedTxs,
		EncryptedTransfers:    encryptedTransfers,
		EncryptedDeposits:     encryptedDeposits,
		Signature:             signature,
		BlockNumber:           uint64(blockNumber),
		PrevId:                prevID,
	}

	bd, err := json.Marshal(&ucInput)
	if err != nil {
		return nil, fmt.Errorf(ErrFailedToMarshalJSON, err)
	}
	// fmt.Printf("bd: %s", bd)

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/balance", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the balance proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "balance proof request error occurred"
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(node.BackupBalanceResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(ErrFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to balance proof: %s", response.Data)
	}

	return &BackupBalanceData{
		ID:                   response.Data.Balance.Id,
		BalanceProofBody:     response.Data.Balance.EncryptedBalanceProof,
		EncryptedBalanceData: response.Data.Balance.EncryptedBalanceData,
		EncryptedTxs:         response.Data.Balance.EncryptedTxs,
		EncryptedTransfers:   response.Data.Balance.EncryptedTransfers,
		EncryptedDeposits:    response.Data.Balance.EncryptedDeposits,
		BlockNumber:          response.Data.Balance.BlockNumber,
		CreatedAt:            response.Data.Balance.CreatedAt,
	}, nil
}

func GetBackupBalance(
	ctx context.Context,
	cfg *configs.Config,
	userPublicKey *intMaxAcc.PublicKey,
) (*BackupBalanceData, error) {
	resp, err := getBackupBalanceRawRequest(
		ctx,
		cfg,
		userPublicKey.ToAddress().String(),
	)
	if err != nil {
		if err.Error() == "no assets found" || err.Error() == "failed to start Balance Prover Service: no assets found" {
			// default value
			return &BackupBalanceData{
				ID:                   "",
				BalanceProofBody:     "",
				EncryptedBalanceData: "",
				BlockNumber:          0,
			}, nil
		}

		fmt.Printf("fail to GetBackupBalance: %v\n", err.Error())
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to get backup balance: %s", resp.Error.Message)
	}

	return resp.Data, nil
}

func getBackupBalanceRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	user string,
) (*GetBackupBalanceResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
		emptyKey    = ""
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/balance?sender=%s", cfg.API.DataStoreVaultUrl, user)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get backup transaction: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return &GetBackupBalanceResponse{
			Error: &GetBackupBalanceError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetBackupBalanceResponse{
			Error: &GetBackupBalanceError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetBackupBalancesResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(ErrFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		if response.Error == nil {
			return nil, fmt.Errorf("failed to get backup balance with unknown reason")
		}

		return nil, fmt.Errorf("failed to get backup balance: %s", response.Error.Message)
	}

	if response.Data == nil {
		return nil, ErrNoAssetsFound
	}

	if len(response.Data.Balances) == 0 {
		return nil, ErrNoAssetsFound
	}

	result := new(GetBackupBalanceResponse)
	result.Success = response.Success
	balanceData := response.Data.Balances[0] // latest
	result.Data = &BackupBalanceData{
		ID:                   balanceData.ID,
		BalanceProofBody:     balanceData.EncryptedBalanceProof,
		EncryptedBalanceData: balanceData.EncryptedBalanceData,
		EncryptedTxs:         balanceData.EncryptedTxs,
		EncryptedTransfers:   balanceData.EncryptedTransfers,
		EncryptedDeposits:    balanceData.EncryptedDeposits,
		BlockNumber:          balanceData.BlockNumber,
		CreatedAt:            balanceData.CreatedAt,
	}
	fmt.Printf("result.Data: %+v\n", result.Data)

	return result, nil
}
