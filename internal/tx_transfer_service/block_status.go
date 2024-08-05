package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/use_cases/block_status"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
)

const (
	retryBlockStatusInterval   = 10 * time.Second
	timeoutBlockStatusInterval = 2 * time.Minute
)

func GetBlockStatus(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	txTreeRoot goldenposeidon.PoseidonHashOut,
) (*block_status.UCBlockStatus, error) {
	res, err := retryBlockStatusRequest(
		ctx, cfg, log, txTreeRoot,
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
	txTreeRoot goldenposeidon.PoseidonHashOut,
) (*block_status.UCBlockStatus, error) {
	ticker := time.NewTicker(retryBlockStatusInterval)

	gbpCtx, cancel := context.WithTimeout(ctx, timeoutBlockStatusInterval)
	defer cancel()

	var latestError error
	for {
		select {
		case <-gbpCtx.Done():
			const msg = "failed to get block status"
			return nil, fmt.Errorf(msg)
		case <-ticker.C:
			response, err := GetBlockStatusRawRequest(
				ctx,
				cfg,
				log,
				txTreeRoot,
			)
			if err == nil {
				return response, nil
			}

			const msg = "Cannot get successful response. Retry in %f second(s)"
			log.WithError(err).Errorf(msg, retryInterval.Seconds())
		}
	}

	return nil, 
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
	ucInput := block_status.UCBlockStatusInput{
		TxTreeRoot: txTreeRoot,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/block/status", cfg.API.BlockBuilderUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send ot the block proposed request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
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
