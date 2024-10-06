package tx_transfer_service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/pow"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
	"github.com/gosuri/uilive"
	"github.com/holiman/uint256"
)

func SendTransferTransaction(
	ctx context.Context,
	cfg *configs.Config,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash intMaxTypes.PoseidonHashOut,
	nonce uint64,
	// encodedEncryptedTx *transaction.BackupTransactionData,
	// encodedEncryptedTransfers []*transaction.BackupTransferInput,
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
		ctx, cfg, senderAccount, transfersHash, nonce, expiration, powNonceStr, signatureInput,
		// encodedEncryptedTx, encodedEncryptedTransfers,
	)
}

func SendTransactionWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash intMaxTypes.PoseidonHashOut,
	nonce uint64,
	expiration time.Time,
	powNonce string,
	signature *bn254.G2Affine,
	// encodedEncryptedTx *transaction.BackupTransactionData,
	// encodedEncryptedTransfers []*transaction.BackupTransferInput,
) error {
	return sendTransactionRawRequest(
		ctx,
		cfg,
		senderAccount.ToAddress().String(),
		transfersHash.String(),
		nonce,
		expiration,
		powNonce,
		hexutil.Encode(signature.Marshal()),
		// encodedEncryptedTx,
		// encodedEncryptedTransfers,
	)
}

func sendTransactionRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	senderAddress, transfersHash string,
	nonce uint64,
	expiration time.Time,
	powNonce, signature string,
	// backupTx *transaction.BackupTransactionData,
	// backupTransfers []*transaction.BackupTransferInput,
) error {
	ucInput := transaction.UCTransactionInput{
		Sender:        senderAddress,
		TransfersHash: transfersHash,
		Nonce:         nonce,
		PowNonce:      powNonce,
		Expiration:    expiration,
		Signature:     signature,
		// BackupTx:        backupTx,
		// BackupTransfers: backupTransfers,
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
		return fmt.Errorf("failed to get response")
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

func sendTransactionGasFeeInfoMessage(
	ctx context.Context,
	wg *sync.WaitGroup,
	writer *uilive.Writer,
	noCheck *bool,
	feeStr string,
) bool {
	if wg != nil {
		defer wg.Done()
	}
	if writer == nil {
		return false
	}
	var (
		i  = 60
		tm = time.NewTicker(time.Second)
	)
	for {
		select {
		case <-ctx.Done():
			tm.Stop()
			return false
		case <-tm.C:
			if i <= 0 {
				tm.Stop()
				return true
			}
			const msg = "Current transfer fee value is %s. For continue transaction press, please: 'y'+Enter, for cancel - 'n'+Enter. Next update transfer fee do after %d seconds\n"
			_, _ = fmt.Fprintf(writer, msg, feeStr, i)
			i--
			*noCheck = false
		}
	}
}

func transferFee(
	inputCtx context.Context,
	cfg *configs.Config,
	tokenIndex uint32,
) (*uint256.Int, string, error) {
	const (
		int1Key           = 1
		emptyKey          = ""
		ctrlKey           = '\n'
		yKey              = "y"
		nKey              = "n"
		maskETH           = "%v ETH"
		defaultTokenIndex = "0"
		val1e18Key        = 1e18
	)

	dataBlockInfo, err := GetBlockInfo(inputCtx, cfg)
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

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(inputCtx)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	wg.Add(int1Key)
	var next, noCheck bool
	noCheck = true
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {
			var rs string
			rs, err = reader.ReadString(ctrlKey)
			if err != nil {
				continue
			}
			rs = strings.ToLower(strings.TrimSpace(rs))
			switch rs {
			case yKey:
				if noCheck {
					continue
				}
				next = true
				fallthrough
			case nKey:
				if noCheck {
					continue
				}
				if cancel != nil {
					cancel()
				}
				return
			default:
				continue
			}
		}
	}()
	var amountGasFee uint256.Int
	errInfoMessage := make(chan error)
	wg.Add(int1Key)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				errInfoMessage <- nil
				return
			default:
				err = amountGasFee.Scan(gasFee)
				if err != nil {
					const msg = "failed to convert string to uint256.Int: %w"
					errInfoMessage <- fmt.Errorf(msg, err)
					return
				}

				wg.Add(int1Key)
				if sendTransactionGasFeeInfoMessage(
					ctx, &wg, writer, &noCheck, fmt.Sprintf(maskETH, amountGasFee.Float64()/val1e18Key),
				) {
					noCheck = true
					const msgFetchNewFeeV = "Fetching new transfer fee value...\n"
					_, _ = fmt.Fprintf(writer, msgFetchNewFeeV)
					dataBlockInfo, err = GetBlockInfo(inputCtx, cfg)
					if err != nil {
						const msg = "failed to get the block info data: %w"
						errInfoMessage <- fmt.Errorf(msg, err)
						return
					}
					gasFee, gasOK = dataBlockInfo.TransferFee[new(big.Int).SetUint64(uint64(tokenIndex)).String()]
					if !gasOK {
						gasFee, gasOK = dataBlockInfo.TransferFee[defaultTokenIndex]
					}
					if !gasOK {
						const msg = "failed to get default gas fee from the block info data"
						errInfoMessage <- fmt.Errorf(msg)
						return
					}
				}
			}
		}
	}()
	if errIM := <-errInfoMessage; errIM != nil {
		return nil, emptyKey, errIM
	}
	wg.Wait()
	if next {
		const msgTrFeeIsApproved = "Transfer fee is approved.\n"
		_, _ = fmt.Fprintf(writer.Bypass(), msgTrFeeIsApproved)
		return &amountGasFee, dataBlockInfo.IntMaxAddress, nil
	}
	const msgTrIsCanceled = "Transaction is canceled.\n"
	_, _ = fmt.Fprintf(writer.Bypass(), msgTrIsCanceled)

	return nil, emptyKey, nil
}
