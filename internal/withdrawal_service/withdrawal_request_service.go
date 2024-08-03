//nolint:gocritic
package withdrawal_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
)

var ErrScrollNetworkChainLinkEvmJSONRPCFail = errors.New(
	"failed to get the chain-link-evm-json-rpc of scroll network",
)

var ErrCreateNewClientOfRPCEthFail = errors.New(
	"failed to create new RPC Eth client",
)

type WithdrawalRequestService struct {
	ctx    context.Context
	cfg    *configs.Config
	log    logger.Logger
	db     SQLDriverApp
	sb     ServiceBlockchain
	rollup *bindings.Rollup
}

func newWithdrawalRequestService(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) (*WithdrawalRequestService, error) {
	const (
		hName = "BlockBuilderRegistryService func:StopBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	link, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.Blockchain.RollupContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &WithdrawalRequestService{
		ctx:    ctx,
		cfg:    cfg,
		log:    log,
		db:     db,
		sb:     sb,
		rollup: rollup,
	}, nil
}

// TODO: NEED_TO_BE_CHANGED
func PostWithdrawalRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	input *postWithdrwalRequest.UCPostWithdrawalRequestInput,
) error {
	service, err := newWithdrawalRequestService(ctx, cfg, log, db, sb)
	if err != nil {
		return fmt.Errorf("failed to create new withdrawal request service: %w", err)
	}

	err = service.verifyBalanceProof()
	if err != nil {
		return fmt.Errorf("failed to verify balance proof: %w", err)
	}

	err = service.checkBlockNumber(input)
	if err != nil {
		return fmt.Errorf("failed to send withdrawal request to prover: %w", err)
	}

	id := uuid.New().String()
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

// TODO: NEED_TO_BE_IMPLEMENTED
func (s *WithdrawalRequestService) verifyBalanceProof() error {
	// Access to the Balance Validatity Prover
	return nil
}

// Check the block number
func (s *WithdrawalRequestService) checkBlockNumber(input *postWithdrwalRequest.UCPostWithdrawalRequestInput) error {
	if input.BlockNumber >= uint64(1)<<32 {
		return fmt.Errorf("block number is too large")
	}

	blockHash := common.HexToHash(input.BlockHash)
	opts := bind.CallOpts{
		Pending: false,
		Context: s.ctx,
	}

	actualBlockHash, err := s.rollup.GetBlockHash(&opts, uint32(input.BlockNumber))
	if err != nil {
		return fmt.Errorf("failed to get block hash: %w", err)
	}

	if blockHash != actualBlockHash {
		return fmt.Errorf("block hash is not matched")
	}

	return nil
}

// TODO: NEED_TO_BE_CHANGED
func (w *WithdrawalRequestService) requestWithdrawalProofToProver(id string, input *postWithdrwalRequest.UCPostWithdrawalRequestInput) error {
	return nil
	/*
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
		if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return fmt.Errorf("failed to decode JSON response: %w", err)
		}

		if !res.Success {
			return fmt.Errorf("prover request failed %s", res.ErrorMessage)
		}

		return nil
	*/
}
