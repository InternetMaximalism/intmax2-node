package balance_prover_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const timeoutForFetchingBalanceValidityProof = 3 * time.Second

var ErrBalanceProofNotGenerated = errors.New("balance proof is not generated")

type BalanceProofWithPublicInputs struct {
	Proof        string
	PublicInputs *BalancePublicInputs
}

type BalanceProcessor struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
}

func NewBalanceProcessor(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) *BalanceProcessor {
	return &BalanceProcessor{
		ctx,
		cfg,
		log,
	}
}

func (s *BalanceProcessor) ProveUpdate(
	publicKey *intMaxAcc.PublicKey,
	updateWitness *UpdateWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	fmt.Println("ProveUpdate")
	fmt.Printf("publicKey: %v\n", publicKey)
	requestID, err := s.requestUpdateBalanceValidityProof(publicKey, updateWitness, lastBalanceProof)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(timeoutForFetchingBalanceValidityProof)
	for {
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-ticker.C:
			proof, err := s.fetchUpdateBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				return nil, err
			}

			balanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
			if err != nil {
				return nil, err
			}

			return &BalanceProofWithPublicInputs{
				Proof:        *proof.Proof,
				PublicInputs: balancePublicInputs,
			}, nil
		}
	}
}

func (s *BalanceProcessor) ProveReceiveDeposit(
	publicKey *intMaxAcc.PublicKey,
	receiveDepositWitness *ReceiveDepositWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	fmt.Println("ProveReceiveDeposit")
	fmt.Printf("publicKey: %v\n", publicKey)
	db, err := json.Marshal(receiveDepositWitness)
	if err != nil {
		return nil, err
	}
	fmt.Printf("receiveDepositWitness: %s\n", db)
	fmt.Printf("lastBalanceProof: %v\n", lastBalanceProof)
	requestID, err := s.requestReceiveDepositBalanceValidityProof(publicKey, receiveDepositWitness, lastBalanceProof)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(timeoutForFetchingBalanceValidityProof)
	for {
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-ticker.C:
			proof, err := s.fetchReceiveDepositBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				fmt.Printf("err: %v\n", err)
				return nil, err
			}

			balanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
			if err != nil {
				return nil, err
			}

			return &BalanceProofWithPublicInputs{
				Proof:        *proof.Proof,
				PublicInputs: balancePublicInputs,
			}, nil
		}
	}
}

func (s *BalanceProcessor) ProveSend(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *UpdateWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	fmt.Printf("ProveSend")
	fmt.Printf("publicKey: %v", publicKey)
	requestID, err := s.requestSendBalanceValidityProof(publicKey, sendWitness, updateWitness, lastBalanceProof)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(timeoutForFetchingBalanceValidityProof)
	for {
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-ticker.C:
			proof, err := s.fetchSendBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				return nil, err
			}

			balanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
			if err != nil {
				return nil, err
			}

			return &BalanceProofWithPublicInputs{
				Proof:        *proof.Proof,
				PublicInputs: balancePublicInputs,
			}, nil
		}
	}
}

func (s *BalanceProcessor) ProveReceiveTransfer(
	publicKey *intMaxAcc.PublicKey,
	receiveTransferWitness *ReceiveTransferWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	fmt.Printf("ProveReceiveTransfer")
	fmt.Printf("publicKey: %v", publicKey)
	fmt.Printf("receiveTransferWitness: %v", receiveTransferWitness)
	fmt.Printf("lastBalanceProof: %v", lastBalanceProof)
	requestID, err := s.requestReceiveTransferBalanceValidityProof(publicKey, receiveTransferWitness, lastBalanceProof)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(timeoutForFetchingBalanceValidityProof)
	for {
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-ticker.C:
			proof, err := s.fetchReceiveTransferBalanceValidityProof(publicKey, requestID)

			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				return nil, err
			}

			balanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
			if err != nil {
				return nil, err
			}

			return &BalanceProofWithPublicInputs{
				Proof:        *proof.Proof,
				PublicInputs: balancePublicInputs,
			}, nil
		}
	}
}

type BalanceValidityResponse struct {
	Success bool   `json:"success"`
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}

type MerkleProofInput = []string

type IndexedMembershipProofInput struct {
	IsIncluded bool                    `json:"isIncluded"`
	LeafProof  IndexedMerkleProofInput `json:"leafProof"`
	LeafIndex  LeafIndexInput          `json:"leafIndex"`
	Leaf       IndexedMerkleLeafInput  `json:"leaf"`
}

type UpdateWitnessInput struct {
	ValidityProof          string                      `json:"validityProof"`
	BlockMerkleProof       MerkleProofInput            `json:"blockMerkleProof"`
	AccountMembershipProof IndexedMembershipProofInput `json:"accountMembershipProof"`
}

type UpdateBalanceValidityInput struct {
	UpdateWitness *UpdateWitnessInput `json:"balanceUpdateWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

func (input *UpdateWitnessInput) FromUpdateWitness(updateWitness *UpdateWitness) *UpdateWitnessInput {
	input.ValidityProof = updateWitness.ValidityProof
	input.BlockMerkleProof = make(MerkleProofInput, len(updateWitness.BlockMerkleProof.Siblings))
	for i := range updateWitness.BlockMerkleProof.Siblings {
		input.BlockMerkleProof[i] = updateWitness.BlockMerkleProof.Siblings[i].String()
	}

	input.AccountMembershipProof.IsIncluded = updateWitness.AccountMembershipProof.IsIncluded
	input.AccountMembershipProof.LeafProof = make(IndexedMerkleProofInput, len(updateWitness.AccountMembershipProof.LeafProof.Siblings))
	for i := range updateWitness.AccountMembershipProof.LeafProof.Siblings {
		input.AccountMembershipProof.LeafProof[i] = new(poseidonHashOut).Set(updateWitness.AccountMembershipProof.LeafProof.Siblings[i])
	}
	input.AccountMembershipProof.LeafIndex = updateWitness.AccountMembershipProof.LeafIndex
	input.AccountMembershipProof.Leaf = IndexedMerkleLeafInput{}
	fmt.Printf("updateWitness.AccountMembershipProof.Leaf: %v\n", updateWitness.AccountMembershipProof.Leaf)
	input.AccountMembershipProof.Leaf.FromIndexedMerkleLeaf(&updateWitness.AccountMembershipProof.Leaf)

	return input
}

type ReceiveDepositBalanceValidityInput struct {
	ReceiveDepositWitness *ReceiveDepositWitnessInput `json:"receiveDepositWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

type SendBalanceValidityInput struct {
	SendWitness   *SendWitnessInput   `json:"sendWitness"`
	UpdateWitness *UpdateWitnessInput `json:"balanceUpdateWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

type ReceiveTransferBalanceValidityInput struct {
	ReceiveTransferWitness *ReceiveTransferWitnessInput `json:"receiveTransferWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

// Execute the following request:
// curl -X POST -d '{ "sendWitness":'$(cat data/send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)',
// "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)',
// "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq
func (p *BalanceProcessor) requestUpdateBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	updateWitness *UpdateWitness,
	prevBalanceProof *string,
) (string, error) {
	requestBody := UpdateBalanceValidityInput{
		UpdateWitness:    new(UpdateWitnessInput).FromUpdateWitness(updateWitness),
		PrevBalanceProof: prevBalanceProof,
	}

	// bd2, _ := json.Marshal(requestBody.UpdateWitness)
	// fmt.Printf("requestBody: %s\n", bd2)

	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/update", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = " the balance proof request for SendWitnessfailed to send: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return "", fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for SendWitness: %s", response.Message)
	}

	validityProofWithPlonky2Proof, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(updateWitness.ValidityProof)
	if err != nil {
		return "", err
	}

	requestID := new(block_validity_prover.ValidityPublicInputs).FromPublicInputs(validityProofWithPlonky2Proof.PublicInputs).PublicState.BlockHash.Hex()

	return requestID, nil
}

// Execute the following request:
// curl -X POST -d '{ "receiveDepositWitness":'$(cat data/receive_deposit_witness_0.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_deposit_proof_0.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/deposit | jq
func (p *BalanceProcessor) requestReceiveDepositBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	receiveDepositWitness *ReceiveDepositWitness,
	prevBalanceProof *string,
) (string, error) {
	requestBody := ReceiveDepositBalanceValidityInput{
		ReceiveDepositWitness: new(ReceiveDepositWitnessInput).FromReceiveDepositWitness(receiveDepositWitness),
		PrevBalanceProof:      prevBalanceProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON request body: %w", err)
	}
	fmt.Printf("size of requestBody: %d bytes\n", len(bd))

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/deposit", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send the balance proof request for ReceiveDepositWitness: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return "", fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for ReceiveDepositWitness: %s", response.Message)
	}

	requestID := strconv.FormatUint(uint64(receiveDepositWitness.DepositWitness.DepositIndex), 10)

	return requestID, nil
}

// Execute the following request:
// curl -X POST -d '{ "balanceUpdateWitness":'$(cat data/balance_update_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)',
// "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq
func (p *BalanceProcessor) requestSendBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *UpdateWitness,
	prevBalanceProof *string,
) (string, error) {
	requestBody := SendBalanceValidityInput{
		SendWitness:      new(SendWitnessInput).FromSendWitness(sendWitness),
		UpdateWitness:    new(UpdateWitnessInput).FromUpdateWitness(updateWitness),
		PrevBalanceProof: prevBalanceProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON request body: %w", err)
	}
	fmt.Printf("requestSendBalanceValidityProof: %s\n", bd)

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/send", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send the balance proof request for SendWitness: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return "", fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for SendWitness: %s", response.Message)
	}

	requestID := sendWitness.PrevBalancePis.PublicState.BlockHash.Hex()

	return requestID, nil
}

// Execute the following request:
// curl -X POST -d '{ "receiveTransferWitness":'$(cat data/receive_transfer_witness_0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.json)',
// "prevBalanceProof":"'$(base64 --input data/prev_receive_transfer_proof_0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer | jq
func (p *BalanceProcessor) requestReceiveTransferBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	receiveTransferWitness *ReceiveTransferWitness,
	prevBalanceProof *string,
) (string, error) {
	requestBody := ReceiveTransferBalanceValidityInput{
		ReceiveTransferWitness: new(ReceiveTransferWitnessInput).FromReceiveTransferWitness(receiveTransferWitness),
		PrevBalanceProof:       prevBalanceProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/transfer", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send the balance proof request for ReceiveTransferWitness: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return "", fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for ReceiveTransferWitness: %s", response.Message)
	}

	requestID := receiveTransferWitness.PrivateWitness.PrevPrivateState.Commitment().String()

	return requestID, nil
}

type BalanceValidityProofResponse struct {
	Success      bool    `json:"success"`
	Proof        *string `json:"proof,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/update/{:blockHash} | jq
func (p *BalanceProcessor) fetchUpdateBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/update/%s", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

	r := resty.New().R()
	resp, err := r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send updateWitness balance proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get response")
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, "balance proof is not generated") {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf("ErrorMessage: %v\n", response.ErrorMessage)
		return nil, fmt.Errorf("failed to get updateWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get updateWitness balance proof response: %v", response)
	}

	return response, nil
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/deposit/{:depositIndex} | jq
func (p *BalanceProcessor) fetchReceiveDepositBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/deposit/%s", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

	r := resty.New().R()
	resp, err := r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send depositWitness balance proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get response")
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, "balance proof is not generated") {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf("ErrorMessage: %v\n", response.ErrorMessage)
		return nil, fmt.Errorf("failed to get depositWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get depositWitness balance proof response: %v", response)
	}

	return response, nil
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/send/{:blockHash} | jq
func (p *BalanceProcessor) fetchSendBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/send/%s", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

	r := resty.New().R()
	resp, err := r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send sendWitness balance proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get response")
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, "balance proof is not generated") {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf("ErrorMessage: %v\n", response.ErrorMessage)
		return nil, fmt.Errorf("failed to get sendWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get sendWitness balance proof response: %v", response)
	}

	return response, nil
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/send/{:blockHash} | jq
func (p *BalanceProcessor) fetchReceiveTransferBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/transfer/%s", p.cfg.BlockValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

	r := resty.New().R()
	resp, err := r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send transferWitness balance proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get response")
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, "balance proof is not generated") {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf("ErrorMessage: %v\n", response.ErrorMessage)
		return nil, fmt.Errorf("failed to get transferWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get transferWitness balance proof response: %v", response)
	}

	return response, nil
}
