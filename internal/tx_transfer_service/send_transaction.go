package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pow"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"net/http"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
)

func SendTransactionRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash intMaxTypes.PoseidonHashOut,
	nonce uint64,
	encodedEncryptedTx *transaction.BackupTransactionData,
	encodedEncryptedTransfers []*transaction.BackupTransferInput,
) error {
	const duration = 300 * time.Minute
	expiration := time.Now().Add(duration)

	pw := pow.New(cfg.PoW.Difficulty)
	pWorker := pow.NewWorker(cfg.PoW.Workers, pw)
	pwNonce := pow.NewPoWNonce(pw, pWorker)

	tx, err := intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()
	log.Printf("transfersHash: %v", transfersHash.String())
	messageForPow := txHash.Marshal()
	powNonceStr, err := pwNonce.Nonce(ctx, messageForPow)
	if err != nil {
		return fmt.Errorf("failed to get PoW nonce: %w", err)
	}

	err = pwNonce.Verify(powNonceStr, messageForPow)
	if err != nil {
		panic(fmt.Sprintf("failed to verify PoW nonce: %v", err))
	}

	message, err := transaction.MakeMessage(
		transfersHash.Marshal(),
		nonce,
		powNonceStr,
		senderAccount.ToAddress(),
		expiration,
	)
	if err != nil {
		return fmt.Errorf("failed to make message: %w", err)
	}

	signatureInput, err := senderAccount.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	return SendTransactionWithRawRequest(
		ctx, cfg, log, senderAccount, transfersHash, nonce, expiration, powNonceStr, signatureInput,
		encodedEncryptedTx, encodedEncryptedTransfers,
	)
}

func SendTransactionWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash intMaxTypes.PoseidonHashOut,
	nonce uint64,
	expiration time.Time,
	powNonce string,
	signature *bn254.G2Affine,
	encodedEncryptedTx *transaction.BackupTransactionData,
	encodedEncryptedTransfers []*transaction.BackupTransferInput,
) error {
	return sendTransactionRawRequest(
		ctx,
		cfg,
		log,
		senderAccount.ToAddress().String(),
		transfersHash.String(),
		nonce,
		expiration,
		powNonce,
		hexutil.Encode(signature.Marshal()),
		encodedEncryptedTx,
		encodedEncryptedTransfers,
	)
}

func sendTransactionRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress, transfersHash string,
	nonce uint64,
	expiration time.Time,
	powNonce, signature string,
	backupTx *transaction.BackupTransactionData,
	backupTransfers []*transaction.BackupTransferInput,
) error {
	fmt.Printf("backupTx: %s\n", backupTx)
	ucInput := transaction.UCTransactionInput{
		Sender:          senderAddress,
		TransfersHash:   transfersHash,
		Nonce:           nonce,
		PowNonce:        powNonce,
		Expiration:      expiration,
		Signature:       signature,
		BackupTx:        backupTx,
		BackupTransfers: backupTransfers,
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

	apiUrl := fmt.Sprintf("%s/v1/transaction", cfg.HTTP.BlockBuilderUrl)

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

	response := new(SendTransactionResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}
