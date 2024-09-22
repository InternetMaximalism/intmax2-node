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
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const (
	timeoutForFetchingBalanceValidityProof = 3 * time.Second

	messageErrorMessage                   = "ErrorMessage: %v\n"
	messageBalanceProofIsNotGenerated     = "balance proof is not generated"
	msgFailedToGetResponse                = "failed to get response"
	messageFailedToUnmarshalResponse      = "failed to unmarshal response: %w"
	messageFailedToMarshalJSONRequestBody = "failed to marshal JSON request body: %w"
	unexpectedStatusCode                  = "Unexpected status code"
)

type SenderProofWithPublicInputs struct {
	Proof        string
	PublicInputs *SenderPublicInputs
}

type BalanceProofWithPublicInputs struct {
	// NOTICE: include public inputs
	Proof        string
	PublicInputs *BalancePublicInputs
}

type balanceProcessor struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
}

type BalanceProcessor interface {
	ProveReceiveDeposit(publicKey *intMaxAcc.PublicKey, receiveDepositWitness *ReceiveDepositWitness, lastBalanceProof *string) (*BalanceProofWithPublicInputs, error)
	ProveReceiveTransfer(publicKey *intMaxAcc.PublicKey, receiveTransferWitness *ReceiveTransferWitness, lastBalanceProof *string) (*BalanceProofWithPublicInputs, error)
	ProveSendTransition(publicKey *intMaxAcc.PublicKey, sendWitness *SendWitness, updateWitness *block_validity_prover.UpdateWitness, lastBalanceProof *string) (*SenderProofWithPublicInputs, error)
	ProveSend(publicKey *intMaxAcc.PublicKey, sendWitness *SendWitness, updateWitness *block_validity_prover.UpdateWitness, lastBalanceProof *string) (*BalanceProofWithPublicInputs, error)
	ProveUpdate(publicKey *intMaxAcc.PublicKey, updateWitness *block_validity_prover.UpdateWitness, lastBalanceProof *string) (*BalanceProofWithPublicInputs, error)
}

func NewBalanceProcessor(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) BalanceProcessor {
	return &balanceProcessor{
		ctx,
		cfg,
		log,
	}
}

func (s *balanceProcessor) ProveUpdate(
	publicKey *intMaxAcc.PublicKey,
	updateWitness *block_validity_prover.UpdateWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	s.log.Debugf("ProveUpdate\n")
	s.log.Debugf("publicKey: %v\n", publicKey)
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
			var proof *BalanceValidityProofResponse
			proof, err = s.fetchUpdateBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) ||
					errors.Is(err, ErrStatusRequestTimeout) {
					continue
				}

				return nil, err
			}

			var balanceProofWithPis *intMaxTypes.Plonky2Proof
			balanceProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			var balancePublicInputs *BalancePublicInputs
			balancePublicInputs, err = new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
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

func (s *balanceProcessor) ProveReceiveDeposit(
	publicKey *intMaxAcc.PublicKey,
	receiveDepositWitness *ReceiveDepositWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	s.log.Debugf("ProveReceiveDeposit\n")
	s.log.Debugf("publicKey: %v\n", publicKey)

	if lastBalanceProof != nil {
		lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*lastBalanceProof)
		if err != nil {
			return nil, err
		}

		lastBalancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
		if err != nil {
			return nil, err
		}
		encodedLastBalancePublicInputs, err := json.Marshal(lastBalancePublicInputs)
		if err != nil {
			return nil, err
		}
		fmt.Printf("encodedLastBalancePublicInputs: %s\n", encodedLastBalancePublicInputs)

		lastBalanceProofPrivateCommitment := lastBalancePublicInputs.PrivateCommitment
		receiveDepositWitnessPrivateCommitment := receiveDepositWitness.PrivateWitness.PrevPrivateState.Commitment()
		fmt.Printf("last balance proof commitment: %s\n", lastBalanceProofPrivateCommitment.String())
		fmt.Printf("receive deposit commitment: %s\n", receiveDepositWitnessPrivateCommitment.String())
		fmt.Printf("receive deposit private state: %v\n", receiveDepositWitness.PrivateWitness.PrevPrivateState)
		if !receiveDepositWitnessPrivateCommitment.Equal(lastBalanceProofPrivateCommitment) {
			return nil, fmt.Errorf("last balance proof commitment is not equal to receive deposit commitment")
		}
	} else {
		fmt.Println("private state should be equal to default private state")
		lastBalanceProofPrivateCommitment := new(PrivateState).SetDefault().Commitment()
		receiveDepositWitnessPrivateCommitment := receiveDepositWitness.PrivateWitness.PrevPrivateState.Commitment()
		fmt.Printf("last balance proof commitment: %s\n", lastBalanceProofPrivateCommitment.String())
		fmt.Printf("receive deposit commitment: %s\n", receiveDepositWitnessPrivateCommitment.String())
		if !receiveDepositWitnessPrivateCommitment.Equal(lastBalanceProofPrivateCommitment) {
			return nil, fmt.Errorf("last balance proof commitment is not equal to receive deposit commitment")
		}
	}

	fmt.Printf("default PrivateState: %v\n", new(PrivateState).SetDefault().Commitment())

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
			var proof *BalanceValidityProofResponse
			proof, err := s.fetchReceiveDepositBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) || errors.Is(err, ErrStatusRequestTimeout) {
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

func (s *balanceProcessor) ProveSendTransition(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *block_validity_prover.UpdateWitness,
	lastBalanceProof *string,
) (*SenderProofWithPublicInputs, error) {
	s.log.Debugf("ProveSend\n")
	s.log.Debugf("publicKey: %v\n", publicKey)
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
			var proof *BalanceValidityProofResponse
			proof, err = s.fetchSendBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				return nil, err
			}

			senderProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			senderPublicInputs, err := new(SenderPublicInputs).FromPublicInputs(senderProofWithPis.PublicInputs)
			if err != nil {
				return nil, err
			}

			return &SenderProofWithPublicInputs{
				Proof:        *proof.Proof,
				PublicInputs: senderPublicInputs,
			}, nil
		}
	}
}

func (s *balanceProcessor) ProveSend(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *block_validity_prover.UpdateWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	s.log.Debugf("ProveSend\n")
	s.log.Debugf("publicKey: %v\n", publicKey)
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
			var proof *BalanceValidityProofResponse
			proof, err = s.fetchSendBalanceValidityProof(publicKey, requestID)
			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				return nil, err
			}

			var balanceProofWithPis *intMaxTypes.Plonky2Proof
			balanceProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			var balancePublicInputs *BalancePublicInputs
			balancePublicInputs, err = new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
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

func (s *balanceProcessor) ProveReceiveTransfer(
	publicKey *intMaxAcc.PublicKey,
	receiveTransferWitness *ReceiveTransferWitness,
	lastBalanceProof *string,
) (*BalanceProofWithPublicInputs, error) {
	// request balance prover
	s.log.Debugf("ProveReceiveTransfer\n")
	s.log.Debugf("publicKey: %v\n", publicKey)
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
			var proof *BalanceValidityProofResponse
			proof, err = s.fetchReceiveTransferBalanceValidityProof(publicKey, requestID)

			if err != nil {
				if errors.Is(err, ErrBalanceProofNotGenerated) {
					continue
				}

				return nil, err
			}

			var balanceProofWithPis *intMaxTypes.Plonky2Proof
			balanceProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*proof.Proof)
			if err != nil {
				return nil, err
			}

			var balancePublicInputs *BalancePublicInputs
			balancePublicInputs, err = new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
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
	RequestID     string              `json:"requestId"`
	UpdateWitness *UpdateWitnessInput `json:"balanceUpdateWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

func (input *UpdateWitnessInput) FromUpdateWitness(updateWitness *block_validity_prover.UpdateWitness) *UpdateWitnessInput {
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
	input.AccountMembershipProof.Leaf.FromIndexedMerkleLeaf(&updateWitness.AccountMembershipProof.Leaf)
	// fmt.Printf("updateWitness.AccountMembershipProof.Leaf: %v\n", updateWitness.AccountMembershipProof)
	// fmt.Printf("input.AccountMembershipProof.Leaf: %v\n", input.AccountMembershipProof)

	return input
}

type ReceiveDepositBalanceValidityInput struct {
	RequestID             string                      `json:"requestId"`
	ReceiveDepositWitness *ReceiveDepositWitnessInput `json:"receiveDepositWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

type SendBalanceValidityInput struct {
	RequestID     string              `json:"requestId"`
	SendWitness   *SendWitnessInput   `json:"sendWitness"`
	UpdateWitness *UpdateWitnessInput `json:"balanceUpdateWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

type SpendProofInput struct {
	RequestID     string              `json:"requestId"`
	SendWitness   *SendWitnessInput   `json:"sendWitness"`
	UpdateWitness *UpdateWitnessInput `json:"balanceUpdateWitness"`
}

type ReceiveTransferBalanceValidityInput struct {
	RequestID              string                       `json:"requestId"`
	ReceiveTransferWitness *ReceiveTransferWitnessInput `json:"receiveTransferWitness"`

	// base64 encoded string
	PrevBalanceProof *string `json:"prevBalanceProof,omitempty"`
}

// Execute the following request:
// curl -X POST -d '{ "requestId": "1", "sendWitness":'$(cat data/send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)',
// "balanceUpdateWitness":'$(cat data/balance_update_for_send_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)',
// "prevBalanceProof":"'$(base64 --input data/prev_balance_update_for_send_proof_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/send | jq
func (p *balanceProcessor) requestUpdateBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	updateWitness *block_validity_prover.UpdateWitness,
	prevBalanceProof *string,
) (string, error) {
	requestID := uuid.New().String()
	requestBody := UpdateBalanceValidityInput{
		RequestID:        requestID,
		UpdateWitness:    new(UpdateWitnessInput).FromUpdateWitness(updateWitness),
		PrevBalanceProof: prevBalanceProof,
	}
	// bd2, _ := json.Marshal(requestBody.UpdateWitness)
	// fmt.Printf("requestBody: %s\n", bd2)

	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(messageFailedToMarshalJSONRequestBody, err)
	}
	p.log.Debugf("size of requestUpdateBalanceValidityProof: %d bytes\n", len(bd))

	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/update", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = " the balance proof request for UpdateWitnessfailed to send: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "update request error occurred"
		return "", errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response: status code %d, response: %v", resp.StatusCode(), resp.String())
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for UpdateWitness: %s", response.Message)
	}

	return requestID, nil
}

// Execute the following request:
// curl -X POST -d '{ "requestId": "1", "receiveDepositWitness":'$(cat data/receive_deposit_witness_0.json)', "prevBalanceProof":"'$(base64 --input data/prev_receive_deposit_proof_0.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/deposit | jq
func (p *balanceProcessor) requestReceiveDepositBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	receiveDepositWitness *ReceiveDepositWitness,
	prevBalanceProof *string,
) (string, error) {
	requestID := uuid.New().String()
	requestBody := ReceiveDepositBalanceValidityInput{
		RequestID:             requestID,
		ReceiveDepositWitness: new(ReceiveDepositWitnessInput).FromReceiveDepositWitness(receiveDepositWitness),
		PrevBalanceProof:      prevBalanceProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(messageFailedToMarshalJSONRequestBody, err)
	}
	p.log.Debugf("size of requestReceiveDepositBalanceValidityProof: %d bytes\n", len(bd))

	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/deposit", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress)

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
		const msg = "receive deposit request error occurred"
		return "", errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for ReceiveDepositWitness: %s", response.Message)
	}

	return requestID, nil
}

// Execute the following request:
// curl -X POST -d '{ "requestId": "1", "balanceUpdateWitness":'$(cat data/balance_update_witness_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.json)',
// "prevBalanceProof":"'$(base64 --input data/prev_balance_update_proof_0xb183d250d266cb05408a4c37d7b3bb20474a439336ac09a892cc29e08f2eba8c.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/update | jq
func (p *balanceProcessor) requestSendBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *block_validity_prover.UpdateWitness,
	prevBalanceProof *string,
) (string, error) {
	requestID := uuid.New().String()
	requestBody := SendBalanceValidityInput{
		RequestID:        requestID,
		SendWitness:      new(SendWitnessInput).FromSendWitness(sendWitness),
		UpdateWitness:    new(UpdateWitnessInput).FromUpdateWitness(updateWitness),
		PrevBalanceProof: prevBalanceProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(messageFailedToMarshalJSONRequestBody, err)
	}
	p.log.Debugf("size of requestSendBalanceValidityProof: %d bytes\n", len(bd))

	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/send", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress)

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
		return "", errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for SendWitness: %s", response.Message)
	}

	return requestID, nil
}

func (p *balanceProcessor) requestSpendProof(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *block_validity_prover.UpdateWitness,
) (string, error) {
	requestID := uuid.New().String()
	requestBody := SpendProofInput{
		RequestID:     requestID,
		SendWitness:   new(SendWitnessInput).FromSendWitness(sendWitness),
		UpdateWitness: new(UpdateWitnessInput).FromUpdateWitness(updateWitness),
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(messageFailedToMarshalJSONRequestBody, err)
	}
	p.log.Debugf("size of requestSpendProof: %d bytes\n", len(bd))

	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/spend", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send the balance proof request for SpendWitness: %w"
		return "", fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "spend request error occurred"
		return "", errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for SpendWitness: %s", response.Message)
	}

	return requestID, nil
}

// Execute the following request:
// curl -X POST -d '{ "requestId": "1", "receiveTransferWitness":'$(cat data/receive_transfer_witness_0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.json)',
// "prevBalanceProof":"'$(base64 --input data/prev_receive_transfer_proof_0x7a00b7dbf1994ff9fb05a5897b7dc459dd9167ee7a4ad049b9850cbaf286bbee.bin)'" }'
// -H "Content-Type: application/json" $API_BALANCE_VALIDITY_PROVER_URL/proof/0x17600a0095835a6637a9532fd68d19b5b2e9c5907de541617a95c198b8fe7c37/transfer | jq
func (p *balanceProcessor) requestReceiveTransferBalanceValidityProof(
	publicKey *intMaxAcc.PublicKey,
	receiveTransferWitness *ReceiveTransferWitness,
	prevBalanceProof *string,
) (string, error) {
	requestID := uuid.New().String()
	requestBody := ReceiveTransferBalanceValidityInput{
		RequestID:              requestID,
		ReceiveTransferWitness: new(ReceiveTransferWitnessInput).FromReceiveTransferWitness(receiveTransferWitness),
		PrevBalanceProof:       prevBalanceProof,
	}
	bd, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(messageFailedToMarshalJSONRequestBody, err)
	}
	p.log.Debugf("size of requestReceiveTransferBalanceValidityProof: %d bytes\n", len(bd))

	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/transfer", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress)

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
		const msg = "receive transfer request error occurred"
		return "", errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return "", err
	}

	response := new(BalanceValidityResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return "", fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to send the balance proof request for ReceiveTransferWitness: %s", response.Message)
	}

	return requestID, nil
}

type BalanceValidityProofResponse struct {
	Success      bool    `json:"success"`
	RequestID    string  `json:"requestId"`
	Proof        *string `json:"proof,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/update/{:blockHash} | jq
func (p *balanceProcessor) fetchUpdateBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/update/%s", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

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
		return nil, errors.New(msg)
	}

	if resp.StatusCode() == http.StatusRequestTimeout {
		return nil, ErrStatusRequestTimeout
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, messageBalanceProofIsNotGenerated) {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf(messageErrorMessage, response.ErrorMessage)
		return nil, fmt.Errorf("failed to get updateWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get updateWitness balance proof response: %v", response)
	}

	return response, nil
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/deposit/{:depositIndex} | jq
func (p *balanceProcessor) fetchReceiveDepositBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/deposit/%s", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

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
		return nil, errors.New(msg)
	}

	if resp.StatusCode() == http.StatusRequestTimeout {
		return nil, ErrStatusRequestTimeout
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, messageBalanceProofIsNotGenerated) {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf(messageErrorMessage, response.ErrorMessage)
		return nil, fmt.Errorf("failed to get depositWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get depositWitness balance proof response: %v", response)
	}

	return response, nil
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/send/{:blockHash} | jq
func (p *balanceProcessor) fetchSendBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/send/%s", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

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
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, messageBalanceProofIsNotGenerated) {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf(messageErrorMessage, response.ErrorMessage)
		return nil, fmt.Errorf("failed to get sendWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get sendWitness balance proof response: %v", response)
	}

	return response, nil
}

func (p *balanceProcessor) fetchSpendBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/spend/%s", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

	r := resty.New().R()
	resp, err := r.SetContext(p.ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send spendWitness balance proof request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, messageBalanceProofIsNotGenerated) {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf(messageErrorMessage, response.ErrorMessage)
		return nil, fmt.Errorf("failed to get spendWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get spendWitness balance proof response: %v", response)
	}

	return response, nil
}

// Execute the following request:
// curl $API_BALANCE_VALIDITY_PROVER_URL/proof/{:intMaxAddress}/send/{:blockHash} | jq
func (p *balanceProcessor) fetchReceiveTransferBalanceValidityProof(publicKey *intMaxAcc.PublicKey, requestID string) (*BalanceValidityProofResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	intMaxAddress := publicKey.ToAddress().String()
	apiUrl := fmt.Sprintf("%s/proof/%s/transfer/%s", p.cfg.BalanceValidityProver.BalanceValidityProverUrl, intMaxAddress, requestID)

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
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.New(msgFailedToGetResponse)
		p.log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(BalanceValidityProofResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf(messageFailedToUnmarshalResponse, err)
	}

	if !response.Success {
		if response.ErrorMessage != nil && strings.HasPrefix(*response.ErrorMessage, messageBalanceProofIsNotGenerated) {
			return nil, ErrBalanceProofNotGenerated
		}

		p.log.Warnf(messageErrorMessage, response.ErrorMessage)
		return nil, fmt.Errorf("failed to get transferWitness balance proof response: %v", response)
	}

	if response.Proof == nil {
		return nil, fmt.Errorf("failed to get transferWitness balance proof response: %v", response)
	}

	return response, nil
}
