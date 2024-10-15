package tx_transfer_service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pow"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
	"github.com/holiman/uint256"
)

func SendTransferTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash intMaxTypes.PoseidonHashOut,
	nonce uint32,
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
		uint64(nonce),
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
	)
}

func SendTransactionWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash intMaxTypes.PoseidonHashOut,
	nonce uint32,
	expiration time.Time,
	powNonce string,
	signature *bn254.G2Affine,
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
	)
}

func sendTransactionRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress, transfersHash string,
	nonce uint32,
	expiration time.Time,
	powNonce, signature string,
) error {
	ucInput := transaction.UCTransactionInput{
		Sender:        senderAddress,
		TransfersHash: transfersHash,
		Nonce:         uint64(nonce),
		PowNonce:      powNonce,
		Expiration:    expiration,
		Signature:     signature,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/transaction", cfg.API.BlockBuilderUrl)

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
			"api_url":     apiUrl,
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

func transferFee(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	tokenIndex uint32,
) (*uint256.Int, string, error) {
	const (
		emptyKey           = ""
		ctrlKey            = '\n'
		yKey               = "y"
		nKey               = "n"
		uKey               = "u"
		msgFetch           = "Fetching transfer fee..."
		maskETH            = "Current transfer fee is %.18f ETH. For continue transaction press, please: Y|n|u (yes|no|update):"
		msgTimeout         = "Confirmation of transfer fee is expired (time out equal to or greater than 1 minute). Repeat your selection."
		msgTrFeeIsApproved = "Transfer fee is approved."
		msgTrIsCanceled    = "Transaction is canceled."
		defaultTokenIndex  = "0"
		val1e18Key         = 1e18
	)

	var amountGasFee uint256.Int
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(msgFetch)

		dataBlockInfo, err := GetBlockInfo(ctx, cfg, log)
		if err != nil {
			const msg = "failed to get the block info data: %w"
			return nil, emptyKey, fmt.Errorf(msg, err)
		}
		gasFee, gasOK := dataBlockInfo.TransferFee[new(big.Int).SetUint64(uint64(tokenIndex)).String()]
		if !gasOK {
			gasFee, gasOK = dataBlockInfo.TransferFee[defaultTokenIndex]
		}
		if !gasOK {
			const msg = "failed to get default gas fee from the block info data"
			return nil, emptyKey, fmt.Errorf(msg)
		}

		err = amountGasFee.Scan(gasFee)
		if err != nil {
			const msg = "failed to convert string to uint256.Int: %w"
			return nil, emptyKey, fmt.Errorf(msg, err)
		}

		start := time.Now().UTC()
		fmt.Printf(maskETH, amountGasFee.Float64()/val1e18Key)

		var rs string
		rs, err = reader.ReadString(ctrlKey)
		if err != nil {
			continue
		}
		rs = strings.ToLower(strings.TrimSpace(rs))
		switch rs {
		case yKey:
			if time.Now().UTC().Unix() >= start.Add(time.Minute).Unix() {
				fmt.Println(msgTimeout)
				continue
			}

			fmt.Println(msgTrFeeIsApproved)

			return &amountGasFee, dataBlockInfo.IntMaxAddress, nil
		case nKey:
			fmt.Println(msgTrIsCanceled)

			return nil, emptyKey, nil
		case uKey:
			continue
		default:
			continue
		}
	}
}
