package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	"net/http"

	"github.com/go-resty/resty/v2"
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
			TxHash:      txDetails.TransferTreeRoot.String(),
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

	apiUrl := fmt.Sprintf("%s/v1/backups/transaction", cfg.API.DataStoreVaultUrl)

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
