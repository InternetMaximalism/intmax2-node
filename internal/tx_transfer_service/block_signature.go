package tx_transfer_service

import (
	"encoding/json"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/use_cases/block_signature"
	"io"
	"net/http"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type BlockSignatureResponse struct {
	Success bool                             `json:"success"`
	Data    block_signature.UCBlockSignature `json:"data"`
}

func SendSignedProposedBlock(
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
	fmt.Printf("encodedPrevBalanceProof: %v", encodedPrevBalanceProof)

	return PostBlockSignatureRawRequest(
		senderAccount.ToAddress(), txTreeRoot, signature,
		prevBalanceProof, transferStepProof,
	)
}

func PostBlockSignatureRawRequest(
	senderAddress intMaxAcc.Address,
	txHash goldenposeidon.PoseidonHashOut,
	signature *bn254.G2Affine,
	prevBalanceProof block_signature.Plonky2Proof,
	transferStepProof block_signature.Plonky2Proof,
) error {
	return postBlockSignatureRawRequest(
		senderAddress.String(),
		hexutil.Encode(txHash.Marshal()),
		hexutil.Encode(signature.Marshal()),
		prevBalanceProof,
		transferStepProof,
	)
}

func postBlockSignatureRawRequest(
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

	body := strings.NewReader(string(bd))

	const (
		apiBaseUrl  = "http://localhost"
		contentType = "application/json"
	)
	apiUrl := fmt.Sprintf("%s/v1/block/signature", apiBaseUrl)

	resp, err := http.Post(apiUrl, contentType, body) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d", resp.StatusCode)
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		return fmt.Errorf("response body: %s", string(bodyBytes))
	}

	var res BlockSignatureResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("failed to get proposed block: %+v", res)
	}

	return nil
}
