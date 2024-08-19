package block_validity_prover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

type ProveBlockValidity struct {
	Success bool   `json:"success"`
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}

type ProveBlockValidityInput struct {
	BlockHash       string                     `json:"blockHash"`
	ValidityWitness *CompressedValidityWitness `json:"validityWitness"`

	// base64 encoded string
	PrevValidityProof *string `json:"prevValidityProof"`
}

// Execute the following request:
// curl -X POST -d '{"blockHash":"0x01", "validityWitness":'$(cat data/validity_witness_1.json)', "prevValidityProof":null }'
// -H "Content-Type: application/json" $BLOCK_VALIDITY_PROVER_URL/proof | jq
func (p *blockValidityProver) requestBlockValidityProof(blockHash common.Hash, validityWitness *ValidityWitness, prevValidityProof *string) error {
	apiUrl := fmt.Sprintf("%s/proof", p.cfg.BlockValidityProver.BlockValidityProverUrl)

	maxUsedAccountID := p.blockBuilder.AccountTree.Count()
	compressedValidityWitness, err := validityWitness.Compress(maxUsedAccountID)
	if err != nil {
		return fmt.Errorf("failed to compress validity witness: %w", err)
	}

	requestBody := ProveBlockValidityInput{
		BlockHash:         blockHash.String(),
		ValidityWitness:   compressedValidityWitness,
		PrevValidityProof: prevValidityProof,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var res ProveBlockValidity
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("failed to request API: %s", res.Message)
	}

	p.log.Debugf("Prove block validity request success: %s", res.Message)

	return nil
}
