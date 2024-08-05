package withdrawal_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5"
)

const blocksToLookBack = 10000

type WithdrawalsQueuedEventInfo struct {
	MaxDirectWithdrawalId    *uint64
	MaxClaimableWithdrawalId *uint64
}

type LastWithdrawalIds struct {
	DirectWithdrawalId    *uint64
	ClaimableWithdrawalId *uint64
}

type WithdrawalRelayerService struct {
	ctx                context.Context
	cfg                *configs.Config
	log                logger.Logger
	db                 SQLDriverApp
	scrollClient       *ethclient.Client
	withdrawalContract *bindings.Withdrawal
}

func newWithdrawalRelayerService(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) (*WithdrawalRelayerService, error) {
	scrollLink, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Scroll network chain link: %w", err)
	}

	scrollClient, err := utils.NewClient(scrollLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create new scrollClient: %w", err)
	}

	withdrawalContract, err := bindings.NewWithdrawal(common.HexToAddress(cfg.Blockchain.WithdrawalContractAddress), scrollClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Withdrawal contract: %w", err)
	}

	return &WithdrawalRelayerService{
		ctx:                ctx,
		cfg:                cfg,
		log:                log,
		scrollClient:       scrollClient,
		db:                 db,
		withdrawalContract: withdrawalContract,
	}, nil
}

func WithdrawalRelayer(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	service, err := newWithdrawalRelayerService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalRelayerService: %v", err.Error()))
	}

	lastWithdrawalIds, err := service.fetchLastRelayedWithdrawalIds()
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch last withdraral ids: %v", err.Error()))
	}

	event, err := db.EventBlockNumberByEventName(mDBApp.WithdrawalsQueuedEvent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "not found" {
			event = &mDBApp.EventBlockNumber{
				EventName:                mDBApp.WithdrawalsQueuedEvent,
				LastProcessedBlockNumber: 0,
			}
		} else {
			panic(fmt.Sprintf("Error fetching event block number: %v", err.Error()))
		}
	} else if event == nil {
		event = &mDBApp.EventBlockNumber{
			EventName:                mDBApp.WithdrawalsQueuedEvent,
			LastProcessedBlockNumber: 0,
		}
	}

	currentBlockNumber, err := service.scrollClient.BlockNumber(service.ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to get current block number: %v", err.Error()))
	}

	eventInfo, err := service.fetchLastWithdrawalsQueuedEvent(currentBlockNumber, event.LastProcessedBlockNumber)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch last WithdrawalsQueued event: %v", err.Error()))
	}

	lastDirectWithdrawalId := *eventInfo.MaxDirectWithdrawalId
	lastClaimableWithdrawalId := *eventInfo.MaxClaimableWithdrawalId

	if lastWithdrawalIds.DirectWithdrawalId != nil && *lastWithdrawalIds.DirectWithdrawalId > lastDirectWithdrawalId {
		lastDirectWithdrawalId = *lastWithdrawalIds.DirectWithdrawalId
	}
	if lastWithdrawalIds.ClaimableWithdrawalId != nil && *lastWithdrawalIds.ClaimableWithdrawalId > lastClaimableWithdrawalId {
		lastClaimableWithdrawalId = *lastWithdrawalIds.ClaimableWithdrawalId
	}

	if lastDirectWithdrawalId != *lastWithdrawalIds.DirectWithdrawalId || lastClaimableWithdrawalId != *lastWithdrawalIds.ClaimableWithdrawalId {
		var receipt *types.Receipt
		receipt, err = service.relayWithdrawals(lastDirectWithdrawalId, lastClaimableWithdrawalId)
		if err != nil {
			panic(fmt.Sprintf("Failed to relay withdrawals: %v", err.Error()))
		}

		if receipt == nil {
			panic("Received nil receipt for transaction")
		}

		switch receipt.Status {
		case types.ReceiptStatusSuccessful:
			log.Infof(
				"Successfully relay withdrawals lastDirectWithdrawalId: %d, lastClaimableWithdrawalId: %d. Transaction Hash: %v",
				lastDirectWithdrawalId,
				lastClaimableWithdrawalId,
				receipt.TxHash.Hex(),
			)
		case types.ReceiptStatusFailed:
			panic(fmt.Sprintf("Transaction failed: relay withdrawals unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
		default:
			panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
		}

		_, err = db.UpsertEventBlockNumber(mDBApp.WithdrawalsQueuedEvent, currentBlockNumber)
		if err != nil {
			panic(fmt.Sprintf("Error updating event block number: %v", err.Error()))
		}
	} else {
		log.Infof("No new withdrawals to relay")
	}
}

func (w *WithdrawalRelayerService) relayWithdrawals(maxDirectWithdrawalId, maxClaimableWithdrawalId uint64) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(w.cfg.Blockchain.WithdrawalPrivateKeyHex, w.cfg.Blockchain.ScrollNetworkChainID)
	if err != nil {
		return nil, err
	}

	tx, err := w.withdrawalContract.RelayWithdrawals(transactOpts, big.NewInt(int64(maxDirectWithdrawalId)), big.NewInt(int64(maxClaimableWithdrawalId)))
	if err != nil {
		return nil, fmt.Errorf("failed to send relay withdrawals transaction: %w", err)
	}

	receipt, err := bind.WaitMined(w.ctx, w.scrollClient, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func (w *WithdrawalRelayerService) fetchLastWithdrawalsQueuedEvent(currentBlockNumber, lastProcessedBlockNumber uint64) (*WithdrawalsQueuedEventInfo, error) {
	startBlockNumber := w.calculateStartBlockNumber(currentBlockNumber, lastProcessedBlockNumber)
	iterator, err := w.withdrawalContract.FilterWithdrawalsQueued(&bind.FilterOpts{
		Start: startBlockNumber,
		End:   &currentBlockNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	defer iterator.Close()

	var maxDirectWithdrawalId uint64 = 0
	var maxClaimableWithdrawalId uint64 = 0

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
		}
		lastDirectWithdrawalId := iterator.Event.LastDirectWithdrawalId.Uint64()
		lastClaimableWithdrawalId := iterator.Event.LastClaimableWithdrawalId.Uint64()

		if lastDirectWithdrawalId > maxDirectWithdrawalId {
			maxDirectWithdrawalId = lastDirectWithdrawalId
		}
		if lastClaimableWithdrawalId > maxClaimableWithdrawalId {
			maxClaimableWithdrawalId = lastClaimableWithdrawalId
		}
	}

	return &WithdrawalsQueuedEventInfo{
		MaxDirectWithdrawalId:    &maxDirectWithdrawalId,
		MaxClaimableWithdrawalId: &maxClaimableWithdrawalId,
	}, nil
}

func (w *WithdrawalRelayerService) calculateStartBlockNumber(currentBlockNumber, lastProcessedBlockNumber uint64) uint64 {
	if lastProcessedBlockNumber == 0 {
		return max(currentBlockNumber-blocksToLookBack, 0)
	}
	return lastProcessedBlockNumber + 1
}

func (w *WithdrawalRelayerService) fetchLastRelayedWithdrawalIds() (LastWithdrawalIds, error) {
	type result struct {
		lastWithdrawalId uint64
		err              error
	}

	lastRelayedDirectWithdrawalCh := make(chan result)
	lastRelayedClaimableWithdrawalCh := make(chan result)

	go func() {
		lastWithdrawalId, err := w.fetchLastRelayedDirectWithdrawalId()
		lastRelayedDirectWithdrawalCh <- result{lastWithdrawalId, err}
	}()

	go func() {
		lastWithdrawalId, err := w.fetchLastRelayedClaimableWithdrawalId()
		lastRelayedClaimableWithdrawalCh <- result{lastWithdrawalId, err}
	}()

	var li LastWithdrawalIds
	for {
		if li.DirectWithdrawalId != nil && li.ClaimableWithdrawalId != nil {
			return li, nil
		}
		select {
		case lastRelayedDirectWithdrawalResult := <-lastRelayedDirectWithdrawalCh:
			if lastRelayedDirectWithdrawalResult.err != nil {
				return LastWithdrawalIds{}, lastRelayedDirectWithdrawalResult.err
			}
			li.DirectWithdrawalId = &lastRelayedDirectWithdrawalResult.lastWithdrawalId
		case lastRelayedClaimabletWithdrawalResult := <-lastRelayedClaimableWithdrawalCh:
			if lastRelayedClaimabletWithdrawalResult.err != nil {
				return LastWithdrawalIds{}, lastRelayedClaimabletWithdrawalResult.err
			}
			li.ClaimableWithdrawalId = &lastRelayedClaimabletWithdrawalResult.lastWithdrawalId
		}
	}
}

func (w *WithdrawalRelayerService) fetchLastRelayedDirectWithdrawalId() (uint64, error) {
	lastLastRelayedDirectWithdrawalId, err := w.withdrawalContract.GetLastRelayedDirectWithdrawalId(&bind.CallOpts{
		Pending: false,
		Context: w.ctx,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get last relayed direct withdrawal id: %w", err)
	}
	if !lastLastRelayedDirectWithdrawalId.IsUint64() {
		return 0, fmt.Errorf("last relayed direct withdrawal id exceeds uint64 range")
	}
	return lastLastRelayedDirectWithdrawalId.Uint64(), nil
}

func (w *WithdrawalRelayerService) fetchLastRelayedClaimableWithdrawalId() (uint64, error) {
	lastLastRelayedClaimableWithdrawalId, err := w.withdrawalContract.GetLastRelayedClaimableWithdrawalId(&bind.CallOpts{
		Pending: false,
		Context: w.ctx,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get last relayed claimable withdrawal id: %w", err)
	}
	if !lastLastRelayedClaimableWithdrawalId.IsUint64() {
		return 0, fmt.Errorf("last relayed claimable withdrawal id exceeds uint64 range")
	}
	return lastLastRelayedClaimableWithdrawalId.Uint64(), nil
}
