package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	"net/http"
	"net/url"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func GetTransactionsListWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	startBlockNumber, limit uint64,
	senderAccount *intMaxAcc.PrivateKey,
) (json.RawMessage, error) {
	resp, err := getTransactionsListRawRequest(
		ctx,
		cfg,
		startBlockNumber, limit,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	transactionsList := make([]*GetTransactionsListTransaction, len(resp.Transactions))
	for key := range resp.Transactions {
		var txDetails *intMaxTypes.TxDetails
		txDetails, err = GetTransactionFromBackupData(
			resp.Transactions[key],
			senderAccount,
		)
		if err != nil {
			return nil, err
		}
		transactionsList[key] = &GetTransactionsListTransaction{
			BlockNumber: resp.Transactions[key].BlockNumber,
			TxHash:      txDetails.Hash().String(),
			CreatedAt:   resp.Transactions[key].CreatedAt,
		}
	}

	txList := GetTransactionsList{
		Transactions: transactionsList,
		Meta:         resp.Meta,
	}

	return json.MarshalIndent(txList, "", "  ")
}

func getTransactionsListRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	startBlockNumber, limit uint64,
	senderAccount *intMaxAcc.PrivateKey,
) (*GetTransactionsListData, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/transactions", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetQueryParams(map[string]string{
		"sender":           senderAccount.ToAddress().String(),
		"limit":            fmt.Sprintf("%d", limit),
		"startBlockNumber": fmt.Sprintf("%d", startBlockNumber),
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get response")
	}

	response := new(GetTransactionsListResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get transfers list: %s", response.Error.Message)
	}

	return response.Data, nil
}

func GetTransactionByHashWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	txHash string,
	senderAccount *intMaxAcc.PrivateKey,
) (json.RawMessage, error) {
	resp, err := getTransactionByHashRawRequest(
		ctx,
		cfg,
		txHash,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	var txDetails *intMaxTypes.TxDetails
	txDetails, err = GetTransactionFromBackupData(
		resp.Transaction,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	var sign *bn254.G2Affine
	sign, err = GetSignatureFromBackupData(
		resp.Transaction.Signature,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	js, err := json.MarshalIndent(&GetTransactionTxData{
		ID:          resp.Transaction.ID,
		Sender:      resp.Transaction.Sender,
		Signature:   hexutil.Encode(sign.Marshal()),
		BlockNumber: resp.Transaction.BlockNumber,
		TxDetails:   txDetails,
		CreatedAt:   resp.Transaction.CreatedAt,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction by hash: %w", err)
	}

	arrTxDetails := gjson.GetBytes(js, "txDetails.Transfers").Array()
	for key := range arrTxDetails {
		var address string
		if arrTxDetails[key].Get("Recipient.TypeOfAddress").String() == "INTMAX" {
			var addr intMaxAcc.Address
			addr, err = txDetails.Transfers[key].Recipient.ToINTMAXAddress()
			if err != nil {
				return nil, fmt.Errorf("failed to convert recipient address to INTMAX address: %w", err)
			}
			address = addr.String()
		} else {
			var addr common.Address
			addr, err = txDetails.Transfers[key].Recipient.ToEthereumAddress()
			if err != nil {
				return nil, fmt.Errorf("failed to convert recipient address to Ethereum address: %w", err)
			}
			address = addr.String()
		}

		js, err = sjson.SetBytes(
			js,
			fmt.Sprintf("txDetails.Transfers.%d.Recipient.Address", key),
			address,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set recipient address: %w", err)
		}
	}

	return js, nil
}

func getTransactionByHashRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	txHash string,
	senderAccount *intMaxAcc.PrivateKey,
) (*GetTransactionByHashData, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf(
		"%s/v1/backups/transaction/%s",
		cfg.API.DataStoreVaultUrl,
		url.QueryEscape(txHash),
	)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetQueryParams(map[string]string{
		"sender": senderAccount.ToAddress().String(),
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() == http.StatusBadRequest {
		return nil, fmt.Errorf("not found")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get response")
	}

	response := new(GetTransactionByHashResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get transaction by hash")
	}

	return response.Data, nil
}
