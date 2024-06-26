package withdrawal_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const defaultPage = 1

type WithdrawalRelayerService struct {
	ctx             context.Context
	cfg             *configs.Config
	log             logger.Logger
	client          *ethclient.Client
	scrollMessenger *bindings.L1ScrollMessanger
}

func newWithdrawalRelayerService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) (*WithdrawalRelayerService, error) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	}

	client, err := utils.NewClient(link)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	scrollMessenger, err := bindings.NewL1ScrollMessanger(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ScrollMessenger contract: %w", err)
	}

	return &WithdrawalRelayerService{
		ctx:             ctx,
		cfg:             cfg,
		log:             log,
		client:          client,
		scrollMessenger: scrollMessenger,
	}, nil
}

func WithdrawalRelayer(ctx context.Context, cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) {
	withdrawalRelayerService, err := newWithdrawalRelayerService(ctx, cfg, log, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalRelayerService: %v", err.Error()))
	}

	claimableRequests, err := withdrawalRelayerService.fetchClaimableScrollMessengerRequests()
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch claimable requests: %v", err.Error()))
	}

	if len(claimableRequests) == 0 {
		log.Infof("No claimable requests found")
		return
	}

	for _, claimableRequest := range claimableRequests {
		var receipt *types.Receipt
		receipt, err = withdrawalRelayerService.relayMessageWithProof(claimableRequest)
		if err != nil {
			log.Warnf("Failed to submit relayMessageWithProof: %v", err.Error())
			continue
		}

		if receipt == nil {
			log.Warnf("Received nil receipt for transaction")
			continue
		}

		if receipt.Status == types.ReceiptStatusSuccessful {
			log.Infof("Successfully submitted relayMessageWithProof claimableRequest hash: %s", claimableRequest.Hash)
		} else {
			log.Warnf("Failed to submit relayMessageWithProof")
		}

		log.Infof("Transaction Hash: %v", receipt.TxHash.Hex())
	}

	log.Infof("Submitted relay message with proof for %d claimableRequests", len(claimableRequests))
}

func (w *WithdrawalRelayerService) fetchClaimableScrollMessengerRequests() ([]*ScrollMessengerResult, error) {
	apiUrl := fmt.Sprintf("%s/api/l2/unclaimed/withdrawals?address=%s&page_size=10&page=%d",
		w.cfg.Blockchain.ScrollBridgeApiUrl,
		w.cfg.Blockchain.ScrollMessengerL2ContractAddress,
		defaultPage,
	)

	resp, err := http.Get(apiUrl) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var res ScrollMessengerResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if res.Data.Results == nil || len(res.Data.Results) == 0 {
		return []*ScrollMessengerResult{}, nil
	}

	return filterClaimableResults(res.Data.Results), nil
}

func filterClaimableResults(results []*ScrollMessengerResult) (filtered []*ScrollMessengerResult) {
	for _, result := range results {
		if result.ClaimInfo != nil && result.ClaimInfo.Claimable {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func (w *WithdrawalRelayerService) relayMessageWithProof(result *ScrollMessengerResult) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(w.cfg)
	if err != nil {
		return nil, err
	}

	message := []byte(result.ClaimInfo.Message)

	value, err := utils.StringToBigInt(result.ClaimInfo.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid value string: %w", err)
	}

	nonce, err := utils.StringToBigInt(result.ClaimInfo.Nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce string: %w", err)
	}

	batchIndex, err := utils.StringToBigInt(result.ClaimInfo.Proof.BatchIndex)
	if err != nil {
		return nil, fmt.Errorf("invalid batchIndex string: %w", err)
	}

	merkleProof := []byte(result.ClaimInfo.Proof.MerkleProof)
	proof := bindings.IL1ScrollMessengerL2MessageProof{
		BatchIndex:  batchIndex,
		MerkleProof: merkleProof,
	}

	tx, err := w.scrollMessenger.RelayMessageWithProof(
		transactOpts,
		common.HexToAddress(result.ClaimInfo.From),
		common.HexToAddress(result.ClaimInfo.To),
		value,
		nonce,
		message,
		proof,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send relayMessageWithProof transaction: %w", err)
	}

	receipt, err := bind.WaitMined(w.ctx, w.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}