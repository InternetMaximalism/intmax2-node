package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	verifyDepositConfirmation "intmax2-node/internal/use_cases/verify_deposit_confirmation"
	"intmax2-node/pkg/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type VerifyDepositConfirmationService struct {
	ctx          context.Context
	cfg          *configs.Config
	log          logger.Logger
	client       *ethclient.Client
	scrollClient *ethclient.Client
	liquidity    *bindings.Liquidity
	rollup       *bindings.Rollup
}

func newVerifyDepositConfirmationService(ctx context.Context, cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) (*VerifyDepositConfirmationService, error) {
	scrollLink, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Scroll network chain link: %w", err)
	}

	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	scrollClient, err := utils.NewClient(scrollLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create new scroll client: %w", err)
	}
	defer client.Close()

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.Blockchain.RollupContractAddress), scrollClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Rollup contract: %w", err)
	}

	return &VerifyDepositConfirmationService{
		ctx:          ctx,
		cfg:          cfg,
		log:          log,
		client:       client,
		scrollClient: scrollClient,
		liquidity:    liquidity,
		rollup:       rollup,
	}, nil
}

func GetVerifyDepositConfirmation(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	input *verifyDepositConfirmation.UCGetVerifyDepositConfirmationInput,
) (bool, error) {
	service, err := newVerifyDepositConfirmationService(ctx, cfg, log, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize VerifyDepositConfirmationService: %v", err.Error()))
	}

	depositId := new(big.Int)
	_, success := depositId.SetString(input.DepositId, 10)
	if !success {
		panic(fmt.Sprintf("Failed to set depositId: %v", input.DepositId))
	}

	service.getDepositData(depositId)
	service.getDepositCanceled(depositId)
	service.getLastProcessedDepositId(depositId)

	return true, nil
}

func (v *VerifyDepositConfirmationService) getDepositData(depositId *big.Int) (*bindings.DepositQueueLibDepositData, error) {
	result, err := v.liquidity.GetDepositData(&bind.CallOpts{
		Pending: false,
		Context: v.ctx,
	}, depositId)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit data: %w", err)
	}
	return &result, nil
}

func (v *VerifyDepositConfirmationService) getDepositCanceled(depositId *big.Int) (error, error) {
	depositIds := []*big.Int{
		depositId,
	}
	iterator, err := v.liquidity.FilterDepositCanceled(&bind.FilterOpts{
		Start: 0,
		End:   nil,
	}, depositIds)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	defer iterator.Close()

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
		}

		DepositId := iterator.Event.DepositId
		fmt.Println("DepositId: ", DepositId)
	}

	return nil, nil
}

func (v *VerifyDepositConfirmationService) getLastProcessedDepositId(depositId *big.Int) (error, error) {
	depositIds := []*big.Int{
		depositId,
	}
	iterator, err := v.rollup.FilterDepositsProcessed(&bind.FilterOpts{
		Start: 0,
		End:   nil,
	}, depositIds)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	defer iterator.Close()

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
		}

		LastProcessedDepositId := iterator.Event.LastProcessedDepositId
		fmt.Println("LastProcessedDepositId: ", LastProcessedDepositId)
	}

	return nil, nil
}
