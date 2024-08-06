package tx_transfer_service

import (
	"context"
	"encoding/json"
	"errors"
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
	timeoutBlockStatusInterval = 3 * time.Minute
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

	for {
		select {
		case <-gbpCtx.Done():
			return nil, fmt.Errorf("failed to get block status")
		case <-ticker.C:
			response, err := GetBlockStatusRawRequest(
				ctx,
				cfg,
				log,
				txTreeRoot,
			)
			if err == nil {
				if !response.IsPosted {
					log.Infof("The Block containing your tx is not posted yet. Retry in %f second(s). The searching tx hash is 0x%x.", retryBlockStatusInterval.Seconds(), txTreeRoot.Marshal())
					continue
				}

				return response, nil
			}

			fmt.Printf("err: %v\n", err.Error())
			if err.Error() == "block not found" {
				log.Infof("The Block containing your tx is not found. Retry in %f second(s). The searching tx hash is 0x%x.", retryBlockStatusInterval.Seconds(), txTreeRoot.Marshal())
				continue
			}

			log.WithError(err).Errorf("Cannot get successful response. The searching tx hash is 0x%x.\n", txTreeRoot.Marshal())

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
	// ucInput := block_status.UCBlockStatusInput{
	// 	TxTreeRoot: txTreeRoot,
	// }

	// bd, err := json.Marshal(ucInput)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	// }

	const (
		httpKey     = "http"
		httpsKey    = "https"
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
		respStr := resp.String()
		respJSON := ErrorResponse{}
		err = json.Unmarshal([]byte(respStr), &respJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if respJSON.Message != "" {
			return nil, errors.New(respJSON.Message)
		}

		return nil, fmt.Errorf("failed to get response")
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

	fmt.Printf("res: %v\n", res)

	return &res, nil
}
