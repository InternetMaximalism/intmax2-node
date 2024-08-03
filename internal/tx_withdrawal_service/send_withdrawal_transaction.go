package tx_transfer_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pow"
	"intmax2-node/internal/tx_transfer_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/use_cases/withdrawal_request"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
)

func SendWithdrawalRequest(
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
	log.Printf("transferTreeRoot: %v", transfersHash.String())
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

	err = tx_transfer_service.SendTransactionWithRawRequest(
		ctx, cfg, log, senderAccount, transfersHash, nonce, expiration, powNonceStr, signatureInput, encodedEncryptedTx, encodedEncryptedTransfers,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

func SendWithdrawalWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAccount *intMaxAcc.PrivateKey,
	transfer *intMaxTypes.Transfer,
	transferTreeRoot goldenposeidon.PoseidonHashOut,
	nonce uint64,
	transferMerkleProof []*goldenposeidon.PoseidonHashOut,
	transferIndex int32,
	txMerkleProof []*goldenposeidon.PoseidonHashOut,
	txIndex int32,
	blockNumber uint32,
	blockHash common.Hash,
) error {
	transferMerkleProofStr := make([]string, len(transferMerkleProof))
	for i, v := range transferMerkleProof {
		transferMerkleProofStr[i] = v.String()
	}
	txMerkleProofStr := make([]string, len(txMerkleProof))
	for i, v := range txMerkleProof {
		txMerkleProofStr[i] = v.String()
	}
	return sendWithdrawalRawRequest(
		ctx,
		cfg,
		log,
		transfer,
		transferTreeRoot.String(),
		nonce,
		transferMerkleProofStr,
		transferIndex,
		txMerkleProofStr,
		txIndex,
		blockNumber,
		blockHash.Hex(),
	)
}

func sendWithdrawalRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	transfer *intMaxTypes.Transfer,
	transferTreeRoot string,
	nonce uint64,
	transferMerkleProof []string,
	transferIndex int32,
	txMerkleProof []string,
	txIndex int32,
	blockNumber uint32,
	blockHash string,
) error {
	if transfer.Recipient.TypeOfAddress == "INTMAX" {
		return fmt.Errorf("intmax address is not supported")
	}

	transferHash := hexutil.Encode(transfer.Hash().Marshal())

	ethAddress := hexutil.Encode(transfer.Recipient.Address)
	ucInput := withdrawal_request.UCWithdrawalInput{
		TransferData: &withdrawal_request.TransferDataTransaction{
			Recipient:  ethAddress,
			TokenIndex: strconv.Itoa(int(transfer.TokenIndex)),
			Amount:     transfer.Amount.String(),
			Salt:       hexutil.Encode(transfer.Salt.Marshal()),
		},
		TransferMerkleProof: withdrawal_request.TransferMerkleProof{
			Siblings: transferMerkleProof,
			Index:    transferIndex,
		},
		Transaction: withdrawal_request.Transaction{
			Nonce:            int32(nonce),
			TransferTreeRoot: transferTreeRoot,
		},
		TxMerkleProof: withdrawal_request.TxMerkleProof{
			Siblings: txMerkleProof,
			Index:    txIndex,
		},
		TransferHash: transferHash,
		BlockNumber:  blockNumber,
		BlockHash:    blockHash,
		EnoughBalanceProof: withdrawal_request.EnoughBalanceProof{
			Proof:        "AA==", // dummy
			PublicInputs: "AA==", // dummy
		},
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

	apiUrl := fmt.Sprintf("%s/v1/withdrawals/request", cfg.HTTP.WithdrawalServerUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the withdrawal request: %w"
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

	response := new(SendWithdrawalResponse)
	if innerErr := json.Unmarshal(resp.Body(), response); innerErr != nil {
		ErrUnmarshalResponse := errors.New("failed to unmarshal response")
		return errors.Join(ErrUnmarshalResponse, innerErr)
	}

	if !response.Success {
		return errors.New("failed to request withdrawal")
	}

	return nil
}

type SendWithdrawalResponse struct {
	// Success is a flag that indicates the success of the request
	Success bool `json:"success"`

	// Data is a response data
	Data struct {
		// Message is a message from the server
		Message string `json:"message"`
	} `json:"data"`
}
