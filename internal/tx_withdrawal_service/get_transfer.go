package tx_withdrawal_service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/tx_transfer_service"
	"net/http"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type GetTransfersListFilter struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

type GetTransfersListPaginationCursor struct {
	BlockNumber  string `json:"blockNumber"`
	SortingValue string `json:"sortingValue"`
}

type GetTransfersListPagination struct {
	Direction string                            `json:"direction"`
	Limit     string                            `json:"limit"`
	Cursor    *GetTransfersListPaginationCursor `json:"cursor"`
}

type GetTransfersListInput struct {
	Sorting    string                      `json:"sorting"`
	Pagination *GetTransfersListPagination `json:"pagination"`
	Filter     *GetTransfersListFilter     `json:"filter"`
}

func GetTransfersListWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	input *GetTransfersListInput,
	userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		emptyKey  = ""
		indentKey = "  "
	)

	resp, err := getTransfersListRawRequest(
		ctx,
		cfg,
		input,
		userEthPrivateKey,
	)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		var js []byte
		js, err = json.MarshalIndent(&GetTxWithdrawalTransfersList{
			GetTransfersListError: GetTransfersListError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			},
		}, emptyKey, indentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal transfers list: %w", err)
		}

		return js, nil
	}

	transfersList := GetTxWithdrawalTransfersList{
		Success: true,
		Data: &GetTxWithdrawalTransfersListData{
			Transfers: make([]*tx_transfer_service.BackupWithdrawal, len(resp.Data.Transfers)),
		},
	}

	for key := range resp.Data.Transfers {
		var transferDetails *tx_transfer_service.BackupWithdrawal
		transferDetails, err = GetTransferFromBackupData(
			resp.Data.Transfers[key],
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get transfer from backup bata: %w", err)
		}

		transfersList.Data.Transfers[key] = transferDetails
	}

	var pg interface{}
	err = json.Unmarshal(resp.Data.Pagination, &pg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON with pagination: %w", err)
	}

	var js []byte
	js, err = json.MarshalIndent(transfersList, emptyKey, indentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal txList: %w", err)
	}

	arrTxDetails := gjson.GetBytes(js, "data.transfers").Array()
	for key := range arrTxDetails {
		var addr common.Address
		addr, err = transfersList.Data.Transfers[key].Transfer.Recipient.ToEthereumAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address to Ethereum address: %w", err)
		}

		js, err = sjson.SetBytes(
			js,
			fmt.Sprintf("data.transfers.%d.transfer.recipient", key),
			addr.String(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set recipient address: %w", err)
		}
	}

	js, err = sjson.SetBytes(js, "data.pagination", pg)
	if err != nil {
		return nil, fmt.Errorf("failed to update JSON with pagination: %w", err)
	}

	return js, nil
}

func getTransfersListRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	input *GetTransfersListInput,
	userEthPrivateKey string,
) (*GetTransfersListResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
		emptyKey    = ""
	)

	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		return nil, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	apiUrl := fmt.Sprintf("%s/v1/backups/transfers/list", cfg.API.DataStoreVaultUrl)

	body := map[string]interface{}{
		"sorting":   input.Sorting,
		"recipient": wallet.WalletAddress.String(),
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
		return nil, fmt.Errorf("failed to marshal body for send to route with transfers list: %w", err)
	}

	if strings.TrimSpace(input.Pagination.Cursor.BlockNumber) != emptyKey &&
		strings.TrimSpace(input.Pagination.Cursor.SortingValue) != emptyKey {
		jsBody, err = sjson.SetBytes(jsBody, "pagination.cursor", map[string]interface{}{
			"blockNumber":  input.Pagination.Cursor.BlockNumber,
			"sortingValue": input.Pagination.Cursor.SortingValue,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal pagination cursor of body for send to route with transfers list: %w", err)
		}
	}

	err = json.Unmarshal(jsBody, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body for send to route with transfers list: %w", err)
	}

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(body).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the transfer request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return &GetTransfersListResponse{
			Error: &GetTransfersListError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetTransfersListResponse{
			Error: &GetTransfersListError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetTransfersListResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func GetTransferFromBackupData(
	encryptedTransfer *GetTransferData,
) (*tx_transfer_service.BackupWithdrawal, error) {
	message, err := base64.StdEncoding.DecodeString(encryptedTransfer.EncryptedTransfer)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecodeFromBase64, err)
	}

	var bwd tx_transfer_service.BackupWithdrawal
	err = json.Unmarshal(message, &bwd)
	if err != nil {
		return nil, errors.Join(ErrFailedToUnmarshal, err)
	}

	return &bwd, nil
}

func GetTransferByHashWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	transferHash string,
	userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		emptyKey  = ""
		indentKey = "  "
	)

	resp, err := getTransferByHashRawRequest(
		ctx,
		cfg,
		transferHash,
		userEthPrivateKey,
	)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		var js []byte
		js, err = json.MarshalIndent(&GetTransferTxResponse{
			GetTransfersListError: GetTransfersListError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			},
		}, emptyKey, indentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal transfer by hash: %w", err)
		}

		return js, nil
	}

	var transferDetails *tx_transfer_service.BackupWithdrawal
	transferDetails, err = GetTransferFromBackupData(
		resp.Data.Transfer,
	)
	if err != nil {
		return nil, err
	}

	var js []byte
	js, err = json.MarshalIndent(&GetTransferTxResponse{
		Success: true,
		Data:    transferDetails,
	}, emptyKey, indentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transfer by hash: %w", err)
	}

	var address string
	if transferDetails.Transfer.Recipient.TypeOfAddress == intMaxAccTypes.INTMAXAddressType {
		var addr intMaxAcc.Address
		addr, err = transferDetails.Transfer.Recipient.ToINTMAXAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address to INTMAX address: %w", err)
		}
		address = addr.String()
	} else {
		var addr common.Address
		addr, err = transferDetails.Transfer.Recipient.ToEthereumAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address to Ethereum address: %w", err)
		}
		address = addr.String()
	}

	js, err = sjson.SetBytes(
		js,
		"data.transfer.recipient",
		address,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set recipient address: %w", err)
	}

	return js, nil
}

func getTransferByHashRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	transferHash string,
	userEthPrivateKey string,
) (*GetTransferByHashResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		return nil, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	apiUrl := fmt.Sprintf(
		"%s/v1/backups/transfer/%s",
		cfg.API.DataStoreVaultUrl,
		url.QueryEscape(transferHash),
	)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetQueryParams(map[string]string{
		"recipient": wallet.WalletAddress.String(),
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send of the transfer request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return &GetTransferByHashResponse{
			Error: &GetTransfersListError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetTransferByHashResponse{
			Error: &GetTransfersListError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetTransferByHashResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}
