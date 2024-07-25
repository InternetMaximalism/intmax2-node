package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/pow"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func SendTransactionRequest(
	cfg *configs.Config,
	ctx context.Context,
	senderAccount *intMaxAcc.PrivateKey,
	txHash intMaxTypes.PoseidonHashOut,
	nonce uint64,
) error {
	const duration = 60 * time.Minute
	expiration := time.Now().Add(duration)

	pw := pow.New(cfg.PoW.Difficulty)
	pWorker := pow.NewWorker(cfg.PoW.Workers, pw)
	pwNonce := pow.NewPoWNonce(pw, pWorker)

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
		txHash.Marshal(),
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

	return SendTransactionWithRawRequest(senderAccount, txHash, nonce, expiration, powNonceStr, signatureInput)
}

func SendTransactionWithRawRequest(
	senderAccount *intMaxAcc.PrivateKey,
	txHash intMaxTypes.PoseidonHashOut,
	nonce uint64,
	expiration time.Time,
	powNonce string,
	signature *bn254.G2Affine,
) error {
	return sendTransactionRawRequest(
		senderAccount.ToAddress().String(),
		txHash.String(),
		nonce,
		expiration,
		powNonce,
		hexutil.Encode(signature.Marshal()),
	)
}

func sendTransactionRawRequest(
	senderAddress, txHash string,
	nonce uint64,
	expiration time.Time,
	powNonce, signature string,
) error {
	ucInput := transaction.UCTransactionInput{
		Sender:        senderAddress,
		TransfersHash: txHash,
		Nonce:         nonce,
		PowNonce:      powNonce,
		Expiration:    expiration,
		Signature:     signature,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	body := strings.NewReader(string(bd))

	const (
		apiBaseUrl  = "http://localhost"
		contentType = "application/json"
	)
	apiUrl := fmt.Sprintf("%s/v1/transaction", apiBaseUrl)

	resp, err := http.Post(apiUrl, contentType, body) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d", resp.StatusCode)
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		return fmt.Errorf("response body: %s", string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	response := new(SendTransactionResponse)
	if err = json.Unmarshal(bodyBytes, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}
