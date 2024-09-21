package block_validity_prover

import (
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/internal/logger"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-resty/resty/v2"
)

type FetchBlockValidityProof struct {
	Success      bool    `json:"success"`
	Proof        *string `json:"proof"`
	ErrorMessage *string `json:"errorMessage"`
}

type FetchBlockValidityProofInput struct {
	BlockHash string `json:"blockHash"`
}

// Execute the following request:
// curl $BLOCK_VALIDITY_PROVER_URL/proof/{:blockHash} | jq
func (p *blockValidityProver) fetchBlockValidityProof(blockHash common.Hash) (string, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/proof/%s", p.cfg.BlockValidityProver.BlockValidityProverUrl, blockHash.String())

	r := resty.New().R()
	resp, err := r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get validity proof request: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return "", errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return "", err
	}

	response := new(FetchBlockValidityProof)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to get verify deposit confirmation response: %v", response)
	}

	if response.Proof == nil {
		return "", fmt.Errorf("failed to get verify deposit confirmation response: %v", response)
	}

	return *response.Proof, nil
}
