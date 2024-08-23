package block_validity_prover

import (
	"errors"
	"fmt"
	"intmax2-node/internal/bindings"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type RelayMessageInput struct {
	From    common.Address
	To      common.Address
	Value   *big.Int
	Nonce   *big.Int
	Message []byte
}

type ProcessDepositsInput struct {
	LastProcessedDepositId *big.Int
	DepositHashes          [][int32Key]byte
}

func (b *mockBlockBuilder) LastSeenProcessDepositsEventBlockNumber() (uint64, error) {
	return b.lastSeenProcessDepositsEventBlockNumber, nil
}

func (b *mockBlockBuilder) SetLastSeenProcessDepositsEventBlockNumber(blockNumber uint64) error {
	b.lastSeenProcessDepositsEventBlockNumber = blockNumber

	return nil
}

func (b *mockBlockBuilder) LastDepositTreeRoot() (common.Hash, error) {
	return b.DepositTreeRoots[len(b.DepositTreeRoots)-1], nil
}

func (b *mockBlockBuilder) AppendDepositTreeRoot(root common.Hash) error {
	b.DepositTreeRoots = append(b.DepositTreeRoots, root)

	return nil
}

func (b *mockBlockBuilder) AppendDepositTreeLeaf(depositHash common.Hash) error {
	_, count, _ := b.DepositTree.GetCurrentRootCountAndSiblings()
	_, err := b.DepositTree.AddLeaf(count, depositHash)
	if err != nil {
		return err
	}

	return nil
}

func (p *blockValidityProver) SyncDepositTree() error {
	b := p.blockBuilder

	latestBlockNumber, err := p.scrollClient.BlockNumber(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %v", err.Error())
	}

	lastSeenProcessDepositsEventBlockNumber, err := b.LastSeenProcessDepositsEventBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get last seen process deposits event block number: %v", err.Error())
	}
	for lastSeenProcessDepositsEventBlockNumber < latestBlockNumber {
		p.log.Infof("Syncing deposits from block %d\n", lastSeenProcessDepositsEventBlockNumber)
		endBlock := lastSeenProcessDepositsEventBlockNumber + eventBlockRange

		var depositsProcessedEvents []*bindings.RollupDepositsProcessed
		depositsProcessedEvents, err = p.getDepositsProcessedEvent(lastSeenProcessDepositsEventBlockNumber, &endBlock)
		if err != nil {
			return err
		}
		p.log.Infof("Found %d deposits processed events\n", len(depositsProcessedEvents))

		for _, deposit := range depositsProcessedEvents {
			select {
			case <-p.ctx.Done():
				p.log.Warnf("Received cancel signal from context, stopping...")
				return p.ctx.Err()
			default:
				p.log.Infof("Processing deposits from block %d, depositId %d\n", deposit.Raw.BlockNumber, deposit.LastProcessedDepositId)
				var calldata []byte
				calldata, err = p.FetchScrollCalldataByHash(deposit.Raw.TxHash)
				if err != nil {
					return fmt.Errorf("failed to fetch calldata for tx %v: %v", deposit.Raw.TxHash, err.Error())
				}

				var relayMessageCalldata *RelayMessageInput
				relayMessageCalldata, err = formatRelayMessageCalldata(calldata)
				if err != nil {
					return fmt.Errorf("failed to decode relay message calldata: %v", err.Error())
				}

				var processDepositsCalldata *ProcessDepositsInput
				processDepositsCalldata, err = formatProcessDepositsCalldata(relayMessageCalldata.Message)
				if err != nil {
					return fmt.Errorf("failed to for relay message calldata: %v", err.Error())
				}

				for i := range processDepositsCalldata.DepositHashes {
					depositHash := processDepositsCalldata.DepositHashes[i]

					err = b.AppendDepositTreeLeaf(common.Hash(depositHash))
					if err != nil {
						return fmt.Errorf("failed to add deposit leaf: %v", err.Error())
					}
				}

				lastDepositRoot, err := b.LastDepositTreeRoot()
				if err != nil {
					return fmt.Errorf("failed to get latest deposit tree root: %v", err.Error())
				}
				if lastDepositRoot != common.Hash(deposit.DepositTreeRoot) {
					return fmt.Errorf("DepositTreeRoot mismatch: expected %v, got %v", common.Hash(deposit.DepositTreeRoot), lastDepositRoot)
				}

				b.AppendDepositTreeRoot(lastDepositRoot)
			}
		}

		b.SetLastSeenProcessDepositsEventBlockNumber(endBlock)

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (p *blockValidityProver) getDepositsProcessedEvent(
	startBlock uint64,
	endBlock *uint64,
) ([]*bindings.RollupDepositsProcessed, error) {
	nextBlock := startBlock + 1
	iterator, err := p.rollup.FilterDepositsProcessed(&bind.FilterOpts{
		Start:   nextBlock,
		End:     endBlock,
		Context: p.ctx,
	}, []*big.Int{})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	defer iterator.Close()

	var events []*bindings.RollupDepositsProcessed
	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	if err = iterator.Error(); err != nil {
		return nil, fmt.Errorf("error encountered while iterating: %w", err)
	}

	return events, nil
}

func formatRelayMessageCalldata(calldata []byte) (*RelayMessageInput, error) {
	messengerABI := io.Reader(strings.NewReader(bindings.L2ScrollMessengerMetaData.ABI))
	parsedABI, err := abi.JSON(messengerABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	method, err := parsedABI.MethodById(calldata[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse calldata: %w", err)
	}
	decodedInputs, err := decodeRelayMessageCalldata(method, calldata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode process deposits calldata: %w", err)
	}

	return decodedInputs, nil
}

func formatProcessDepositsCalldata(calldata []byte) (*ProcessDepositsInput, error) {
	rollupABI := io.Reader(strings.NewReader(bindings.RollupMetaData.ABI))
	parsedABI, err := abi.JSON(rollupABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	method, err := parsedABI.MethodById(calldata[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse calldata: %w", err)
	}
	decodedInputs, err := decodeProcessDepositsCalldata(method, calldata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode process deposits calldata: %w", err)
	}

	return decodedInputs, nil
}

func decodeRelayMessageCalldata(
	method *abi.Method,
	calldata []byte,
) (*RelayMessageInput, error) {
	// relayMessage(from common.Address, to common.Address, value *big.Int, nonce *big.Int, message []byte)
	if method.Name != relayMessageMethod {
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}

	args, err := method.Inputs.Unpack(calldata[int4Key:])
	if err != nil {
		return nil, errors.Join(ErrUnpackCalldataFail, err)
	}

	decodedInput := RelayMessageInput{
		From:    args[int0Key].(common.Address),
		To:      args[int1Key].(common.Address),
		Value:   args[int2Key].(*big.Int),
		Nonce:   args[int3Key].(*big.Int),
		Message: args[int4Key].([]byte),
	}

	return &decodedInput, nil
}

func decodeProcessDepositsCalldata(
	method *abi.Method,
	calldata []byte,
) (*ProcessDepositsInput, error) {
	// processDeposits(_lastProcessedDepositId *big.Int, depositHashes [][32]byte)
	if method.Name != processDepositsMethod {
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}

	args, err := method.Inputs.Unpack(calldata[int4Key:])
	if err != nil {
		return nil, errors.Join(ErrUnpackCalldataFail, err)
	}

	decodedInput := ProcessDepositsInput{
		LastProcessedDepositId: args[int0Key].(*big.Int),
		DepositHashes:          args[int1Key].([][int32Key]byte),
	}

	return &decodedInput, nil
}

// func (p *blockValidityProver) getRelayedDepositData(
// 	depositId *big.Int,
// ) (*intMaxTree.DepositLeaf, error) {
// 	// TODO: Execute the following three tasks concurrently.
// 	lastProcessedDepositId, err := p.getLastProcessedDepositId()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get last processed depositId: %w", err)
// 	}
// 	if lastProcessedDepositId.Cmp(depositId) == -1 {
// 		return nil, fmt.Errorf("DepositId %v is greater than last processed depositId %v", depositId, lastProcessedDepositId)
// 	}

// 	depositExists, err := p.checkDepositDataExists(depositId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to check deposit data: %w", err)
// 	}
// 	if !depositExists {
// 		return nil, errors.New("this deposit is rejected")
// 	}

// 	isDepositCanceled, err := p.checkIfDepositCanceled(depositId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to check deposit canceled: %w", err)
// 	}
// 	if isDepositCanceled {
// 		return nil, errors.New("this deposit is canceled")
// 	}

// 	deposits, err := p.getDepositData(p.cfg.Blockchain.RollupContractDeployedBlockNumber, []*big.Int{depositId})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get deposit data: %w", err)
// 	}
// 	if len(deposits) != 1 {
// 		return nil, errors.New("no deposit data found")
// 	}
// 	depositData := deposits[0]

// 	depositLeaf := intMaxTree.DepositLeaf{
// 		RecipientSaltHash: depositData.RecipientSaltHash,
// 		TokenIndex:        depositData.TokenIndex,
// 		Amount:            depositData.Amount,
// 	}
// 	fmt.Printf("depositLeaf.RecipientSaltHash: %x\n", depositLeaf.RecipientSaltHash)
// 	fmt.Printf("depositLeaf.TokenIndex: %d\n", depositLeaf.TokenIndex)
// 	fmt.Printf("depositLeaf.Amount: %s\n", depositLeaf.Amount)

// 	return &depositLeaf, nil
// }

// func (p *blockValidityProver) getLastProcessedDepositId() (*big.Int, error) {
// 	result, err := p.rollup.LastProcessedDepositId(&bind.CallOpts{
// 		Pending: false,
// 		Context: p.ctx,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get last processed depositId: %w", err)
// 	}
// 	return result, nil
// }

// func (p *blockValidityProver) checkDepositDataExists(depositId *big.Int) (bool, error) {
// 	result, err := p.liquidity.GetDepositData(&bind.CallOpts{
// 		Pending: false,
// 		Context: p.ctx,
// 	}, depositId)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "execution reverted: out-of-bounds access of an array or bytesN") {
// 			return false, nil
// 		}
// 		return false, fmt.Errorf("failed to get deposit data: %w", err)
// 	}
// 	return !result.IsRejected, nil
// }

// func (p *blockValidityProver) getDepositData(startBlock uint64, depositIds []*big.Int) (_ []*bindings.LiquidityDeposited, _ error) {
// 	nextBlock := startBlock + 1
// 	iterator, err := p.liquidity.FilterDeposited(&bind.FilterOpts{
// 		Start:   nextBlock,
// 		End:     nil,
// 		Context: p.ctx,
// 	}, depositIds, []common.Address{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to filter logs: %w", err)
// 	}

// 	defer iterator.Close()

// 	var events []*bindings.LiquidityDeposited
// 	for iterator.Next() {
// 		event := iterator.Event
// 		events = append(events, event)
// 	}

// 	if err = iterator.Error(); err != nil {
// 		return nil, fmt.Errorf("error encountered while iterating: %w", err)
// 	}

// 	return events, nil
// }

// func (p *blockValidityProver) checkIfDepositCanceled(depositId *big.Int) (bool, error) {
// 	depositIds := []*big.Int{depositId}
// 	iterator, err := p.liquidity.FilterDepositCanceled(&bind.FilterOpts{
// 		Start: 0,
// 		End:   nil,
// 	}, depositIds)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to filter logs: %v", err)
// 	}

// 	defer iterator.Close()

// 	isCanceled := false
// 	for iterator.Next() {
// 		if iterator.Error() != nil {
// 			return false, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
// 		}
// 		isCanceled = true
// 	}

// 	return isCanceled, nil
// }
