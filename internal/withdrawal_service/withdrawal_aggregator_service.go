//nolint:gocritic
package withdrawal_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const WithdrawalThreshold = 8

type WithdrawalAggregatorService struct {
	ctx             context.Context
	cfg             *configs.Config
	log             logger.Logger
	db              SQLDriverApp
	client          *ethclient.Client
	scrollMessenger *bindings.L1ScrollMessanger
}

func newWithdrawalAggregatorService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	_ ServiceBlockchain,
) (*WithdrawalAggregatorService, error) {
	// link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	// }

	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	scrollMessenger, err := bindings.NewL1ScrollMessanger(common.HexToAddress(cfg.Blockchain.ScrollMessengerL1ContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ScrollMessenger contract: %w", err)
	}

	return &WithdrawalAggregatorService{
		ctx:             ctx,
		cfg:             cfg,
		log:             log,
		db:              db,
		client:          client,
		scrollMessenger: scrollMessenger,
	}, nil
}

func WithdrawalAggregator(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) error {
	withdrawalAggregatorService, err := newWithdrawalAggregatorService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalAggregatorService: %v", err.Error()))
	}

	withdrawals, err := withdrawalAggregatorService.getWithdrawalsFromDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to retrieve withdrawals %v", err.Error()))
	}

	if len(*withdrawals) == 0 {
		fmt.Println("Not found withdrawals")
		return nil
	}

	for _, withdrawal := range *withdrawals {
		_, err := withdrawalAggregatorService.generateZKProof(withdrawal.ID, withdrawal.Recipient, withdrawal.TokenIndex, withdrawal.Amount, withdrawal.Salt, withdrawal.TransferHash)
		if err != nil {
			panic(fmt.Sprintf("Failed to retrieve withdrawals %v", err.Error()))
		}
	}

	return nil
}

func SubmitWithdrawals(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) error {
	_, err := newWithdrawalAggregatorService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalAggregatorService: %v", err.Error()))
	}
	return nil
}

func (w *WithdrawalAggregatorService) getWithdrawalsFromDB() (*[]mDBApp.Withdrawal, error) {
	withdrawals, err := w.db.FindWithdrawalsByGroupStatus(mDBApp.PENDING)
	if err != nil {
		return nil, fmt.Errorf("failed to find withdrawals: %w", err)
	}
	if withdrawals == nil {
		return nil, fmt.Errorf("failed to get withdrawals because withdrawals is nil")
	}

	return withdrawals, nil
}

func (w *WithdrawalAggregatorService) generateZKProof(id string, recipient string, tokenIndex int, amount string, salt string, blockHash string) (error, error) {
	requestBody := map[string]interface{}{
		"id":         id,
		"recipient":  recipient,
		"tokenIndex": tokenIndex,
		"amount":     amount,
		"salt":       salt,
		"blockHash":  blockHash,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	resp, err := http.Post(w.cfg.API.WithdrawalProverApiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer resp.Body.Close()

	var res WithdrwalProverResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil, nil
}
