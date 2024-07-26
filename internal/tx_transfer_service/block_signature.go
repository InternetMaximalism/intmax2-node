package tx_transfer_service

import (
	"encoding/json"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/pb/gen/service/node"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_signature"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
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
		return fmt.Errorf("failed to get proposed block")
	}

	return nil
}
