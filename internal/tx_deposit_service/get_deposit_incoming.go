package tx_deposit_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/sjson"
)

type GetDepositsListFilter struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

type GetDepositsListPaginationCursor struct {
	BlockNumber  string `json:"blockNumber"`
	SortingValue string `json:"sortingValue"`
}

type GetDepositsListPagination struct {
	Direction string                           `json:"direction"`
	Limit     string                           `json:"limit"`
	Cursor    *GetDepositsListPaginationCursor `json:"cursor"`
}

type GetDepositsListInput struct {
	Sorting    string                     `json:"sorting"`
	Pagination *GetDepositsListPagination `json:"pagination"`
	Filter     *GetDepositsListFilter     `json:"filter"`
}

func GetDepositsListWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	input *GetDepositsListInput,
	receiverAccount *intMaxAcc.PrivateKey,
) (json.RawMessage, error) {
	resp, err := getDepositsListRawRequest(
		ctx,
		cfg,
		input,
		receiverAccount,
	)
	if err != nil {
		return nil, err
	}

	depositsList := GetDepositsList{
		Deposits: make([]*Deposit, len(resp.Deposits)),
	}

	for key := range resp.Deposits {
		var deposit *intMaxTypes.Deposit
		deposit, err = GetDepositFromBackupData(
			resp.Deposits[key],
			receiverAccount,
		)
		if err != nil {
			return nil, err
		}

		recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(deposit.Recipient.Pk.X.BigInt(new(big.Int)), deposit.Salt)

		depositHash := new(intMaxTypes.DepositLeaf).Set(&intMaxTypes.DepositLeaf{
			RecipientSaltHash: recipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		}).Hash()

		depositsList.Deposits[key] = &Deposit{
			Hash:       depositHash.String(),
			Recipient:  deposit.Recipient.String(),
			TokenIndex: deposit.TokenIndex,
			Amount:     deposit.Amount,
			Salt:       deposit.Salt.String(),
		}
	}

	var pg interface{}
	err = json.Unmarshal(resp.Pagination, &pg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON with pagination: %w", err)
	}

	var js []byte
	js, err = json.MarshalIndent(depositsList, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal depositsList: %w", err)
	}

	js, err = sjson.SetBytes(js, "pagination", pg)
	if err != nil {
		return nil, fmt.Errorf("failed to update JSON with pagination: %w", err)
	}

	return js, nil
}

func getDepositsListRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	input *GetDepositsListInput,
	recipientAccount *intMaxAcc.PrivateKey,
) (*GetDepositsListData, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
		emptyKey    = ""
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/deposits/list", cfg.API.DataStoreVaultUrl)

	body := map[string]interface{}{
		"sorting":   input.Sorting,
		"recipient": recipientAccount.ToAddress().String(),
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
		return nil, fmt.Errorf("failed to marshal body for send to route with deposits list: %w", err)
	}

	if strings.TrimSpace(input.Pagination.Cursor.BlockNumber) != emptyKey &&
		strings.TrimSpace(input.Pagination.Cursor.SortingValue) != emptyKey {
		jsBody, err = sjson.SetBytes(jsBody, "pagination.cursor", map[string]interface{}{
			"blockNumber":  input.Pagination.Cursor.BlockNumber,
			"sortingValue": input.Pagination.Cursor.SortingValue,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal pagination cursor of body for send to route with deposits list: %w", err)
		}
	}

	err = json.Unmarshal(jsBody, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body for send to route with deposits list: %w", err)
	}

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(body).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the deposit request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get deposits list: %s", resp.String())
	}

	response := new(GetDepositsListResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get deposits list: %s", resp.String())
	}

	return response.Data, nil
}

func GetDepositByHashIncomingWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	depositHash string,
	recipientAccount *intMaxAcc.PrivateKey,
) (json.RawMessage, error) {
	resp, err := getDepositByHashIncomingRawRequest(
		ctx,
		cfg,
		depositHash,
		recipientAccount,
	)
	if err != nil {
		return nil, err
	}

	var deposit *intMaxTypes.Deposit
	deposit, err = GetDepositFromBackupData(
		resp.Deposit,
		recipientAccount,
	)
	if err != nil {
		return nil, err
	}

	js, err := json.MarshalIndent(&GetDepositTxData{
		ID:          resp.Deposit.ID,
		Recipient:   resp.Deposit.Recipient,
		BlockNumber: resp.Deposit.BlockNumber,
		Deposit:     deposit,
		CreatedAt:   resp.Deposit.CreatedAt,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deposit by hash (incoming): %w", err)
	}

	return js, nil
}

func getDepositByHashIncomingRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	depositHash string,
	recipientAccount *intMaxAcc.PrivateKey,
) (*GetDepositByHashIncomingData, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf(
		"%s/v1/backups/deposit/%s",
		cfg.API.DataStoreVaultUrl,
		url.QueryEscape(depositHash),
	)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetQueryParams(map[string]string{
		"recipient": recipientAccount.ToAddress().String(),
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send of the deposit request: %w"
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

	response := new(GetDepositByHashIncomingResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get deposit by hash")
	}

	return response.Data, nil
}
