package block_post_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"
	"io"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	scrollNetworkRpcUrl   = "https://sepolia-rpc.scroll.io"
	senderPublicKeysIndex = 5
	numOfSenders          = intMaxTypes.NumOfSenders
)

type BlockPostService struct {
	ctx context.Context
	// cfg       *configs.Config
	// log       logger.Logger
	ethClient    *ethclient.Client
	scrollClient *ethclient.Client
	liquidity    *bindings.Liquidity
	rollup       *bindings.Rollup
}

func NewBlockPostService(ctx context.Context, cfg *configs.Config) (*BlockPostService, error) {
	ethClient, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Ethereum client: %w", err)
	}
	defer ethClient.Close()

	scrollClient, err := utils.NewClient(scrollNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Scroll client: %w", err)
	}
	defer scrollClient.Close()

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}
	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.Blockchain.RollupContractAddress), scrollClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Rollup contract: %w", err)
	}

	return &BlockPostService{
		ctx:          ctx,
		ethClient:    ethClient,
		scrollClient: scrollClient,
		liquidity:    liquidity,
		rollup:       rollup,
	}, nil
}

func (d *BlockPostService) FetchNewPostedBlocks(startBlock uint64) ([]*bindings.RollupBlockPosted, *big.Int, error) {
	nextBlock := startBlock + 1
	iterator, err := d.rollup.FilterBlockPosted(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: d.ctx,
	}, [][32]byte{}, []common.Address{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	defer iterator.Close()

	var events []*bindings.RollupBlockPosted
	maxBlockNumber := new(big.Int)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
		if event.BlockNumber.Cmp(maxBlockNumber) > 0 {
			maxBlockNumber.Set(event.BlockNumber)
		}
	}

	if err = iterator.Error(); err != nil {
		return nil, nil, fmt.Errorf("error encountered while iterating: %w", err)
	}

	return events, maxBlockNumber, nil
}

func (d *BlockPostService) FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error) {
	tx, isPending, err := d.scrollClient.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, errors.Join(ErrTransactionByHashNotFound, err)
	}

	if isPending {
		return nil, ErrTransactionIsStillPending
	}

	calldata := tx.Data()

	return calldata, nil
}

// FetchIntMaxBlockContentByHash fetches the block content by transaction hash.
//
// Example:
//
//	ctx := context.Background()
//	cfg := configs.Config{}
//	var startScrollBlockNumber uint64 = 0
//	d, err := block_post_service.NewBlockPostService(ctx, &cfg)
//	events, lastIntMaxBlockNumber, err := d.FetchNewPostedBlocks(startScrollBlockNumber)
//	calldata, err := d.FetchScrollCalldataByHash(events[0].Raw.TxHash)
//	blockContent, err := FetchIntMaxBlockContentByCalldata(calldata)
func FetchIntMaxBlockContentByCalldata(calldata []byte, accountInfoMap AccountInfoMap) (*intMaxTypes.BlockContent, error) {
	// Parse calldata
	rollupABI := io.Reader(strings.NewReader(bindings.RollupABI))
	parsedABI, err := abi.JSON(rollupABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	method, err := parsedABI.MethodById(calldata[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse calldata: %w", err)
	}

	blockContent, err := recoverBlockContent(method, calldata, accountInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to decode calldata: %w", err)
	}

	return blockContent, nil
}

type AccountInfoMap struct {
	LastAccountID uint64
	AccountMap    map[uint64]*intMaxAcc.PublicKey
	PublicKeyMap  map[string]uint64
}

func NewAccountInfoMap() AccountInfoMap {
	return AccountInfoMap{
		AccountMap:    make(map[uint64]*intMaxAcc.PublicKey),
		PublicKeyMap:  make(map[string]uint64),
		LastAccountID: 0,
	}
}

func recoverBlockContent(method *abi.Method, calldata []byte, accountInfoMap AccountInfoMap) (_ *intMaxTypes.BlockContent, err error) {
	switch method.Name {
	case "postRegistrationBlock":
		decodedInput, err := decodePostRegistrationBlockCalldata(method, calldata)
		if err != nil {
			return nil, fmt.Errorf("failed to decode calldata: %w", err)
		}

		blockContent, err := recoverRegistrationBlockContent(decodedInput)
		if err != nil {
			return nil, fmt.Errorf("failed to decode calldata: %w", err)
		}

		dummyAddress := intMaxAcc.NewDummyPublicKey().ToAddress().String()
		for _, sender := range blockContent.Senders {
			if accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] == 0 {
				accountInfoMap.LastAccountID += 1
				accountInfoMap.AccountMap[accountInfoMap.LastAccountID] = sender.PublicKey
				accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] = accountInfoMap.LastAccountID
			} else {
				// TODO: error handling
				address := sender.PublicKey.ToAddress().String()
				if address != dummyAddress {
					fmt.Printf("Account already exists %v\n", address)
				}
			}
		}

		return blockContent, nil
	case "postNonRegistrationBlock":
		decodedInput, err := decodePostNonRegistrationBlockCalldata(method, calldata)
		if err != nil {
			return nil, fmt.Errorf("failed to decode calldata: %w", err)
		}

		blockContent, err := recoverNonRegistrationBlockContent(decodedInput, accountInfoMap)
		if err != nil {
			return nil, fmt.Errorf("failed to decode calldata: %w", err)
		}

		// TODO
		// for _, sender := range blockContent.Senders {
		// 	if accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] == 0 {
		// 		accountInfoMap.LastAccountID += 1
		// 		accountInfoMap.AccountMap[accountInfoMap.LastAccountID] = sender.PublicKey
		// 		accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] = accountInfoMap.LastAccountID
		// 	} else {
		// 		// TODO: error handling
		// 		if
		// 		fmt.Println("Account already exists")
		// 	}
		// }

		return blockContent, nil
	default:
		return nil, fmt.Errorf("invalid method name: %s", method.Name)
	}
}

func decodePostRegistrationBlockCalldata(method *abi.Method, calldata []byte) (*intMaxTypes.PostRegistrationBlockInput, error) {
	if method.Name != "postRegistrationBlock" {
		return nil, fmt.Errorf("invalid method name: %s", method.Name)
	}

	args, err := method.Inputs.Unpack(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack calldata: %w", err)
	}

	decodedInput := intMaxTypes.PostRegistrationBlockInput{
		TxTreeRoot:          args[0].([32]byte),
		SenderFlags:         args[1].([16]byte),
		AggregatedPublicKey: args[2].([2][32]byte),
		AggregatedSignature: args[3].([4][32]byte),
		MessagePoint:        args[4].([4][32]byte),
		SenderPublicKeys:    args[5].([]*big.Int),
	}

	return &decodedInput, nil
}

func decodePostNonRegistrationBlockCalldata(method *abi.Method, calldata []byte) (*intMaxTypes.PostNonRegistrationBlockInput, error) {
	if method.Name != "postNonRegistrationBlock" {
		return nil, fmt.Errorf("invalid method name: %s", method.Name)
	}

	args, err := method.Inputs.Unpack(calldata[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack calldata: %w", err)
	}

	decodedInput := intMaxTypes.PostNonRegistrationBlockInput{
		TxTreeRoot:          args[0].([32]byte),
		SenderFlags:         args[1].([16]byte),
		AggregatedPublicKey: args[2].([2][32]byte),
		AggregatedSignature: args[3].([4][32]byte),
		MessagePoint:        args[4].([4][32]byte),
		PublicKeysHash:      args[5].([32]byte),
		SenderAccountIds:    args[6].([]byte),
	}

	return &decodedInput, nil
}

func recoverRegistrationBlockContent(decodedInput *intMaxTypes.PostRegistrationBlockInput) (_ *intMaxTypes.BlockContent, err error) {
	senderPublicKeys := make([]*intMaxAcc.PublicKey, numOfSenders)
	for i, key := range decodedInput.SenderPublicKeys {
		senderPublicKeys[i], err = intMaxAcc.NewPublicKeyFromAddressInt(key)
		if err != nil {
			return nil, errors.Join(ErrCannotDecodeAddress, err)
		}
	}
	for i := len(decodedInput.SenderPublicKeys); i < numOfSenders; i++ {
		senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
	}

	const (
		int3Key  = 3
		int8Key  = 8
		int32Key = 32
	)

	senderFlags := make([]bool, numOfSenders)
	for i := 0; i < numOfSenders; i++ {
		byteIndex := i / int8Key
		bitIndex := i % int8Key
		senderFlags[i] = (decodedInput.SenderFlags[byteIndex] & (1 << bitIndex)) != 0
	}

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		if senderFlags[i] {
			senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for i, isSigned := range senderFlags {
		if isSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, senderPublicKeys[i].WeightByHash(publicKeysHash))
		}
	}

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, sender := range senderPublicKeys {
		senders[i] = intMaxTypes.Sender{
			PublicKey: sender,
			AccountID: 0,
			IsSigned:  senderFlags[i],
		}
	}

	txTreeRoot := new(intMaxTypes.PoseidonHashOut)
	if err = txTreeRoot.Unmarshal(decodedInput.TxTreeRoot[:]); err != nil {
		return nil, fmt.Errorf("failed to set tx tree root: %w", err)
	}

	// Recover aggregatedSignature from decodedInput.AggregatedSignature
	aggregatedSignature := new(bn254.G2Affine)
	aggregatedSignature.X.A1.SetBytes(decodedInput.AggregatedSignature[0][:])
	aggregatedSignature.X.A0.SetBytes(decodedInput.AggregatedSignature[1][:])
	aggregatedSignature.Y.A1.SetBytes(decodedInput.AggregatedSignature[2][:])
	aggregatedSignature.Y.A0.SetBytes(decodedInput.AggregatedSignature[int3Key][:])

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		*txTreeRoot,
		aggregatedSignature,
	)

	return blockContent, nil
}

func recoverNonRegistrationBlockContent(
	decodedInput *intMaxTypes.PostNonRegistrationBlockInput,
	accountInfoMap AccountInfoMap,
) (_ *intMaxTypes.BlockContent, err error) {
	senderAccountIds, err := intMaxTypes.UnmarshalAccountIds(decodedInput.SenderAccountIds)
	if err != nil {
		return nil, fmt.Errorf("failed to recover account IDs from bytes: %w", err)
	}

	senderPublicKeys := make([]*intMaxAcc.PublicKey, numOfSenders)
	for i, accountId := range senderAccountIds {
		if accountId == 0 {
			senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
			continue
		}

		key, ok := accountInfoMap.AccountMap[accountId]
		if !ok {
			return nil, errors.Join(ErrUnknownAccountID, fmt.Errorf("%d", accountId))
		}

		senderPublicKeys[i] = key
	}
	for i := len(senderAccountIds); i < numOfSenders; i++ {
		senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
	}

	const (
		int3Key  = 3
		int8Key  = 8
		int32Key = 32
	)

	senderFlags := make([]bool, numOfSenders)
	for i := 0; i < numOfSenders; i++ {
		byteIndex := i / int8Key
		bitIndex := i % int8Key
		senderFlags[i] = (decodedInput.SenderFlags[byteIndex] & (1 << bitIndex)) != 0
	}

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		if senderFlags[i] {
			senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for i, isSigned := range senderFlags {
		if isSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, senderPublicKeys[i].WeightByHash(publicKeysHash))
		}
	}

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, sender := range senderPublicKeys {
		senders[i] = intMaxTypes.Sender{
			PublicKey: sender,
			AccountID: 0,
			IsSigned:  senderFlags[i],
		}
	}

	txTreeRoot := new(intMaxTypes.PoseidonHashOut)
	if err = txTreeRoot.Unmarshal(decodedInput.TxTreeRoot[:]); err != nil {
		return nil, fmt.Errorf("failed to set tx tree root: %w", err)
	}

	// Recover aggregatedSignature from decodedInput.AggregatedSignature
	aggregatedSignature := new(bn254.G2Affine)
	aggregatedSignature.X.A1.SetBytes(decodedInput.AggregatedSignature[0][:])
	aggregatedSignature.X.A0.SetBytes(decodedInput.AggregatedSignature[1][:])
	aggregatedSignature.Y.A1.SetBytes(decodedInput.AggregatedSignature[2][:])
	aggregatedSignature.Y.A0.SetBytes(decodedInput.AggregatedSignature[int3Key][:])

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		*txTreeRoot,
		aggregatedSignature,
	)

	return blockContent, nil
}

// func (d *BlockPostService) FetchNewDeposits(startBlock uint64) ([]*bindings.LiquidityDeposited, *big.Int, map[uint32]bool, error) {
// 	nextBlock := startBlock + 1
// 	iterator, err := d.liquidity.FilterDeposited(&bind.FilterOpts{
// 		Start:   nextBlock,
// 		End:     nil,
// 		Context: d.ctx,
// 	}, []*big.Int{}, []common.Address{})
// 	if err != nil {
// 		return nil, nil, nil, fmt.Errorf("failed to filter logs: %w", err)
// 	}

// 	defer iterator.Close()

// 	var events []*bindings.LiquidityDeposited
// 	maxDepositIndex := new(big.Int)
// 	tokenIndexMap := make(map[uint32]bool)

// 	for iterator.Next() {
// 		event := iterator.Event
// 		events = append(events, event)
// 		tokenIndexMap[event.TokenIndex] = true
// 		if event.DepositId.Cmp(maxDepositIndex) > 0 {
// 			maxDepositIndex.Set(event.DepositId)
// 		}
// 	}

// 	if err = iterator.Error(); err != nil {
// 		return nil, nil, nil, fmt.Errorf("error encountered while iterating: %w", err)
// 	}

// 	return events, maxDepositIndex, tokenIndexMap, nil
// }
