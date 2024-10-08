package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	intMaxTypes "intmax2-node/internal/types"
	"net/http"
	"net/url"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type GetTransactionsListFilter struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

type GetTransactionsListPaginationCursor struct {
	BlockNumber  string `json:"blockNumber"`
	SortingValue string `json:"sortingValue"`
}

type GetTransactionsListPagination struct {
	Direction string                               `json:"direction"`
	Limit     string                               `json:"limit"`
	Cursor    *GetTransactionsListPaginationCursor `json:"cursor"`
}

type GetTransactionsListInput struct {
	Sorting    string                         `json:"sorting"`
	Pagination *GetTransactionsListPagination `json:"pagination"`
	Filter     *GetTransactionsListFilter     `json:"filter"`
}

func GetTransactionsListWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	input *GetTransactionsListInput,
	senderAccount *intMaxAcc.PrivateKey,
) (json.RawMessage, error) {
	const (
		emptyKey  = ""
		indentKey = "  "
	)

	resp, err := getTransactionsListRawRequest(
		ctx,
		cfg,
		input,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		var js []byte
		js, err = json.MarshalIndent(&GetTransactionsList{
			GetTransactionsListError: GetTransactionsListError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			},
		}, emptyKey, indentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal txList: %w", err)
		}

		return js, nil
	}

	txList := GetTransactionsList{
		Success: true,
		Data:    &GetTxTransactionsListData{TxHashes: make([]string, len(resp.Data.Transactions))},
	}

	for key := range resp.Data.Transactions {
		var txDetails *intMaxTypes.TxDetails
		txDetails, err = GetTransactionFromBackupData(
			resp.Data.Transactions[key],
			senderAccount,
		)
		if err != nil {
			return nil, err
		}
		txList.Data.TxHashes[key] = txDetails.Hash().String()
	}

	var pg interface{}
	err = json.Unmarshal(resp.Data.Pagination, &pg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON with pagination: %w", err)
	}

	var js []byte
	js, err = json.MarshalIndent(txList, emptyKey, indentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal txList: %w", err)
	}

	js, err = sjson.SetBytes(js, "data.pagination", pg)
	if err != nil {
		return nil, fmt.Errorf("failed to update JSON with pagination: %w", err)
	}

	return js, nil
}

func getTransactionsListRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	input *GetTransactionsListInput,
	senderAccount *intMaxAcc.PrivateKey,
) (*GetTransactionsListResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
		emptyKey    = ""
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/transactions/list", cfg.API.DataStoreVaultUrl)

	body := map[string]interface{}{
		"sorting": input.Sorting,
		"sender":  senderAccount.ToAddress().String(),
		"pagination": map[string]interface{}{
			"direction": input.Pagination.Direction,
			"perPage":   input.Pagination.Limit,
		},
	}

	if strings.TrimSpace(input.Filter.Name) != emptyKey &&
		strings.TrimSpace(input.Filter.Condition) != emptyKey &&
		strings.TrimSpace(input.Filter.Value) != emptyKey {
		body["filter"] = []map[string]interface{}{
			{
				"relation":  "and",
				"dataField": input.Filter.Name,
				"condition": input.Filter.Condition,
				"value":     input.Filter.Value,
			},
		}
	}

	jsBody, err := json.Marshal(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body for send to route with tx list: %w", err)
	}

	if strings.TrimSpace(input.Pagination.Cursor.BlockNumber) != emptyKey &&
		strings.TrimSpace(input.Pagination.Cursor.SortingValue) != emptyKey {
		jsBody, err = sjson.SetBytes(jsBody, "pagination.cursor", map[string]interface{}{
			"blockNumber":  input.Pagination.Cursor.BlockNumber,
			"sortingValue": input.Pagination.Cursor.SortingValue,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal pagination cursor of body for send to route with tx list: %w", err)
		}
	}

	err = json.Unmarshal(jsBody, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body for send to route with tx list: %w", err)
	}

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(body).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return &GetTransactionsListResponse{
			Error: &GetTransactionsListError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetTransactionsListResponse{
			Error: &GetTransactionsListError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetTransactionsListResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func GetTransactionByHashWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	txHash string,
	senderAccount *intMaxAcc.PrivateKey,
) (json.RawMessage, error) {
	const (
		emptyKey  = ""
		indentKey = "  "
	)

	resp, err := getTransactionByHashRawRequest(
		ctx,
		cfg,
		txHash,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		var js []byte
		js, err = json.MarshalIndent(&GetTransactionTxResponse{
			GetTransactionsListError: GetTransactionsListError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			},
		}, emptyKey, indentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal transaction by hash: %w", err)
		}

		return js, nil
	}

	var txDetails *intMaxTypes.TxDetails
	txDetails, err = GetTransactionFromBackupData(
		resp.Data.Transaction,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	var sign *bn254.G2Affine
	sign, err = GetSignatureFromBackupData(
		resp.Data.Transaction.Signature,
		senderAccount,
	)
	if err != nil {
		return nil, err
	}

	var js []byte
	js, err = json.MarshalIndent(&GetTransactionTxResponse{
		Success: true,
		Data: &GetTransactionTxData{
			ID:          resp.Data.Transaction.ID,
			Sender:      resp.Data.Transaction.Sender,
			Signature:   hexutil.Encode(sign.Marshal()),
			BlockNumber: resp.Data.Transaction.BlockNumber,
			TxDetails:   txDetails,
			CreatedAt:   resp.Data.Transaction.CreatedAt,
		},
	}, emptyKey, indentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction by hash: %w", err)
	}

	arrTxDetails := gjson.GetBytes(js, "data.txDetails.Transfers").Array()
	for key := range arrTxDetails {
		var address string
		if arrTxDetails[key].Get("Recipient.TypeOfAddress").String() == intMaxAccTypes.INTMAXAddressType {
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
			fmt.Sprintf("data.txDetails.Transfers.%d.Recipient.Address", key),
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
) (*GetTransactionByHashResponse, error) {
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
		return &GetTransactionByHashResponse{
			Error: &GetTransactionsListError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetTransactionByHashResponse{
			Error: &GetTransactionsListError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetTransactionByHashResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}
