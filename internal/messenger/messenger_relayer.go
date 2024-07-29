//nolint:gocritic
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

	"github.com/jackc/pgx/v5"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type MessengerRelayerService struct {
	ctx               context.Context
	cfg               *configs.Config
	log               logger.Logger
	db                SQLDriverApp
	client            *ethclient.Client
	l1ScrollMessenger *bindings.L1ScrollMessenger
	l2ScrollMessenger *bindings.L2ScrollMessenger
}

func newMessengerRelayerService(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) (*MessengerRelayerService, error) {
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

	return &MessengerRelayerService{
		ctx:               ctx,
		cfg:               cfg,
		log:               log,
		db:                db,
		l1ScrollMessenger: l1ScrollMessenger,
		l2ScrollMessenger: l2ScrollMessenger,
	}, nil
}

func MessengerRelayer(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	messengerService, err := newMessengerRelayerService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize MockMessengerService: %v", err.Error()))
	}

	event, err := db.EventBlockNumberByEventName(mDBApp.SentMessageEvent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "not found" {
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

	events, lastBlockNumber, err := messengerService.fetchNewSentMessages(uint64(event.LastProcessedBlockNumber))
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch new sent messages: %v", err.Error()))
	}

	if len(events) == 0 {
		log.Infof("No new SentMessage Events")
		return
	}

	messengerService.relayMessagesforEvents(events)

	err = updateEventBlockNumber(db, log, mDBApp.SentMessageEvent, int64(lastBlockNumber))
	if err != nil {
		panic(fmt.Sprintf("Failed to update event block number: %v", err.Error()))
	}
}

func (m *MessengerRelayerService) fetchNewSentMessages(lastProcessedBlockNumber uint64) (_ []*bindings.L1ScrollMessengerSentMessage, _ uint64, _ error) {
	var lastBlockNumber uint64

	startBlock := lastProcessedBlockNumber + 1
	iterator, err := m.l1ScrollMessenger.FilterSentMessage(&bind.FilterOpts{
		Start:   startBlock,
		End:     nil,
		Context: m.ctx,
	}, []common.Address{}, []common.Address{})
	if err != nil {
		return nil, lastBlockNumber, fmt.Errorf("failed to filter logs: %w", err)
	}
	defer iterator.Close()

	var events []*bindings.L1ScrollMessengerSentMessage

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
		if event.Raw.BlockNumber > lastBlockNumber {
			lastBlockNumber = event.Raw.BlockNumber
		}
	}

	if err = iterator.Error(); err != nil {
		return nil, lastBlockNumber, fmt.Errorf("error encountered while iterating: %w", err)
	}

	return events, lastBlockNumber, nil
}

func (m *MessengerRelayerService) relayMessages(event *bindings.L1ScrollMessengerSentMessage) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(m.cfg.Blockchain.MockMessagingPrivateKeyHex, m.cfg.Blockchain.ScrollNetworkChainID)
	if err != nil {
		return nil, err
	}

	tx, err := m.l2ScrollMessenger.RelayMessage(transactOpts, event.Sender, event.Target, event.Value, event.MessageNonce, event.Message)
	if err != nil {
		return nil, err
	}

	receipt, err := bind.WaitMined(m.ctx, m.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	if receipt == nil {
		return nil, fmt.Errorf("received nil receipt for transaction")
	}

	return receipt, nil
}

func (m *MessengerRelayerService) relayMessagesforEvents(events []*bindings.L1ScrollMessengerSentMessage) {
	for _, event := range events {
		_, err := m.relayMessages(event)
		if err != nil {
			if err.Error() == "execution reverted: Message was already successfully executed" {
				m.log.Infof("Message was already successfully executed")
				continue
			} else {
				panic(fmt.Sprintf("Error relaying message: %v", err))
			}
		}
	}
}

func updateEventBlockNumber(db SQLDriverApp, log logger.Logger, eventName string, blockNumber int64) error {
	updatedEvent, err := db.UpsertEventBlockNumber(eventName, blockNumber)
	if err != nil {
		return err
	}
	log.Infof("Updated %s block number to %d", eventName, updatedEvent.LastProcessedBlockNumber)
	return nil
}
