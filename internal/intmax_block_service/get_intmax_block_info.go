package intmax_block_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

func GetINTMAXBlockInfoByHashWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	blockHash string,
) (json.RawMessage, error) {
	const (
		emptyKey  = ""
		indentKey = "  "
	)

	resp, err := getINTMAXBlockInfoByBlockHashRawRequest(
		ctx,
		cfg,
		blockHash,
	)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		var js []byte
		js, err = json.MarshalIndent(&GetINTMAXBlockInfoResponse{
			Error: &GetDataError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			},
		}, emptyKey, indentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal error of get INTMAX block info by hash: %w", err)
		}

		return js, nil
	}

	var js []byte
	js, err = json.MarshalIndent(&GetINTMAXBlockInfoResponse{
		Success: true,
		Data: &GetINTMAXBlockInfoData{
			BlockNumber:                 resp.Data.BlockNumber,
			BlockHash:                   resp.Data.BlockHash,
			Status:                      resp.Data.Status,
			ExecutedBlockHashOnScroll:   resp.Data.ExecutedBlockHashOnScroll,
			ExecutedBlockHashOnEthereum: resp.Data.ExecutedBlockHashOnEthereum,
		},
	}, emptyKey, indentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal INTMAX block info by hash: %w", err)
	}

	return js, nil
}

func GetINTMAXBlockInfoByNumberWithRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	blockNumber uint64,
) (json.RawMessage, error) {
	const (
		emptyKey  = ""
		indentKey = "  "
	)

	resp, err := getINTMAXBlockInfoByBlockNumberRawRequest(
		ctx,
		cfg,
		blockNumber,
	)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		var js []byte
		js, err = json.MarshalIndent(&GetINTMAXBlockInfoResponse{
			Error: &GetDataError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			},
		}, emptyKey, indentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal error of get INTMAX block info by number: %w", err)
		}

		return js, nil
	}

	var js []byte
	js, err = json.MarshalIndent(&GetINTMAXBlockInfoResponse{
		Success: true,
		Data: &GetINTMAXBlockInfoData{
			BlockNumber:                 resp.Data.BlockNumber,
			BlockHash:                   resp.Data.BlockHash,
			Status:                      resp.Data.Status,
			ExecutedBlockHashOnScroll:   resp.Data.ExecutedBlockHashOnScroll,
			ExecutedBlockHashOnEthereum: resp.Data.ExecutedBlockHashOnEthereum,
		},
	}, emptyKey, indentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal INTMAX block info by hash: %w", err)
	}

	return js, nil
}

func getINTMAXBlockInfoByBlockHashRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	blockHash string,
) (*GetINTMAXBlockInfoResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf(
		"%s/v1/block-hash/%s/status",
		cfg.API.BlockValidityProverUrl,
		url.QueryEscape(blockHash),
	)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get INTMAX block info by hash request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return &GetINTMAXBlockInfoResponse{
			Error: &GetDataError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetINTMAXBlockInfoResponse{
			Error: &GetDataError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetINTMAXBlockInfoResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func getINTMAXBlockInfoByBlockNumberRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	blockNumber uint64,
) (*GetINTMAXBlockInfoResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf(
		"%s/v1/block-number/%s/status",
		cfg.API.BlockValidityProverUrl,
		url.QueryEscape(fmt.Sprintf("%v", blockNumber)),
	)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get INTMAX block info by number request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return &GetINTMAXBlockInfoResponse{
			Error: &GetDataError{
				Code:    http.StatusInternalServerError,
				Message: msg,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		const messageKey = "message"
		return &GetINTMAXBlockInfoResponse{
			Error: &GetDataError{
				Code:    resp.StatusCode(),
				Message: strings.ToLower(gjson.GetBytes(resp.Body(), messageKey).String()),
			},
		}, nil
	}

	response := new(GetINTMAXBlockInfoResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}
