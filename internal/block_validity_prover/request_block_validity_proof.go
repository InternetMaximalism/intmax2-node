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

type ProveBlockValidity struct {
	Success bool   `json:"success"`
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}

type ProveBlockValidityInput struct {
	BlockHash            string                     `json:"blockHash"`
	ValidityWitness      *CompressedValidityWitness `json:"validityWitness"`
	PlainValidityWitness *ValidityWitness           `json:"plainValidityWitness"`

	// base64 encoded string
	PrevValidityProof *string `json:"prevValidityProof,omitempty"`
}

// Execute the following request:
// curl -X POST -d '{"blockHash":"0x01", "validityWitness":'$(cat data/validity_witness_1.json)', "prevValidityProof":null }'
// -H "Content-Type: application/json" $BLOCK_VALIDITY_PROVER_URL/proof | jq
func (p *blockValidityProver) requestBlockValidityProof(blockHash common.Hash, validityWitness *ValidityWitness, prevValidityProof *string) error {
	nextAccountID, err := p.blockBuilder.NextAccountID()
	if err != nil {
		return fmt.Errorf("failed to get next account ID: %w", err)
	}

	compressedValidityWitness, err := validityWitness.Compress(nextAccountID)
	if err != nil {
		return fmt.Errorf("failed to compress validity witness: %w", err)
	}

	nextValidityPis := validityWitness.ValidityPublicInputs()
	p.log.Debugf("nextValidityPis block_proof block number: %d\n", nextValidityPis.PublicState.BlockNumber)
	p.log.Debugf("nextValidityPis block_proof prev account tree root: %s\n", nextValidityPis.PublicState.PrevAccountTreeRoot.String())
	p.log.Debugf("nextValidityPis block_proof account tree root: %s\n", nextValidityPis.PublicState.AccountTreeRoot.String())

	requestBody := ProveBlockValidityInput{
		BlockHash:       blockHash.String(),
		ValidityWitness: compressedValidityWitness,
		// PlainValidityWitness: validityWitness,
		PrevValidityProof: prevValidityProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON request body: %w", err)
	}
	p.log.Debugf("size of requestBlockValidityProof: %d bytes\n", len(bd))

	// encodedValidityWitness, err := json.Marshal(validityWitness)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal JSON request body: %w", err)
	// }
	// p.log.Debugf("encodedValidityWitness: %s\n", encodedValidityWitness)

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/proof", p.cfg.BlockValidityProver.BlockValidityProverUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send block validity proof request: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return err
	}

	response := new(ProveBlockValidity)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Message)
	}

	return nil
}
