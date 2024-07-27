package withdrawal_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	"net/http"

	"github.com/google/uuid"
)

type WithdrawalRequestService struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func newWithdrawalRequestService(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp) *WithdrawalRequestService {
	return &WithdrawalRequestService{
		ctx: ctx,
		cfg: cfg,
		log: log,
		db:  db,
	}
}

// TODO: NEED_TO_BE_UPDATE logic
func PostWithdrawalRequest(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, input postWithdrwalRequest.UCPostWithdrawalRequestInput) error {
	service := newWithdrawalRequestService(ctx, cfg, log, db)

	err := service.verifyBalanceProof()
	if err != nil {
		return fmt.Errorf("failed to verify balance proof: %w", err)
	}

	id := uuid.New().String()
	// TODO: NEED_TO_BE_CHANGED
	err = service.requestWithdrawalProofToProver(id, input)
	if err != nil {
		return fmt.Errorf("failed to send withdrawal request to prover: %w", err)
	}

	_, err = db.CreateWithdrawal(id, input)
	if err != nil {
		return fmt.Errorf("failed to save withdrawal request to db: %w", err)
	}

	return nil
}

// TODO: Access to the Balance Validatity Prover
func (s *WithdrawalRequestService) verifyBalanceProof() error {
	return nil
}

func (w *WithdrawalRequestService) requestWithdrawalProofToProver(id string, input postWithdrwalRequest.UCPostWithdrawalRequestInput) error {
	requestBody := map[string]interface{}{
		"id":         id,
		"recipient":  input.TransferData.Recipient,
		"tokenIndex": input.TransferData.TokenIndex,
		"amount":     input.TransferData.Amount,
		"salt":       input.TransferData.Salt,
		"blockHash":  input.BlockHash,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	apiUrl := fmt.Sprintf("%s/proof",
		w.cfg.API.WithdrawalProverApiURL,
	)
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonBody)) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res GenerateProofResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("prover request failed %s", res.ErrorMessage)
	}

	return nil
}
