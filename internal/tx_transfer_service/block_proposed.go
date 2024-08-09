package tx_transfer_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
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
	log logger.Logger,
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
		ctx, cfg, log, senderAccount.ToAddress(), *txHash, expiration, signature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposed block (retry): %w", err)
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
	log logger.Logger,
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	expiration time.Time,
	signature *bn254.G2Affine,
) (*BlockProposedResponseData, error) {
	ticker := time.NewTicker(retryInterval)

	gbpCtx, cancel := context.WithTimeout(ctx, timeoutInterval)
	defer cancel()

	firstTime := true
	for {
		select {
		case <-gbpCtx.Done():
			const msg = "failed to get proposed block"
			return nil, fmt.Errorf(msg)
		case <-ticker.C:
			response, err := GetBlockProposedRawRequest(
				ctx,
				cfg,
				log,
				senderAddress,
				txHash,
				expiration,
				signature,
			)
			if err == nil {
				return response, nil
			}

			const ErrTxTreeNotBuild = "txHash: the tx tree not build."
			if err.Error() == ErrTxTreeNotBuild {
				if firstTime {
					log.Infof("The Block Builder is currently processing the tx tree...")
					firstTime = false
				}
				continue
			}

			var ErrFailedResponse = errors.New("cannot get successful response")
			return nil, errors.Join(ErrFailedResponse, err)
		}
	}
}

type BlockProposedResponseData struct {
	TxTreeRoot        goldenposeidon.PoseidonHashOut    `json:"txTreeRoot"`
	TxTreeMerkleProof []*goldenposeidon.PoseidonHashOut `json:"txTreeMerkleProof"`
	PublicKeys        []*intMaxAcc.PublicKey            `json:"publicKeys"`
}

type BlockProposedResponse struct {
	Success bool                           `json:"success"`
	Data    block_proposed.UCBlockProposed `json:"data"`
}

func GetBlockProposedRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	expiration time.Time,
	signature *bn254.G2Affine,
) (*BlockProposedResponseData, error) {
	return getBlockProposedRawRequest(
		ctx, cfg, log, senderAddress.String(), hexutil.Encode(txHash.Marshal()), expiration, hexutil.Encode(signature.Marshal()),
	)
}

func getBlockProposedRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
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

	apiUrl := fmt.Sprintf("%s/v1/block/proposed", cfg.API.BlockBuilderUrl)

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
		respJSON := intMaxTypes.ErrorResponse{}
		err = json.Unmarshal([]byte(resp.String()), &respJSON)
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

	var res BlockProposedResponse
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		return nil, err
	}

	if !res.Success {
		err = fmt.Errorf("failed to get proposed block: %v", res)
		return nil, err
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
			err = fmt.Errorf("failed to decode tx tree merkle proof: %w", err)
			return nil, err
		}
		txTreeMerkleProof[i] = sibling
	}

	publicKeys := make([]*intMaxAcc.PublicKey, len(res.Data.PublicKeys))
	for i, address := range res.Data.PublicKeys {
		publicKeys[i], err = intMaxAcc.NewPublicKeyFromAddressHex(address)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public keys hash: %w", err)
		}
	}

	return &BlockProposedResponseData{
		TxTreeRoot:        *txRoot,
		TxTreeMerkleProof: txTreeMerkleProof,
		PublicKeys:        publicKeys,
	}, nil
}
