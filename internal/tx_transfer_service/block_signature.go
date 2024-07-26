package tx_transfer_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gen/service/node"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_signature"
	"math/big"
	"net/http"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-resty/resty/v2"
)

type BlockSignatureResponse struct {
	Success bool                        `json:"success"`
	Data    node.BlockSignatureResponse `json:"data"`
}

func MakeSampleTransferTree() (goldenposeidon.PoseidonHashOut, error) {
	const int10Key = 10

	var tokenIndex uint32 = 0
	amount := big.NewInt(int10Key)

	// Send transfer transaction
	recipient, err := intMaxAcc.NewPublicKeyFromAddressHex("0x06a7b64af8f414bcbeef455b1da5208c9b592b83ee6599824caa6d2ee9141a76")
	if err != nil {
		return goldenposeidon.PoseidonHashOut{}, fmt.Errorf("failed to parse recipient address: %v", err)
	}

	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipient.ToAddress().Bytes())
	if err != nil {
		return goldenposeidon.PoseidonHashOut{}, fmt.Errorf("failed to create recipient address: %v", err)
	}

	transfer := intMaxTypes.NewTransferWithRandomSalt(
		recipientAddress,
		tokenIndex,
		amount,
	)

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	initialLeaves := make([]*intMaxTypes.Transfer, 1)
	initialLeaves[0] = transfer

	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, initialLeaves, zeroTransfer.Hash())
	if err != nil {
		return goldenposeidon.PoseidonHashOut{}, fmt.Errorf("failed to create transfer tree: %v", err)
	}

	transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	return transferTreeRoot, nil
}

func MakeSampleTxTree(transferTreeRoot *goldenposeidon.PoseidonHashOut, nonce uint64) (goldenposeidon.PoseidonHashOut, error) {
	tx, err := intMaxTypes.NewTx(
		transferTreeRoot,
		nonce,
	)
	if err != nil {
		return goldenposeidon.PoseidonHashOut{}, fmt.Errorf("failed to create tx: %v", err)
	}

	zeroTx := new(intMaxTypes.Tx).SetZero()
	initialLeaves := make([]*intMaxTypes.Tx, 1)
	initialLeaves[0] = tx

	txTree, err := intMaxTree.NewTxTree(intMaxTree.TX_TREE_HEIGHT, initialLeaves, zeroTx.Hash())
	if err != nil {
		return goldenposeidon.PoseidonHashOut{}, fmt.Errorf("failed to create transfer tree: %v", err)
	}

	txTreeRoot, _, _ := txTree.GetCurrentRootCountAndSiblings()

	return txTreeRoot, nil
}

func MakePostBlockSignatureRawRequest(
	senderAccount *intMaxAcc.PrivateKey,
	txTreeRoot goldenposeidon.PoseidonHashOut,
	publicKeysHash []byte,
) (*block_signature.UCBlockSignatureInput, error) {
	message := finite_field.BytesToFieldElementSlice(txTreeRoot.Marshal())
	signature, err := senderAccount.WeightByHash(publicKeysHash).Sign(message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	prevBalanceProof := block_signature.Plonky2Proof{
		Proof:        []byte{},
		PublicInputs: []uint64{0, 0, 0, 0},
	} // TODO: This is dummy
	transferStepProof := block_signature.Plonky2Proof{
		Proof:        []byte{},
		PublicInputs: []uint64{1, 1, 1, 1},
	} // TODO: This is dummy
	encodedPrevBalanceProof, err := json.Marshal(prevBalanceProof)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prevBalanceProof: %w", err)
	}
	fmt.Printf("encodedPrevBalanceProof: %v", encodedPrevBalanceProof)

	return &block_signature.UCBlockSignatureInput{
		Sender:    senderAccount.ToAddress().String(),
		TxHash:    hexutil.Encode(txTreeRoot.Marshal()),
		Signature: hexutil.Encode(signature.Marshal()),
		EnoughBalanceProof: new(block_signature.EnoughBalanceProofInput).Set(&block_signature.EnoughBalanceProofInput{
			PrevBalanceProof:  &prevBalanceProof,
			TransferStepProof: &transferStepProof,
		}),
	}, nil
}

func SendSignedProposedBlock(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAccount *intMaxAcc.PrivateKey,
	txTreeRoot goldenposeidon.PoseidonHashOut,
	publicKeysHash []byte,
	// prevBalanceProof block_signature.Plonky2Proof,
	// transferStepProof block_signature.Plonky2Proof,
) error {
	message := finite_field.BytesToFieldElementSlice(txTreeRoot.Marshal())
	signature, err := senderAccount.WeightByHash(publicKeysHash).Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	prevBalanceProof := block_signature.Plonky2Proof{
		Proof:        []byte{},
		PublicInputs: []uint64{0, 0, 0, 0},
	} // TODO: This is dummy
	transferStepProof := block_signature.Plonky2Proof{
		Proof:        []byte{},
		PublicInputs: []uint64{1, 1, 1, 1},
	} // TODO: This is dummy
	encodedPrevBalanceProof, err := json.Marshal(prevBalanceProof)
	if err != nil {
		return fmt.Errorf("failed to marshal prevBalanceProof: %w", err)
	}
	log.Printf("encodedPrevBalanceProof: %v", encodedPrevBalanceProof)

	return PostBlockSignatureRawRequest(
		ctx, cfg, log,
		senderAccount.ToAddress(), txTreeRoot, signature,
		prevBalanceProof, transferStepProof,
	)
}

func PostBlockSignatureRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	signature *bn254.G2Affine,
	prevBalanceProof block_signature.Plonky2Proof,
	transferStepProof block_signature.Plonky2Proof,
) error {
	return postBlockSignatureRawRequest(
		ctx, cfg, log,
		senderAddress.String(),
		hexutil.Encode(txHash.Marshal()),
		hexutil.Encode(signature.Marshal()),
		prevBalanceProof,
		transferStepProof,
	)
}

func postBlockSignatureRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress, txHash, signature string,
	prevBalanceProof block_signature.Plonky2Proof,
	transferStepProof block_signature.Plonky2Proof,
) error {
	ucInput := block_signature.UCBlockSignatureInput{
		Sender:    senderAddress,
		TxHash:    txHash,
		Signature: signature,
		EnoughBalanceProof: new(block_signature.EnoughBalanceProofInput).Set(&block_signature.EnoughBalanceProofInput{
			PrevBalanceProof:  &prevBalanceProof,
			TransferStepProof: &transferStepProof,
		}),
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
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

	apiUrl := fmt.Sprintf("%s://%s/v1/block/signature", schema, cfg.HTTP.Addr())

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the block signature request: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return err
	}

	var res BlockSignatureResponse
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("failed to get proposed block")
	}

	return nil
}
