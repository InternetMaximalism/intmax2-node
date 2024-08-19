package block_validity_prover

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
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
	apiUrl := fmt.Sprintf("%s/proof/%s", p.cfg.BlockValidityProver.BlockValidityProverUrl, blockHash.String())

	resp, err := http.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var res FetchBlockValidityProof
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		if res.ErrorMessage == nil {
			return "", fmt.Errorf("failed to request API")
		}

		return "", fmt.Errorf("failed to request API: %s", *res.ErrorMessage)
	}

	if res.Proof == nil {
		return "", fmt.Errorf("proof is not found")
	}

	return *res.Proof, nil
}
