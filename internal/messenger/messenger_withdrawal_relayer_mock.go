package messenger

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

type MessengerWithdrawalRelayerMockService struct {
	ctx               context.Context
	cfg               *configs.Config
	log               logger.Logger
	db                SQLDriverApp
	ethClient         *ethclient.Client
	scrollClient      *ethclient.Client
	l1ScrollMessenger *bindings.L1ScrollMessenger
	l2ScrollMessenger *bindings.L2ScrollMessenger
}

func newMessengerWithdrawalRelayerMockService(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) (*MessengerWithdrawalRelayerMockService, error) {
	scrollLink, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	}

	ethClient, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	scrollClient, err := utils.NewClient(scrollLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	l1ScrollMessenger, err := bindings.NewL1ScrollMessenger(common.HexToAddress(cfg.Blockchain.ScrollMessengerL1ContractAddress), ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate L1ScrollMessenger contract: %w", err)
	}

	l2ScrollMessenger, err := bindings.NewL2ScrollMessenger(common.HexToAddress(cfg.Blockchain.ScrollMessengerL2ContractAddress), scrollClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate L2ScrollMessenger contract: %w", err)
	}

	return &MessengerWithdrawalRelayerMockService{
		ctx:               ctx,
		cfg:               cfg,
		log:               log,
		db:                db,
		ethClient:         ethClient,
		scrollClient:      scrollClient,
		l1ScrollMessenger: l1ScrollMessenger,
		l2ScrollMessenger: l2ScrollMessenger,
	}, nil
}

func MessengerWithdrawalRelayerMock(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	service, err := newMessengerWithdrawalRelayerMockService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalRelayerMockService: %v", err.Error()))
	}

	event, err := db.EventBlockNumberByEventName(mDBApp.SentMessageEvent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == notFound {
			event = &mDBApp.EventBlockNumber{
				EventName:                mDBApp.SentMessageEvent,
				LastProcessedBlockNumber: 0,
			}
		} else {
			panic(fmt.Sprintf("Error fetching event block number: %v", err.Error()))
		}
	} else if event == nil {
		event = &mDBApp.EventBlockNumber{
			EventName:                mDBApp.SentMessageEvent,
			LastProcessedBlockNumber: 0,
		}
	}

	currentBlockNumber, err := service.scrollClient.BlockNumber(service.ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to get current block number: %v", err.Error()))
	}

	events, err := service.fetchSentMessageEvents(currentBlockNumber, event.LastProcessedBlockNumber)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch sent message events: %v", err.Error()))
	}

	if len(events) == 0 {
		log.Infof("No events found")
		return
	}

	successfulEvents := 0
	for _, event := range events {
		var receipt *types.Receipt
		receipt, err = service.relayMessageWithProofByEvent(event)
		if err != nil {
			log.Warnf("Failed to submit relayMessageWithProofByEvent: %v", err.Error())
			continue
		}

		if receipt == nil {
			log.Warnf("Received nil receipt for transaction")
			continue
		}

		switch receipt.Status {
		case types.ReceiptStatusSuccessful:
			log.Infof("Successfully relayed message with proof by event. Transaction Hash: %v", receipt.TxHash.Hex())
			successfulEvents++
		case types.ReceiptStatusFailed:
			panic(fmt.Sprintf("Transaction failed: relay message with proof by event unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
		default:
			panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
		}
	}

	log.Infof("Successfully submitted relay message with proof by event for %d out of %d events", len(events), successfulEvents)

	_, err = db.UpsertEventBlockNumber(mDBApp.SentMessageEvent, currentBlockNumber)
	if err != nil {
		panic(fmt.Sprintf("Error updating event block number: %v", err.Error()))
	}
}

func (w *MessengerWithdrawalRelayerMockService) fetchSentMessageEvents(currentBlockNumber, lastProcessedBlockNumber uint64) ([]*bindings.L2ScrollMessengerSentMessage, error) {
	startBlockNumber := w.calculateStartBlockNumber(currentBlockNumber, lastProcessedBlockNumber)

	iterator, err := w.l2ScrollMessenger.FilterSentMessage(&bind.FilterOpts{
		Start:   startBlockNumber,
		End:     &currentBlockNumber,
		Context: w.ctx,
	}, []common.Address{}, []common.Address{})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	defer iterator.Close()

	var events []*bindings.L2ScrollMessengerSentMessage

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	if err = iterator.Error(); err != nil {
		return nil, fmt.Errorf("error encountered while iterating: %w", err)
	}

	return events, nil
}

func (w *MessengerWithdrawalRelayerMockService) relayMessageWithProofByEvent(event *bindings.L2ScrollMessengerSentMessage) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(w.cfg.Blockchain.MessengerMockPrivateKeyHex, w.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, err
	}

	batchIndex := big.NewInt(0)
	merkleProof := []byte{}
	proof := bindings.IL1ScrollMessengerL2MessageProof{
		BatchIndex:  batchIndex,
		MerkleProof: merkleProof,
	}

	tx, err := w.l1ScrollMessenger.RelayMessageWithProof(
		transactOpts,
		common.HexToAddress(event.Sender.Hex()),
		common.HexToAddress(event.Target.Hex()),
		event.Value,
		event.MessageNonce,
		event.Message,
		proof,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send relayMessageWithProof by event transaction: %w", err)
	}

	receipt, err := bind.WaitMined(w.ctx, w.ethClient, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func (w *MessengerWithdrawalRelayerMockService) calculateStartBlockNumber(currentBlockNumber, lastProcessedBlockNumber uint64) uint64 {
	if lastProcessedBlockNumber == 0 {
		return max(currentBlockNumber-blocksToLookBack, 0)
	}
	return lastProcessedBlockNumber + 1
}
