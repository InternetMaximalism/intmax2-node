package tx_transfer_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_info"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type BlockInfoResponse struct {
	Success bool                   `json:"success"`
	Data    block_info.UCBlockInfo `json:"data"`
}

type BlockInfoResponseData struct {
	ScrollAddress string            `json:"scrollAddress"`
	IntMaxAddress string            `json:"intMaxAddress"`
	TransferFee   map[string]string `json:"transferFee"`
	Difficulty    int64             `json:"difficulty"`
}

func GetBlockInfo(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) (*BlockInfoResponseData, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/info", cfg.API.BlockBuilderUrl)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send ot the block proposed request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		respJSON := intMaxTypes.ErrorResponse{}
		err = json.Unmarshal([]byte(resp.String()), &respJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if respJSON.Message != "" {
			return nil, errors.New(respJSON.Message)
		}

		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")

		return nil, err
	}

	defer func() {
		if err != nil {
			fmt.Println("Processing ended error occurred")
		}
	}()

	var res BlockInfoResponse
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		return nil, err
	}

	if !res.Success {
		err = fmt.Errorf("failed to get block info: %v", res)
		return nil, err
	}

	return &BlockInfoResponseData{
		ScrollAddress: res.Data.ScrollAddress,
		IntMaxAddress: res.Data.IntMaxAddress,
		TransferFee:   res.Data.TransferFee,
		Difficulty:    res.Data.Difficulty,
	}, nil
}
