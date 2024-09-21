package tx_transfer_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_status"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
)

const (
	retryBlockStatusInterval   = 10 * time.Second
	timeoutBlockStatusInterval = 3 * time.Minute
)

func GetBlockStatus(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	txHash goldenposeidon.PoseidonHashOut,
) (*block_status.UCBlockStatus, error) {
	res, err := retryBlockStatusRequest(
		ctx, cfg, log, txHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get block status: %w", err)
	}

	return res, nil
}

func retryBlockStatusRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	txHash goldenposeidon.PoseidonHashOut,
) (*block_status.UCBlockStatus, error) {
	ticker := time.NewTicker(retryBlockStatusInterval)

	gbpCtx, cancel := context.WithTimeout(ctx, timeoutBlockStatusInterval)
	defer cancel()

	fmt.Println("Waiting the block containing your tx...")
	for {
		select {
		case <-gbpCtx.Done():
			return nil, fmt.Errorf("failed to get block status")
		case <-ticker.C:
			response, err := GetBlockStatusRawRequest(
				ctx,
				cfg,
				log,
				txHash,
			)
			if err == nil {
				if !response.IsPosted {
					// The Block containing your tx is not posted yet
					continue
				}

				return response, nil
			}

			fmt.Printf("Error retryBlockStatusRequest: %v\n", err.Error())
			if errors.Is(err, ErrBlockNotFound) {
				// The Block containing your tx is not found
				continue
			}

			log.WithError(err).Errorf("Cannot get successful response. The searching tx tree root is 0x%x.\n", txHash.Marshal())

			return nil, err
		}
	}
}

func GetBlockStatusRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	txTreeRoot goldenposeidon.PoseidonHashOut,
) (*block_status.UCBlockStatus, error) {
	return getBlockStatusRawRequest(
		ctx, cfg, log, hexutil.Encode(txTreeRoot.Marshal()),
	)
}

func getBlockStatusRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	txTreeRoot string,
) (*block_status.UCBlockStatus, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/block/status/%s", cfg.API.BlockBuilderUrl, txTreeRoot)

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

	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrBlockNotFound
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
			log.WithFields(logger.Fields{
				"status_code": resp.StatusCode(),
				"response":    resp.String(),
			}).WithError(err).Errorf("Processing ended error occurred")
		}
	}()

	var res block_status.UCBlockStatus
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		return nil, err
	}

	return &res, nil
}
