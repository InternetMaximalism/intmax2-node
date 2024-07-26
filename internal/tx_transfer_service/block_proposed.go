package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_proposed"
	"net/http"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const signTimeout = 60 * time.Minute

func GetBlockProposed(
	ctx context.Context,
	cfg *configs.Config,
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

	expiration := time.Now().Add(signTimeout)
	var message []ffg.Element
	message, err = block_proposed.MakeMessage(txHash.String(), senderAccount.ToAddress().String(), expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to make message: %w", err)
	}

	signature, err := senderAccount.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	res, err := retryRequest(
		ctx, cfg, senderAccount.ToAddress(), *txHash, expiration, signature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposed block: %w", err)
	}

	return res, nil
}

const (
	retryInterval   = 1 * time.Second
	timeoutInterval = 120 * time.Second
)

func retryRequest(
	ctx context.Context,
	cfg *configs.Config,
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	expiration time.Time,
	signature *bn254.G2Affine,
) (*BlockProposedResponseData, error) {
	ticker := time.NewTicker(retryInterval)

	gbpCtx, cancel := context.WithTimeout(ctx, timeoutInterval)
	defer cancel()

	for {
		select {
		case <-gbpCtx.Done():
			const msg = "failed to get proposed block"
			return nil, fmt.Errorf(msg)
		case <-ticker.C:
			response, err := GetBlockProposedRawRequest(
				ctx,
				cfg,
				senderAddress,
				txHash,
				expiration,
				signature,
			)
			if err == nil {
				return response, nil
			}

			const msg = "Cannot get successful response (err = %q). Retry in %f second(s)"
			fmt.Println(fmt.Sprintf(msg, err, retryInterval.Seconds()))
		}
	}
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
	ctx context.Context,
	cfg *configs.Config,
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	expiration time.Time,
	signature *bn254.G2Affine,
) (*BlockProposedResponseData, error) {
	return getBlockProposedRawRequest(
		ctx, cfg, senderAddress.String(), hexutil.Encode(txHash.Marshal()), expiration, hexutil.Encode(signature.Marshal()),
	)
}

func getBlockProposedRawRequest(
	ctx context.Context,
	cfg *configs.Config,
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

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	schema := httpKey
	if cfg.HTTP.TLSUse {
		schema = httpsKey
	}

	apiUrl := fmt.Sprintf("%s://%s/v1/block/proposed", schema, cfg.HTTP.Addr())

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send transaction request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		fmt.Printf("Unexpected status code: %d (body: %q)\n", resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("response body: %s", resp.String())
	}

	var res BlockProposedResponse
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
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
