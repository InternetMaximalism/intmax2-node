package tx_transfer_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_proposed"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func GetBlockProposed(
	ctx context.Context,
	senderAccount *intMaxAcc.PrivateKey,
	transfersHash goldenposeidon.PoseidonHashOut,
	nonce uint64,
) (*BlockProposedResponseData, error) {
	tx, err := intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()
	fmt.Printf("transfersHash: %v\n", transfersHash.String())
	fmt.Printf("nonce: %v\n", nonce)
	fmt.Printf("tx hash: %v\n", tx.Hash())

	message := finite_field.BytesToFieldElementSlice(txHash.Marshal())
	signature, err := senderAccount.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	const duration = 60 * time.Minute
	expiration := time.Now().Add(duration)

	res, err := retryRequest(
		ctx, senderAccount.ToAddress(), *txHash, expiration, signature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposed block: %w", err)
	}

	return res, nil
}

const retryInterval = 10 * time.Second

func retryRequest(
	_ context.Context,
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	expiration time.Time,
	signature *bn254.G2Affine,
) (*BlockProposedResponseData, error) {
	// TODO: Implement context with timeout
	const numRetry = 3
	for i := 0; i < numRetry; i++ {
		response, err := GetBlockProposedRawRequest(
			senderAddress,
			txHash,
			expiration,
			signature,
		)
		if err == nil {
			return response, nil
		}

		fmt.Println("Cannot get successful response. Retry in", retryInterval)
		time.Sleep(retryInterval)
	}

	return nil, errors.New("failed to get proposed block")
}

type BlockProposedResponseData struct {
	TxTreeRoot        goldenposeidon.PoseidonHashOut    `json:"txTreeRoot"`
	TxTreeMerkleProof []*goldenposeidon.PoseidonHashOut `json:"txTreeMerkleProof"`
	PublicKeysHash    []byte                            `json:"publicKeysHash"`
}

type BlockProposedResponse struct {
	Success bool                           `json:"success"`
	Data    block_proposed.UCBlockProposed `json:"data"`
}

func GetBlockProposedRawRequest(
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	expiration time.Time,
	signature *bn254.G2Affine,
) (*BlockProposedResponseData, error) {
	return getBlockProposedRawRequest(
		senderAddress.String(), hexutil.Encode(txHash.Marshal()), expiration, hexutil.Encode(signature.Marshal()),
	)
}

func getBlockProposedRawRequest(
	senderAddress, txHash string,
	expiration time.Time,
	signature string,
) (*BlockProposedResponseData, error) {
	ucInput := block_proposed.UCBlockProposedInput{
		Sender:     senderAddress,
		TxHash:     txHash,
		Expiration: expiration,
		Signature:  signature,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	body := strings.NewReader(string(bd))

	const (
		apiBaseUrl  = "http://localhost"
		contentType = "application/json"
	)
	apiUrl := fmt.Sprintf("%s/v1/block/proposed", apiBaseUrl)

	resp, err := http.Post(apiUrl, contentType, body) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		return nil, fmt.Errorf("response body: %s", string(bodyBytes))
	}

	var res BlockProposedResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		return nil, fmt.Errorf("failed to get proposed block: %v", res)
	}

	txRoot := new(goldenposeidon.PoseidonHashOut)
	err = txRoot.FromString(res.Data.TxRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tx root: %w", err)
	}

	txTreeMerkleProof := make([]*goldenposeidon.PoseidonHashOut, len(res.Data.TxTreeMerkleProof))
	for i, proof := range res.Data.TxTreeMerkleProof {
		sibling := new(goldenposeidon.PoseidonHashOut)
		err = sibling.FromString(proof)
		if err != nil {
			return nil, fmt.Errorf("failed to decode tx tree merkle proof: %w", err)
		}
		txTreeMerkleProof[i] = sibling
	}

	publicKeysHash, err := hexutil.Decode(res.Data.PublicKeysHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public keys hash: %w", err)
	}

	return &BlockProposedResponseData{
		TxTreeRoot:        *txRoot,
		TxTreeMerkleProof: txTreeMerkleProof,
		PublicKeysHash:    publicKeysHash,
	}, nil
}
