//nolint:gocritic
package withdrawal_service

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	WithdrawalThreshold = 8
	WaitDuration        = 15 * time.Minute
)

type WithdrawalAggregatorService struct {
	ctx                context.Context
	cfg                *configs.Config
	log                logger.Logger
	db                 SQLDriverApp
	client             *ethclient.Client
	withdrawalContract *bindings.Withdrawal
}

func newWithdrawalAggregatorService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) (*WithdrawalAggregatorService, error) {
	link, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	}

	client, err := utils.NewClient(link)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	withdrawalContract, err := bindings.NewWithdrawal(common.HexToAddress(cfg.Blockchain.WithdrawalContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ScrollMessenger contract: %w", err)
	}

	return &WithdrawalAggregatorService{
		ctx:                ctx,
		cfg:                cfg,
		log:                log,
		db:                 db,
		client:             client,
		withdrawalContract: withdrawalContract,
	}, nil
}

func WithdrawalAggregator(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	service, err := newWithdrawalAggregatorService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalAggregatorService: %v", err.Error()))
	}

	pendingWithdrawals, err := service.fetchPendingWithdrawals()
	if err != nil {
		panic(fmt.Sprintf("Failed to retrieve withdrawals %v", err.Error()))
	}

	if len(*pendingWithdrawals) == 0 {
		log.Infof("No pending withdrawal requests found")
		return
	}

	shouldSubmit := service.shouldProcessWithdrawals(*pendingWithdrawals)
	if !shouldSubmit {
		log.Infof("Not enough pending withdrawal requests to process")
		return
	}

	proofs, err := service.fetchWithdrawalProofsFromProver(*pendingWithdrawals)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch withdrawal proofs %v", err.Error()))
	}

	err = service.buildData(*pendingWithdrawals, proofs)
	if err != nil {
		panic("NEED_TO_BE_IMPLEMENTED")
	}

	// TODO: change status depends on the result of the proof
	receipt, err := service.submitWithdrawalProof()
	if err != nil {
		panic(fmt.Sprintf("Failed to submit withdrawal proof: %v", err.Error()))
	}

	if receipt == nil {
		panic("Received nil receipt for transaction")
	}

	switch receipt.Status {
	case types.ReceiptStatusSuccessful:
		log.Infof("Successfully submit withdrawal proof. Transaction Hash: %v", receipt.TxHash.Hex())
	case types.ReceiptStatusFailed:
		panic(fmt.Sprintf("Transaction failed: submit withdrawal proof unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
	default:
		panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
	}

	// TODO: change status true
}

func (w *WithdrawalAggregatorService) fetchPendingWithdrawals() (*[]mDBApp.Withdrawal, error) {
	limit := int(WithdrawalThreshold)
	withdrawals, err := w.db.WithdrawalsByStatus(mDBApp.PENDING, &limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending withdrawals: %w", err)
	}
	if withdrawals == nil {
		return nil, fmt.Errorf("failed to get pending withdrawals because withdrawals is nil")
	}
	return withdrawals, nil
}

func (w *WithdrawalAggregatorService) shouldProcessWithdrawals(pendingWithdrawals []mDBApp.Withdrawal) bool {
	if len(pendingWithdrawals) < WithdrawalThreshold {
		return false
	}

	minCreatedAt := pendingWithdrawals[0].CreatedAt
	for _, withdrawal := range pendingWithdrawals[1:] {
		if withdrawal.CreatedAt.Before(minCreatedAt) {
			minCreatedAt = withdrawal.CreatedAt
		}
	}

	return time.Since(minCreatedAt) >= WaitDuration
}

func (w *WithdrawalAggregatorService) fetchWithdrawalProofsFromProver(pendingWithdrawals []mDBApp.Withdrawal) ([]ProofValue, error) {
	var idsQuery string
	for _, pendingWithdrawal := range pendingWithdrawals {
		idsQuery += fmt.Sprintf("ids=%s&", pendingWithdrawal.ID)
	}
	if len(idsQuery) > 0 {
		idsQuery = idsQuery[:len(idsQuery)-1]
	}
	apiUrl := fmt.Sprintf("%s/proofs?%s",
		w.cfg.API.WithdrawalProverApiURL,
		idsQuery,
	)

	resp, err := http.Get(apiUrl) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var res ProofsResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		return nil, fmt.Errorf("prover request failed %s", res.ErrorMessage)
	}

	return res.Values, nil
}

func (w *WithdrawalAggregatorService) buildData(pendingWithdrawals []mDBApp.Withdrawal, proofs []ProofValue) error {
	return nil
}

func (w *WithdrawalAggregatorService) submitWithdrawalProof() (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(w.cfg)
	if err != nil {
		return nil, err
	}

	var withdrawals []bindings.ChainedWithdrawalLibChainedWithdrawal
	publicInputs := bindings.WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs{
		LastWithdrawalHash:   common.HexToHash("0x0"),
		WithdrawalAggregator: common.HexToAddress("0x0"),
	}
	var proof []byte

	tx, err := w.withdrawalContract.SubmitWithdrawalProof(transactOpts, withdrawals, publicInputs, proof)
	if err != nil {
		return nil, fmt.Errorf("failed to send submit withdrawal proof transaction: %w", err)
	}

	receipt, err := bind.WaitMined(w.ctx, w.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}
